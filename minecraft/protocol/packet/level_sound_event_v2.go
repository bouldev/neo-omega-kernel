package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type LevelSoundEventV2 struct {
	// Netease
	Unknown1 uint8
	// Netease
	Posistion mgl32.Vec3
	// Netease
	Unknown2 int32
	// Netease
	Unknown3 string
	// Netease
	Unknown4 bool
	// Netease
	Unknown5 bool
}

// ID ...
func (*LevelSoundEventV2) ID() uint32 {
	return IDLevelSoundEventV2
}

func (pk *LevelSoundEventV2) Marshal(io protocol.IO) {
	io.Uint8(&pk.Unknown1)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.Unknown2)
	io.String(&pk.Unknown3)
	io.Bool(&pk.Unknown4)
	io.Bool(&pk.Unknown5)
}
