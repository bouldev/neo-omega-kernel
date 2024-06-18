package packet

import "neo-omega-kernel/minecraft/protocol"

// Netease Packet
type MobBlockActorChanged struct {
	// Netease
	Position protocol.BlockPos
}

// ID ...
func (*MobBlockActorChanged) ID() uint32 {
	return IDMobBlockActorChanged
}

func (pk *MobBlockActorChanged) Marshal(io protocol.IO) {
	io.BlockPos(&pk.Position)
}
