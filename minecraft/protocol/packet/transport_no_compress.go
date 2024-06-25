package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type TransportNoCompress struct {
	// Netease
	Unknown1 int32
	// Netease: uncertain type, read all
	Unknown2 []byte
}

// ID ...
func (*TransportNoCompress) ID() uint32 {
	return IDTransportNoCompress
}

func (pk *TransportNoCompress) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}