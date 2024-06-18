package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type CustomV1 struct {
	// Netease: uncertain type, read all
	Data []byte
}

// ID ...
func (*CustomV1) ID() uint32 {
	return IDCustomV1
}

func (pk *CustomV1) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
