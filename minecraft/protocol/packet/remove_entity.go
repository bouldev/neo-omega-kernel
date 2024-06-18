package packet

import (
	"neo-omega-kernel/minecraft/protocol"
)

// RemoveEntity is sent by the server to the client. Its function is not entirely clear: It does not remove an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type RemoveEntity struct {
	// EntityNetworkID is the network ID of the entity that should be removed.
	// Netease
	EntityNetworkID uint32
}

// ID ...
func (pk *RemoveEntity) ID() uint32 {
	return IDRemoveEntity
}

func (pk *RemoveEntity) Marshal(io protocol.IO) {
	io.Varuint32(&pk.EntityNetworkID) // For Netease
}
