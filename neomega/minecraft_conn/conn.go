package minecraft_conn

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol/login"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/minecraft_neo/game_data"
)

type Conn interface {
	GameData() game_data.GameData
	IdentityData() login.IdentityData
	GetShieldID() int32
	ReadPacketAndBytes() (packet.Packet, []byte)
	WritePacket(packet.Packet)
	WriteBytePacket([]byte)
}
