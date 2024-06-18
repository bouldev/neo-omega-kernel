package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease packet
type HungerAttr struct {
	// Netease: uncertain type, read all
	Data []byte
}

// ID ...
func (*HungerAttr) ID() uint32 {
	return IDHungerAttr
}

func (pk *HungerAttr) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
