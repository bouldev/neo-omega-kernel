package core

import (
	"bytes"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/nodes"
	"sync"
	"time"
)

type EndPointInteractCore struct {
	nodes.APINode
	shieldID             int32
	sendMu               sync.Mutex
	queuedSendingPackets [][]byte
}

func (i *EndPointInteractCore) SendPacketBytes(packet []byte) {
	i.sendMu.Lock()
	defer i.sendMu.Unlock()
	i.queuedSendingPackets = append(i.queuedSendingPackets, packet)
}

func (i *EndPointInteractCore) SendPacket(pk packet.Packet) {
	writer := bytes.NewBuffer(nil)
	hdr := &packet.Header{}
	hdr.PacketID = pk.ID()
	hdr.Write(writer)
	w := protocol.NewWriter(writer, i.shieldID)
	pk.Marshal(w)
	i.sendMu.Lock()
	defer i.sendMu.Unlock()
	i.queuedSendingPackets = append(i.queuedSendingPackets, writer.Bytes())
}

type canNotifyShieldIDChange interface {
	ListenShieldIDUpdate(func(int32))
}

func NewEndPointInteractCore(node nodes.Node, shieldIDProvider canNotifyShieldIDChange) (neomega.InteractCore, error) {
	result, err := node.CallWithResponse("get-shield-id", nodes.Empty).SetTimeout(time.Second * 3).BlockGetResponse()
	if err != nil {
		return nil, err
	}
	currentShieldID, err := result.ToInt32()
	if err != nil {
		return nil, err
	}
	core := &EndPointInteractCore{node, currentShieldID, sync.Mutex{}, [][]byte{}}
	go func() {
		ticker := time.NewTicker(time.Second / 20)
		for {
			select {
			case <-ticker.C:
				if len(core.queuedSendingPackets) > 0 {
					var catBytes []byte
					core.sendMu.Lock()
					catBytes = byteSlicesToBytes(core.queuedSendingPackets)
					core.queuedSendingPackets = [][]byte{}
					core.sendMu.Unlock()
					node.CallOmitResponse("send-packet-bytes", nodes.Empty.ExtendFrags(catBytes))
				}
			case <-node.Dead():
				return
			}

		}
	}()
	shieldIDProvider.ListenShieldIDUpdate(func(newShieldID int32) {
		core.shieldID = newShieldID
	})
	return core, nil
}

type ShieldIDUpdateNotifier struct {
	first           bool
	currentShieldID int32
	listeners       []func(int32)
}

func (n *ShieldIDUpdateNotifier) ListenShieldIDUpdate(listener func(int32)) {
	n.listeners = append(n.listeners, listener)
}

func (n *ShieldIDUpdateNotifier) updateShieldID(shieldID int32) {
	update := false
	if n.first {
		update = true
		n.first = false
	} else if shieldID != n.currentShieldID {
		update = true
	}
	if update {
		n.currentShieldID = shieldID
		for _, l := range n.listeners {
			l(shieldID)
		}
	}
}

func safeDecode(pkt packet.Packet, r *protocol.Reader) (p packet.Packet, err error) {
	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			err = fmt.Errorf("%T: %w", pkt, recoveredErr.(error))
		}
	}()
	pkt.Marshal(r)
	return pkt, nil
}

func NewEndPointReactCore(node nodes.Node) interface {
	neomega.UnStartedReactCore
	canNotifyShieldIDChange
} {
	core := NewReactCore()
	go func() {
		nodeDead := <-node.Dead()
		err := fmt.Errorf("node dead: %v", nodeDead)
		core.DeadReason <- err
	}()
	shieldIDUpdateNotifier := &ShieldIDUpdateNotifier{
		first:           true,
		currentShieldID: 0,
		listeners:       make([]func(int32), 0, 1),
	}
	// prob := block_prob.NewBlockProb("End Point MC Packet Handle Block Prob", time.Second/10)
	node.ListenMessage("packets", func(msg nodes.Values) {
		shieldID, err := msg.ToInt32()
		if err != nil {
			err := fmt.Errorf("end point get shield id dead: %v", err)
			core.DeadReason <- err
			return
		}
		shieldIDUpdateNotifier.updateShieldID(shieldID)
		msg = msg.ConsumeHead()
		catBytes, err := msg.ToBytes()
		if err != nil {
			err := fmt.Errorf("end point get msg dead: %v", err)
			core.DeadReason <- err
			return
		}
		batchedBytes := bytesToBytesSlices(catBytes)
		{
			for _, packetData := range batchedBytes {
				reader := bytes.NewBuffer(packetData)
				header := &packet.Header{}
				if err := header.Read(reader); err != nil {
					core.DeadReason <- fmt.Errorf("end point error reading packet header: %v", err)
					return
				}
				r := protocol.NewReader(reader, shieldID, false)
				if pktMake, found := pool[header.PacketID]; found {
					pk := pktMake()
					pk, err = safeDecode(pk, r)
					if err != nil {
						// fmt.Println(err)
					} else {
						// mark := prob.MarkEventStartByTimeout(func() string {
						// 	bs, _ := json.Marshal(pk)
						// 	return fmt.Sprint(pk.ID()) + string(bs)
						// }, time.Second/5)
						core.handlePacket(pk)
						// prob.MarkEventFinished(mark)
					}
				} else {
					// fmt.Printf("pktID %v not found\n", header.PacketID)
				}
			}
		}
	}, false)
	return struct {
		neomega.UnStartedReactCore
		canNotifyShieldIDChange
	}{
		core,
		shieldIDUpdateNotifier,
	}
}
