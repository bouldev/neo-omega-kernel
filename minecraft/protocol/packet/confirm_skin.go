package packet

import (
	"neo-omega-kernel/minecraft/protocol"
)

// Netease packet
type ConfirmSkin struct {
	// Netease
	Unknown1 []protocol.ConfirmSkinUnknownEntry
	// Netease
	Unknown2 []string
	// Netease
	Unknown3 []string
}

// ID ...
func (*ConfirmSkin) ID() uint32 {
	return IDConfirmSkin
}

func (pk *ConfirmSkin) Marshal(io protocol.IO) {
	protocol.SliceVaruint32Length(io, &pk.Unknown1)
	protocol.FuncSliceOfLen(io, uint32(len(pk.Unknown1)), &pk.Unknown2, io.String)
	protocol.FuncSliceOfLen(io, uint32(len(pk.Unknown1)), &pk.Unknown3, io.String)
}
