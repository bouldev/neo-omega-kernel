package packet

import (
	"neo-omega-kernel/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type LevelSoundEventV1 struct {
	// Netease
	Unknown1 uint8
	// Netease
	Posistion mgl32.Vec3
	// Netease
	Unknown2 int32
	// Netease
	Unknown3 int32
	// Netease
	Unknown4 bool
	// Netease
	Unknown5 bool
}

// ID ...
func (*LevelSoundEventV1) ID() uint32 {
	return IDLevelSoundEventV1
}

func (pk *LevelSoundEventV1) Marshal(io protocol.IO) {
	io.Uint8(&pk.Unknown1)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.Unknown2)
	io.Varint32(&pk.Unknown3)
	io.Bool(&pk.Unknown4)
	io.Bool(&pk.Unknown5)
}
