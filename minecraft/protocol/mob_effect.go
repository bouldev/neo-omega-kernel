package protocol

import "github.com/go-gl/mathgl/mgl32"

// Netease
type MobEffectV2UnknownEntry struct {
	// Netease
	Unknown1 int32
	// Netease
	Posistion mgl32.Vec3
	// Netease
	Unknown2 string
	// Netease
	Unknown3 string
}

// Netease
func (m *MobEffectV2UnknownEntry) Marshal(io IO) {
	io.Varint32(&m.Unknown1)
	io.Vec3(&m.Posistion)
	io.String(&m.Unknown2)
	io.String(&m.Unknown3)
}
