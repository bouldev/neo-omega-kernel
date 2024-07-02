package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// Netease packet
type SetDimensionLocalTime struct {
	// Netease
	Unknown1 int32
	// Netease
	Unknown2 bool
	// Netease: Unknown1 and Unknown2 only, but we read the whole content to this field to avoid panic
	Unknown3 []byte
}

// ID ...
func (*SetDimensionLocalTime) ID() uint32 {
	return IDSetDimensionLocalTime
}

func (pk *SetDimensionLocalTime) Marshal(io protocol.IO) {
	// if len > 0 {
	// 	io.Varint32(&pk.Unknown1)
	// 	io.Bool(&pk.Unknown2)
	// }
	io.Bytes(&pk.Unknown3)
}
