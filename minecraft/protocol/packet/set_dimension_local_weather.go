package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// Netease packet
type SetDimensionLocalWeather struct {
	// Netease
	Unknown1 bool
	// Netease
	Unknown2 float32
	// Netease
	Unknown3 int32
	// Netease
	Unknown4 float32
	// Netease
	Unknown5 int32
	// Netease
	Unknown6 bool
}

// ID ...
func (*SetDimensionLocalWeather) ID() uint32 {
	return IDSetDimensionLocalWeather
}

func (pk *SetDimensionLocalWeather) Marshal(io protocol.IO) {
	io.Bool(&pk.Unknown1)
	io.Float32(&pk.Unknown2)
	io.Varint32(&pk.Unknown3)
	io.Float32(&pk.Unknown4)
	io.Varint32(&pk.Unknown5)
	io.Bool(&pk.Unknown6)
}
