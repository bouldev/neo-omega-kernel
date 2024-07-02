package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease Packet
type AnimateEmoteEntity struct {
	// Netease
	Unknown1 string
	// Netease
	Unknown2 string
	// Netease
	Unknown3 string
	// Netease
	Unknown4 int32
	// Netease
	Unknown5 string
	// Netease
	Unknown6 float32
	// Netease: uncertain, varint32 + slice = []sometype
	Unknown7 []byte
}

// ID ...
func (*AnimateEmoteEntity) ID() uint32 {
	return IDAnimateEmoteEntity
}

func (pk *AnimateEmoteEntity) Marshal(io protocol.IO) {
	io.String(&pk.Unknown1)
	io.String(&pk.Unknown2)
	io.String(&pk.Unknown3)
	io.Int32(&pk.Unknown4)
	io.String(&pk.Unknown5)
	io.Float32(&pk.Unknown6)
	//io.ByteSlice(&pk.Unknown7)
	io.Bytes(&pk.Unknown7)
}
