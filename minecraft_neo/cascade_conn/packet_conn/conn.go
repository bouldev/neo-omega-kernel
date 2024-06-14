package packet_conn

import (
	"bytes"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/cascade_conn/can_close"
	"neo-omega-kernel/minecraft_neo/cascade_conn/defines"
)

type ByteFrameConnOmitClose interface {
	ReadRoutine(func([]byte))
	WriteBytePacket([]byte)
	defines.AdvancedConnControl
}

type PacketConn struct {
	can_close.CanCloseWithError
	ByteFrameConnOmitClose
	pool map[uint32]func() packet.Packet
}

func (c *PacketConn) ReadRoutine(botLogic defines.BotReactLogic) {
	c.ByteFrameConnOmitClose.ReadRoutine(func(rawPacket []byte) {
		pk, err := c.decodePacket(rawPacket, botLogic.ShieldID())
		if err != nil {
			c.CloseWithError(err)
			return
		}
		if pk.ID() == packet.IDDisconnect {
			pkt := pk.(*packet.Disconnect)
			c.CloseWithError(fmt.Errorf(pkt.Message))
			return
		}
		pkAndByte := &defines.RawPacketAndByte{
			Raw: rawPacket,
			Pk:  pk,
		}
		botLogic.HandlePacket(pkAndByte)
	})
}

func (c *PacketConn) decodePacket(rawPacket []byte, shieldID int32) (pk packet.Packet, err error) {
	data, err := parseData(rawPacket)
	if err != nil {
		return nil, err
	}
	pk = c.pool[data.h.PacketID]()
	r := protocol.NewReader(data.payload, shieldID, false)
	pk.Marshal(r)
	if data.payload.Len() != 0 {
		fmt.Println(pk.ID())
		err = fmt.Errorf("%T: %v unread bytes left: 0x%x", pk, data.payload.Len(), data.payload.Bytes())
	}
	return
}

func (c *PacketConn) WritePacket(pk packet.Packet, shieldID int32) error {
	if c.CanCloseWithError.Closed() {
		return fmt.Errorf("write packet on closed conn")
	}
	buf := bytes.NewBuffer([]byte{})
	header := &packet.Header{}
	header.PacketID = pk.ID()
	header.Write(buf)
	pk.Marshal(protocol.NewWriter(buf, shieldID))
	c.WriteBytePacket(buf.Bytes())
	return nil
}

// func (c *PacketConn) Flush() error {
// 	return c.underlayConn.Flush()
// }

// func (c *PacketConn)EnableEncryption([32]byte){

// }
// func (c *PacketConn)EnableCompression(packet.Compression)
// func (c *PacketConn)Flush() error
// func (c *PacketConn)Lock()
// func (c *PacketConn)UnLock()

// func (c *PacketConn) PostPoneHandleUntilOneOf(pktIDs []uint32) {
// 	c.expectedIDs.Store(pktIDs)
// }

// func (c *PacketConn) handleCouldBeLoginAndSpawnPacket(pk *RawPacketAndByte) {
// 	for _, id := range c.expectedIDs.Load().([]uint32) {
// 		if id == pk.pk.ID() {
// 			c.handleLoginAndSpawnPacket(pk.pk)
// 		}
// 	}
// }

// func (c *PacketConn) handleLoginAndSpawnPacket(pk packet.Packet) {

// }

func NewClientFromConn(conn defines.ByteFrameConn) defines.PacketConn {
	client := &PacketConn{
		CanCloseWithError:      can_close.NewClose(conn.Close),
		pool:                   packet.NewPool(),
		ByteFrameConnOmitClose: conn,
	}
	go func() {
		err := <-conn.WaitClosed()
		client.CloseWithError(err)
	}()
	return client
}
