package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease packet
type ChangeModelTexture struct {
	// Netease
	Unknown1 int64
	// Netease
	Unknown2 string
	// Netease
	Unknown3 int64
	// Netease
	Unknown4 uint8
}

// ID ...
func (*ChangeModelTexture) ID() uint32 {
	return IDChangeModelTexture
}

func (pk *ChangeModelTexture) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.String(&pk.Unknown2)
	io.Varint64(&pk.Unknown3)
	io.Uint8(&pk.Unknown4)
}
