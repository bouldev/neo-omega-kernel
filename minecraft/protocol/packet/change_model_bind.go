package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type ChangeModelBind struct {
	// Netease
	Unknown1 int64
	// Netease
	Unknown2 int64
}

// ID ...
func (*ChangeModelBind) ID() uint32 {
	return IDChangeModelBind
}

func (pk *ChangeModelBind) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Varint64(&pk.Unknown2)
}
