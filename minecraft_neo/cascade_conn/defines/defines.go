package defines

import (
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/cascade_conn/can_close"
)

type AdvancedConnControl interface {
	EnableEncryption([32]byte)
	EnableCompression(packet.Compression)
	Flush() error
	Lock()
	UnLock()
}

type ByteFrameConnBase interface {
	can_close.CanClose
	ReadRoutine(func([]byte))
	WriteBytePacket([]byte)
}

type ByteFrameConn interface {
	ByteFrameConnBase
	AdvancedConnControl
}

type RawPacketAndByte struct {
	Raw []byte
	Pk  packet.Packet
}

type BotReactLogic interface {
	HandlePacket(*RawPacketAndByte)
	ShieldID() int32
}

type PacketConnBase interface {
	can_close.CanClose
	ReadRoutine(botLogic BotReactLogic)
	WritePacket(pk packet.Packet, shieldID int32) error
}

type PacketConn interface {
	PacketConnBase
	AdvancedConnControl
}
