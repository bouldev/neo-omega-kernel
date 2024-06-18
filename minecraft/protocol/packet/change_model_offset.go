package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type ChangeModelOffset struct {
	// Netease
	Unknown1 int64
	// Netease: uncertain, 2 same operations
	Unknown2 []byte
}

// ID ...
func (*ChangeModelOffset) ID() uint32 {
	return IDChangeModelOffset
}

func (pk *ChangeModelOffset) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
