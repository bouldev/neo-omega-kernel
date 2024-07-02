package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// RemoveVolumeEntity indicates a volume entity to be removed from server to client.
type RemoveVolumeEntity struct {
	// EntityRuntimeID ...
	// Netease
	EntityRuntimeID uint32
	// Dimension ...
	Dimension int32
}

// ID ...
func (*RemoveVolumeEntity) ID() uint32 {
	return IDRemoveVolumeEntity
}

func (pk *RemoveVolumeEntity) Marshal(io protocol.IO) {
	io.Varuint32(&pk.EntityRuntimeID) // For Netease
	io.Varint32(&pk.Dimension)
}
