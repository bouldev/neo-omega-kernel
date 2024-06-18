package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type Transport struct {
	// Netease
	Unknown1 int32
	// Netease: uncertain type, read all
	Unknown2 []byte
}

// ID ...
func (*Transport) ID() uint32 {
	return IDTransport
}

func (pk *Transport) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
