package bot_action

import (
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"

	"github.com/go-gl/mathgl/mgl32"
)

type BotActionSimple struct {
	uq   neomega.MicroUQHolder
	ctrl neomega.InteractCore
}

func NewBotActionSimple(uq neomega.MicroUQHolder, ctrl neomega.InteractCore) *BotActionSimple {
	return &BotActionSimple{
		uq:   uq,
		ctrl: ctrl,
	}
}

func (b *BotActionSimple) SetStructureBlockData(pos define.CubePos, settings *neomega.StructureBlockSettings) {
	updatePacket := &packet.StructureBlockUpdate{
		Position:           protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		StructureName:      settings.StructureName,
		DataField:          settings.DataField,
		IncludePlayers:     settings.IncludePlayers != 0,
		ShowBoundingBox:    settings.ShowBoundingBox != 0,
		StructureBlockType: settings.StructureBlockType,
		Settings: protocol.StructureSettings{
			PaletteName:               "default",
			IgnoreEntities:            settings.IgnoreEntities != 0,
			IgnoreBlocks:              settings.IgnoreBlocks != 0,
			Size:                      protocol.BlockPos{settings.XStructureSize, settings.YStructureSize, settings.ZStructureSize},
			Offset:                    protocol.BlockPos{settings.XStructureOffset, settings.YStructureOffset, settings.ZStructureOffset},
			LastEditingPlayerUniqueID: b.uq.GetBotUniqueID(),
			Rotation:                  settings.Rotation,
			Mirror:                    settings.Mirror,
			AnimationMode:             settings.AnimationMode,
			AnimationDuration:         settings.AnimationDuration,
			Integrity:                 settings.Integrity,
			Seed:                      uint32(settings.Seed),
			Pivot:                     mgl32.Vec3{0, 0, 0},
		},
		RedstoneSaveMode: settings.RedstoneSaveMode,
		ShouldTrigger:    true,
	}
	b.ctrl.SendPacket(updatePacket)
}

func (s *BotActionSimple) SleepTick(ticks int) {
	time.Sleep(time.Millisecond * 50 * time.Duration(ticks))
}
