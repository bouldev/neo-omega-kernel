package packet

import (
	"neo-omega-kernel/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type MobEffectV2 struct {
	// Netease
	Items []struct {
		// Netease
		Unknown1 int32
		// Netease
		Posistion mgl32.Vec3
		// Netease
		Unknown2 string
		// Netease
		Unknown3 string
	}
}

// ID ...
func (*MobEffectV2) ID() uint32 {
	return IDMobEffectV2
}

func (pk *MobEffectV2) Marshal(io protocol.IO) {
	var count uint32
	io.Varuint32(&count)

	if count > 0 {
		pk.Items = make([]struct {
			Unknown1  int32
			Posistion mgl32.Vec3
			Unknown2  string
			Unknown3  string
		}, count)

		for i := uint32(0); i < count; i++ {
			io.Varint32(&pk.Items[i].Unknown1)
			io.Vec3(&pk.Items[i].Posistion)
			io.String(&pk.Items[i].Unknown2)
			io.String(&pk.Items[i].Unknown3)
		}
	}
}
