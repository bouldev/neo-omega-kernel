package core

import (
	"errors"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/utils/sync_wrapper"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(neomega.ReactCore) {}(&ReactCore{})
	}
}

var ErrRentalServerDisconnected = errors.New(i18n.T(i18n.S_rental_server_disconnected))

func init() {
	i18n.DoAndAddRetranslateHook(func() {
		ErrRentalServerDisconnected = errors.New(i18n.T(i18n.S_rental_server_disconnected))
	})
}

type ReactCore struct {
	onAnyPacketCallBack           []func(packet.Packet)
	onTypedPacketCallBacks        [][]func(packet.Packet)
	noBlockAndDetachableCallbacks []*sync_wrapper.SyncKVMap[string, neomega.NoBlockAndDetachablePacketCallback]
	DeadReason                    chan error
	deferredStart                 func()
	// oneTimeTypedPacketCallBacks *sync_wrapper.SyncKVMap[string,*sync_wrapper.SyncKVMap[string,func(packet.Packet) bool]]
	// slowPacketChan              chan packet.Packet
}

func NewReactCore() *ReactCore {
	core := &ReactCore{
		onAnyPacketCallBack:           make([]func(packet.Packet), 0),
		onTypedPacketCallBacks:        make([][]func(packet.Packet), 0, 300),
		noBlockAndDetachableCallbacks: make([]*sync_wrapper.SyncKVMap[string, neomega.NoBlockAndDetachablePacketCallback], 0, 300),
		DeadReason:                    make(chan error, 16),
		// oneTimeTypedPacketCallBacks: sync_wrapper.NewSyncKVMap[string,*sync_wrapper.SyncKVMap[string,func(packet.Packet) bool]](),
		// slowPacketChan:              make(chan packet.Packet, 1024),
	}
	maxID := 0
	for _, id := range mcPacketNameIDMapping {
		if id > uint32(maxID) {
			maxID = int(id)
		}
	}
	if maxID < 300 {
		// we already know netease mc packet could have id 228 and maybe even greater\
		// these packet event not registered in code so we make room for this
		maxID = 300
	}
	for i := 0; i < maxID; i++ {
		core.noBlockAndDetachableCallbacks = append(core.noBlockAndDetachableCallbacks, sync_wrapper.NewSyncKVMap[string, neomega.NoBlockAndDetachablePacketCallback]())
		core.onTypedPacketCallBacks = append(core.onTypedPacketCallBacks, make([]func(packet.Packet), 0, 16))
	}
	return core
}

func (r *ReactCore) Dead() chan error {
	return r.DeadReason
}

func (r *ReactCore) Start() {
	if r.deferredStart != nil {
		go r.deferredStart()
		r.deferredStart = nil
	}
}

func (r *ReactCore) SetAnyPacketCallBack(callback func(packet.Packet), newGoroutine bool) {
	if newGoroutine {
		r.onAnyPacketCallBack = append(r.onAnyPacketCallBack, func(p packet.Packet) {
			go callback(p)
		})
	} else {
		r.onAnyPacketCallBack = append(r.onAnyPacketCallBack, callback)
	}

}

func (r *ReactCore) SetTypedPacketCallBack(packetID uint32, callback func(packet.Packet), newGoroutine bool) {
	listeners := r.onTypedPacketCallBacks[packetID]
	if newGoroutine {
		r.onTypedPacketCallBacks[packetID] = append(listeners, func(p packet.Packet) {
			go callback(p)
		})
	} else {
		r.onTypedPacketCallBacks[packetID] = append(listeners, callback)
	}

}

func (r *ReactCore) AddNewNoBlockAndDetachablePacketCallBack(wants map[uint32]bool, cb neomega.NoBlockAndDetachablePacketCallback) {
	for id := range wants {
		// 将数据处理函数添加到与协议包 ID 对应的集合中
		subPumpers := r.noBlockAndDetachableCallbacks[id]
		subPumpers.Set(uuid.New().String(), cb)
	}
}

func (r *ReactCore) handlePacket(pkt packet.Packet) {
	if pkt == nil {
		return
	}
	pktID := pkt.ID()
	// if pktID == packet.IDCommandOutput {
	// 	s, _ := json.Marshal(pkt)
	// 	fmt.Println(string(s))
	// }
	for _, cb := range r.onAnyPacketCallBack {
		cb(pkt)
	}
	if pktID > uint32(len(r.onTypedPacketCallBacks)) {
		panic(fmt.Errorf("recv packet id %v", pktID))
	}
	cbs := r.onTypedPacketCallBacks[pktID]
	for _, cb := range cbs {
		cb(pkt)
	}
	// r.slowPacketChan <- pkt
	go r.handleNoBlockAndDetachablePacket(pkt)
}

func (r *ReactCore) handleNoBlockAndDetachablePacket(pk packet.Packet) {
	id := pk.ID()
	// 获取与协议包 ID 对应的数据处理函数集合
	subPumper := r.noBlockAndDetachableCallbacks[id]
	toRemove := []string{}
	// 遍历数据处理函数集合，依次调用其中的函数进行处理
	subPumper.Iter(func(k string, pumper neomega.NoBlockAndDetachablePacketCallback) (continueIter bool) {
		err := pumper(pk)
		// 如果处理函数返回错误(停止读取)，则将其从集合中删除
		if err != nil {
			toRemove = append(toRemove, k)
		}
		return true
	})
	// 将需要删除的处理函数从集合中删除
	for _, k := range toRemove {
		subPumper.Delete(k)
	}
}
