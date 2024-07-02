package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// AddEntity is sent by the server to the client. Its function is not entirely clear: It does not add an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type AddEntity struct {
	// EntityNetworkID is the network ID of the entity that should be added.
	// Netease
	EntityNetworkID uint32
}

// ID ...
func (pk *AddEntity) ID() uint32 {
	return IDAddEntity
}

func (pk *AddEntity) Marshal(io protocol.IO) {
	io.Varuint32(&pk.EntityNetworkID) // For Netease
}
