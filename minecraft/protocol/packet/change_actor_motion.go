package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease packet
type ChangeActorMotion struct {
	// Netease
	Unknown1 int64
	// Netease
	Unknown2 uint8
}

// ID ...
func (*ChangeActorMotion) ID() uint32 {
	return IDChangeActorMotion
}

func (pk *ChangeActorMotion) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Uint8(&pk.Unknown2)
}
