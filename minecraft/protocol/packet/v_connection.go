package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease packet
type VConnection struct {
	// Netease
	Unknown1 int32
	// Netease: uncertain type, read all
	Unknown2 []byte
}

// ID ...
func (*VConnection) ID() uint32 {
	return IDVConnection
}

func (pk *VConnection) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
