package core

import (
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/minecraft_conn"
	"github.com/OmineDev/neomega-core/nodes/defines"
)

type AccessPointInteractCore struct {
	minecraft_conn.Conn
}

func (i *AccessPointInteractCore) SendPacket(packet packet.Packet) {
	i.WritePacket(packet)
}

func (i *AccessPointInteractCore) SendPacketBytes(packet []byte) {
	i.WriteBytePacket(packet)
}

func NewAccessPointInteractCore(node defines.APINode, conn minecraft_conn.Conn) neomega.InteractCore {
	core := &AccessPointInteractCore{Conn: conn}
	node.ExposeAPI("send-packet-bytes", func(args defines.Values) (result defines.Values, err error) {
		packetBytes, err := args.ToBytes()
		if err != nil {
			return defines.Empty, err
		}
		conn.WriteBytePacket(packetBytes)
		return defines.Empty, nil
	}, false)
	node.ExposeAPI("get-shield-id", func(args defines.Values) (result defines.Values, err error) {
		shieldID := conn.GetShieldID()
		return defines.FromInt32(shieldID), nil
	}, false)
	return core
}

func NewAccessPointReactCore(node defines.Node, conn minecraft_conn.Conn) neomega.UnStartedReactCore {
	core := NewReactCore()
	go func() {
		nodeDead := <-node.WaitClosed()
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
			pkt, _ = conn.ReadPacketAndBytes()
			core.handlePacket(pkt)
		}
		// prob := block_prob.NewBlockProb("Access Point MC Packet Handle Block Prob", time.Second/10)
		for {
			pkt, packetData = conn.ReadPacketAndBytes()
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
			node.PublishMessage("packets", defines.FromInt32(conn.GetShieldID()).ExtendFrags(packetData))
			// node.PublishMessage("packet", nodes.FromInt32(conn.GetShieldID()).ExtendFrags(packetData))
			// prob.MarkEventFinished(mark)
		}
		core.DeadReason <- fmt.Errorf("%v: %v", ErrRentalServerDisconnected, i18n.FuzzyTransErr(err))
	}
	return core
}
