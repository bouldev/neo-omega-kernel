package core

import (
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/minecraft_conn"
	"neo-omega-kernel/nodes"
	"time"
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

func NewAccessPointInteractCore(node nodes.APINode, conn minecraft_conn.Conn) neomega.InteractCore {
	core := &AccessPointInteractCore{Conn: conn}
	node.ExposeAPI("send-packet-bytes", func(args nodes.Values) (result nodes.Values, err error) {
		packetDataBytes, err := args.ToBytes()
		if err != nil {
			return nodes.Empty, err
		}
		pks := bytesToBytesSlices(packetDataBytes)
		for _, pk := range pks {
			conn.WriteBytePacket(pk)
		}
		return nodes.Empty, nil
	}, false)
	node.ExposeAPI("get-shield-id", func(args nodes.Values) (result nodes.Values, err error) {
		shieldID := conn.GetShieldID()
		return nodes.FromInt32(shieldID), nil
	}, false)
	return core
}

func NewAccessPointReactCore(node nodes.Node, conn minecraft_conn.Conn) neomega.UnStartedReactCore {
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
			pkt, _ = conn.ReadPacketAndBytes()
			core.handlePacket(pkt)
		}
		// prob := block_prob.NewBlockProb("Access Point MC Packet Handle Block Prob", time.Second/10)
		ticker := time.NewTicker(time.Second / 20)
		batchedBytes := [][]byte{}
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
			batchedBytes = append(batchedBytes, packetData)
			select {
			case <-ticker.C:
				catBytes := byteSlicesToBytes(batchedBytes)
				batchedBytes = [][]byte{}
				node.PublishMessage("packets", nodes.FromInt32(conn.GetShieldID()).ExtendFrags(catBytes))
			default:
			}
			// node.PublishMessage("packet", nodes.FromInt32(conn.GetShieldID()).ExtendFrags(packetData))
			// prob.MarkEventFinished(mark)
		}
		core.DeadReason <- fmt.Errorf("%v: %v", ErrRentalServerDisconnected, i18n.FuzzyTransErr(err))
	}
	return core
}
