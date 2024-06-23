package packet

import (
	"neo-omega-kernel/minecraft/protocol"
)

// Netease packet
type ConfirmSkin struct {
	// Netease
	Count uint32
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
	io.Varuint32(&pk.Count)
	if pk.Count > 0 {
		protocol.SliceOfLen(io, pk.Count, &pk.Unknown1)
		protocol.FuncSliceOfLen(io, pk.Count, &pk.Unknown2, io.String)
		protocol.FuncSliceOfLen(io, pk.Count, &pk.Unknown3, io.String)
	}
}
