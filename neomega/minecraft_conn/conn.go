package minecraft_conn

import (
	"neo-omega-kernel/minecraft/protocol/login"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/game_data"
)

type Conn interface {
	GameData() game_data.GameData
	IdentityData() login.IdentityData
	GetShieldID() int32
	ReadPacketAndBytes() (packet.Packet, []byte)
	WritePacket(packet.Packet)
	WriteBytePacket([]byte)
}
