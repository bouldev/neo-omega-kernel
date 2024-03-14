package core

import (
	"bytes"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/nodes"
	"time"
)

type AccessPointInteractCore struct {
	*minecraft.Conn
}

func (i *AccessPointInteractCore) SendPacket(packet packet.Packet) {
	i.WritePacket(packet)
}

func (i *AccessPointInteractCore) SendPacketBytes(packetID uint32, packet []byte) {
	i.WriteRawPacket(packetID, packet)
}

func NewAccessPointInteractCore(node nodes.APINode, conn *minecraft.Conn) neomega.InteractCore {
	core := &AccessPointInteractCore{Conn: conn}
	node.ExposeAPI("send-packet-bytes", func(args nodes.Values) (result nodes.Values, err error) {
		pktIDBytes, err := args.ToUint32()
		if err != nil {
			return nodes.Empty, err
		}
		args = args.ConsumeHead()
		packetDataBytes, err := args.ToBytes()
		if err != nil {
			return nodes.Empty, err
		}
		core.SendPacketBytes(pktIDBytes, packetDataBytes)
		return nodes.Empty, nil
	}, false)
	node.ExposeAPI("get-shield-id", func(args nodes.Values) (result nodes.Values, err error) {
		shieldID := conn.GetShieldID()
		return nodes.FromInt32(shieldID), nil
	}, false)
	return core
}

func NewAccessPointReactCore(node nodes.Node, conn *minecraft.Conn) neomega.UnStartedReactCore {
	core := NewReactCore()
	go func() {
		nodeDead := <-node.Dead()
		err := fmt.Errorf("node dead: %v", nodeDead)
		core.DeadReason <- err
	}()
	botRuntimeID := conn.GameData().EntityRuntimeID
	// go core.handleSlowPacketChan()
	//counter := 0
	core.deferredStart = func() {
		var pkt packet.Packet
		var err error
		var packetData []byte
		// packets before conn.ReadPacketAndBytes will be queued until conn.ReadPacketAndBytes is called,
		// so at the very beginning there will be a packet burst
		initPacketBurstEnd := time.Now().Add(time.Second * 1)
		for time.Now().Before(initPacketBurstEnd) {
			pkt, _, err = conn.ReadPacketAndBytes()
			if err != nil {
				break
			}
			core.handlePacket(pkt)
		}
		// prob := block_prob.NewBlockProb("Access Point MC Packet Handle Block Prob", time.Second/10)
		for {
			pkt, packetData, err = conn.ReadPacketAndBytes()
			if err != nil {
				break
			}
			// counter++
			// fmt.Printf("recv packet %v\n", counter)
			// fmt.Println(pkt.ID(), pkt)
			// mark := prob.MarkEventStartByTimeout(func() string {
			// 	bs, _ := json.Marshal(pkt)
			// 	return fmt.Sprint(pkt.ID()) + string(bs)
			// }, time.Second/5)

			core.handlePacket(pkt)
			if pkt.ID() == packet.IDMovePlayer {
				pk := pkt.(*packet.MovePlayer)
				if pk.EntityRuntimeID != botRuntimeID {
					// prob.MarkEventFinished(mark)
					continue
				}
			} else if pkt.ID() == packet.IDMoveActorDelta {
				pk := pkt.(*packet.MoveActorDelta)
				if pk.EntityRuntimeID != botRuntimeID {
					// prob.MarkEventFinished(mark)
					continue
				}
			} else if pkt.ID() == packet.IDSetActorData {
				pk := pkt.(*packet.SetActorData)
				if pk.EntityRuntimeID != botRuntimeID {
					// prob.MarkEventFinished(mark)
					continue
				}
			} else if pkt.ID() == packet.IDSetActorMotion {
				pk := pkt.(*packet.SetActorMotion)
				if pk.EntityRuntimeID != botRuntimeID {
					// prob.MarkEventFinished(mark)
					continue
				}
			}
			// gopher tunnel 可能会重用此 []byte 因此我们需要将其拷贝一份... 当然这是有代价的
			// 也许未来可以优化? 然而事实是我们并不知道此数据何时被利用完成
			packetDataCloned := bytes.Clone(packetData)
			node.PublishMessage("packet", nodes.FromInt32(conn.GetShieldID()).ExtendFrags(packetDataCloned))
			// prob.MarkEventFinished(mark)
		}
		core.DeadReason <- fmt.Errorf("%v: %v", ErrRentalServerDisconnected, i18n.FuzzyTransErr(err))
	}
	return core
}
