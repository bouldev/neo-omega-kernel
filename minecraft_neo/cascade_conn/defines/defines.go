package defines

import (
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/can_close"
)

type AdvancedConnControl interface {
	EnableEncryption([32]byte)
	EnableCompression(packet.Compression)
	Flush() error
}

type ByteFrameConnBase interface {
	can_close.CanClose
	ReadRoutine(func([]byte))
	WriteBytePacket([]byte)
}

type ByteFrameConn interface {
	ByteFrameConnBase
	AdvancedConnControl
	Lock()
	UnLock()
}

type RawPacketAndByte struct {
	Raw []byte
	Pk  packet.Packet
}

// type BotReactLogic interface {
// 	HandlePacket(*RawPacketAndByte)
// 	ShieldID() int32
// }

type PacketConnBase interface {
	can_close.CanClose
	ListenRoutine(func(pk packet.Packet, raw []byte))
	WritePacket(packet.Packet) error
	SetShieldID(int32)
	GetShieldID() int32
}

type PacketConn interface {
	PacketConnBase
	AdvancedConnControl
}
