package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type CustomV2 struct {
	// Netease: uncertain type, read all
	Data []byte
}

// ID ...
func (*CustomV2) ID() uint32 {
	return IDCustomV2
}

func (pk *CustomV2) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
