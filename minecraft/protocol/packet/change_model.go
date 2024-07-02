package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease packet
type ChangeModel struct {
	// Netease
	Unknown1 int64
	// Netease
	Unknown2 string
}

// ID ...
func (*ChangeModel) ID() uint32 {
	return IDChangeModel
}

func (pk *ChangeModel) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.String(&pk.Unknown2)
}
