package uqholder

import (
	"fmt"
	"neo-omega-kernel/minecraft"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"runtime/debug"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	if false {
		func(neomega.ExtendInfo) {}(&ExtendInfoHolder{})
	}
}

// 包含窗口ID与请求变更的新物品信息
type ItemStackRequestDetails struct {
	WindowID             uint32
	SlotWithItemInstance map[uint8]protocol.ItemInstance
}

type ExtendInfoHolder struct {
	// WorldName              string
	// knownWorldName         bool
	CompressThreshold            uint16
	knownCompressThreshold       bool
	lastSyncRatioStaticStartTime time.Time
	lastSyncRatioStaticStartTick int64
	syncRatio                    float32
	CurrentTick                  int64
	knownCurrentTick             bool
	WorldGameMode                int32
	knownWorldGameMode           bool
	WorldDifficulty              uint32
	knownWorldDifficulty         bool
	currentContainerOpened       bool
	currentOpenedContainer       *packet.ContainerOpen
	// InventorySlotCount      uint32
	// knownInventorySlotCount bool
	Time                int32
	knownTime           bool
	DayTime             int32
	knownDayTime        bool
	DayTimePercent      float32
	knownDayTimePercent bool
	GameRules           map[string]*neomega.GameRule
	knownGameRules      bool
	Dimension           int32
	knownDimension      bool
	// dup of runtime id in bot basic info
	botRuntimeIDDup    uint64
	PositionUpdateTick int64
	Position           mgl32.Vec3
}

func NewExtendInfoHolder(conn *minecraft.Conn) *ExtendInfoHolder {
	return &ExtendInfoHolder{
		GameRules:          make(map[string]*neomega.GameRule),
		botRuntimeIDDup:    conn.GameData().EntityRuntimeID,
		Position:           conn.GameData().PlayerPosition,
		Dimension:          conn.GameData().Dimension,
		PositionUpdateTick: conn.GameData().Time,
		CurrentTick:        conn.GameData().Time,
	}
}

// func (e *ExtendInfoHolder) GetWorldName() (worldName string, found bool) {
// 	return e.WorldName, e.knownWorldName
// }

// func (e *ExtendInfoHolder) setWorldName(worldName string) {
// 	e.WorldName = worldName
// 	e.knownWorldName = true
// }

func (e *ExtendInfoHolder) GetCompressThreshold() (compressThreshold uint16, found bool) {
	return e.CompressThreshold, e.knownCompressThreshold
}

func (e *ExtendInfoHolder) setCompressThreshold(compressThreshold uint16) {
	e.CompressThreshold = compressThreshold
	e.knownCompressThreshold = true
}

func (e *ExtendInfoHolder) GetCurrentTick() (currentTick int64, found bool) {
	return e.CurrentTick, e.knownCurrentTick
}

func (e *ExtendInfoHolder) setCurrentTick(currentTick int64) {
	e.CurrentTick = currentTick
	e.knownCurrentTick = true
}

func (e *ExtendInfoHolder) GetWorldGameMode() (worldGameMode int32, found bool) {
	return e.WorldGameMode, e.knownWorldGameMode
}

func (e *ExtendInfoHolder) setWorldGameMode(worldGameMode int32) {
	e.WorldGameMode = worldGameMode
	e.knownWorldGameMode = true
}

func (e *ExtendInfoHolder) GetWorldDifficulty() (worldDifficulty uint32, found bool) {
	return e.WorldDifficulty, e.knownWorldDifficulty
}

func (e *ExtendInfoHolder) setWorldDifficulty(worldDifficulty uint32) {
	e.WorldDifficulty = worldDifficulty
	e.knownWorldDifficulty = true
}

// func (e *ExtendInfoHolder) GetInventorySlotCount() (inventorySlotCount uint32, found bool) {
// 	return e.InventorySlotCount, e.knownInventorySlotCount
// }

// func (e *ExtendInfoHolder) setInventorySlotCount(inventorySlotCount uint32) {
// 	e.InventorySlotCount = inventorySlotCount
// 	e.knownInventorySlotCount = true
// }

func (e *ExtendInfoHolder) GetTime() (time int32, found bool) {
	return e.Time, e.knownTime
}

func (e *ExtendInfoHolder) setTime(time int32) {
	e.Time = time
	e.knownTime = true
}

func (e *ExtendInfoHolder) GetDayTime() (dayTime int32, found bool) {
	return e.DayTime, e.knownDayTime
}

func (e *ExtendInfoHolder) setDayTime(dayTime int32) {
	e.DayTime = dayTime
	e.knownDayTime = true
}

func (e *ExtendInfoHolder) GetDayTimePercent() (dayTimePercent float32, found bool) {
	return e.DayTimePercent, e.knownDayTimePercent
}

func (e *ExtendInfoHolder) setDayTimePercent(dayTimePercent float32) {
	e.DayTimePercent = dayTimePercent
	e.knownDayTimePercent = true
}

func (e *ExtendInfoHolder) GetGameRules() (gameRules map[string]*neomega.GameRule, found bool) {
	return e.GameRules, e.knownGameRules
}

func (e *ExtendInfoHolder) GetSyncRatio() (ratio float32, known bool) {
	return e.syncRatio, e.syncRatio == 0
}

func (e *ExtendInfoHolder) setGameRules(gameRuleName string, rule *neomega.GameRule) {
	e.GameRules[gameRuleName] = rule
	e.knownGameRules = true
}

func (e *ExtendInfoHolder) GetCurrentOpenedContainer() (container *packet.ContainerOpen, open bool) {
	return e.currentOpenedContainer, e.currentContainerOpened
}

func (e *ExtendInfoHolder) GetBotDimension() (dimension int32, found bool) {
	if e.knownDimension {
		return e.Dimension, true
	} else {
		return 0, false
	}
}

func (e *ExtendInfoHolder) GetBotPosition() (pos mgl32.Vec3, outOfSyncTick int64) {
	// though currently position is always known,
	// in future we may use it (found) to represent "out of sync" status
	// fmt.Printf("e.CurrentTick %v e.PositionUpdateTick %v\n", e.CurrentTick, e.PositionUpdateTick)
	outOfSyncTick = e.CurrentTick - e.PositionUpdateTick
	if outOfSyncTick < 0 {
		outOfSyncTick = 0
	}
	return e.Position, outOfSyncTick
}

func (uq *ExtendInfoHolder) UpdateFromPacket(pk packet.Packet) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("UQHolder Update Error: ", r)
			debug.PrintStack()
		}
	}()
	switch p := pk.(type) {
	case *packet.NetworkSettings:
		uq.setCompressThreshold(p.CompressionThreshold)
	case *packet.SetTime:
		uq.setTime(p.Time)
		uq.setDayTime(p.Time % 24000)
		uq.setDayTimePercent(float32(uq.DayTime) / 24000)
	case *packet.GameRulesChanged:
		for _, r := range p.GameRules {
			uq.setGameRules(r.Name, &neomega.GameRule{
				CanBeModifiedByPlayer: r.CanBeModifiedByPlayer,
				Value:                 fmt.Sprintf("%v", r.Value),
			})
		}
	case *packet.SetDefaultGameType:
		uq.setWorldGameMode(p.GameType)
	case *packet.SetDifficulty:
		uq.setWorldDifficulty(p.Difficulty)
	case *packet.TickSync:
		nowTime := time.Now()
		if p.ClientRequestTimestamp == 0 {
			uq.setCurrentTick(p.ServerReceptionTimestamp)
			// fmt.Println("tick sync", p)
		} else {
			deltaTime := p.ServerReceptionTimestamp - p.ClientRequestTimestamp
			if deltaTime < 0 {
				deltaTime = 0
			}
			uq.setCurrentTick(p.ServerReceptionTimestamp + deltaTime)
			if uq.lastSyncRatioStaticStartTick != 0 {
				ticksShouldGo := nowTime.Sub(uq.lastSyncRatioStaticStartTime).Milliseconds() / 50
				ticksActualGo := p.ServerReceptionTimestamp - uq.lastSyncRatioStaticStartTick
				syncRatio := float32(ticksActualGo) / float32(ticksShouldGo)
				if syncRatio > 1 {
					uq.syncRatio = 1
				} else {
					uq.syncRatio = syncRatio
				}
			}
			uq.lastSyncRatioStaticStartTick = p.ServerReceptionTimestamp
			uq.lastSyncRatioStaticStartTime = time.Now()
		}
	case *packet.ChangeDimension:
		uq.Dimension = p.Dimension
		uq.knownDimension = true
		uq.Position = p.Position
		uq.PositionUpdateTick = uq.CurrentTick
	case *packet.MovePlayer:
		// fmt.Println(p)
		if p.EntityRuntimeID == uq.botRuntimeIDDup {
			// fmt.Println(p)
			uq.Position = p.Position
			// uq.CurrentTick = int64(p.Tick) + 1 p.Tick is 0
			uq.PositionUpdateTick = uq.CurrentTick
			// EntityRuntimeID:          1,
			// Position:                 p.Position,
			// Pitch:                    p.Pitch,
			// Yaw:                      p.Yaw,
			// HeadYaw:                  p.HeadYaw,
			// Mode:                     p.Mode,
			// OnGround:                 p.OnGround,
			// RiddenEntityRuntimeID:    p.RiddenEntityRuntimeID,
			// TeleportCause:            p.TeleportCause,
			// TeleportSourceEntityType: p.TeleportSourceEntityType,
			// Tick:                     o.Tick + 1,
		}

	case *packet.Respawn:
		if p.EntityRuntimeID == uq.botRuntimeIDDup {
			uq.Position = p.Position
			uq.PositionUpdateTick = uq.CurrentTick
		}
	case *packet.CorrectPlayerMovePrediction:
		uq.Position = p.Position
		uq.CurrentTick = int64(p.Tick) + 1
		uq.PositionUpdateTick = uq.CurrentTick
	case *packet.ContainerOpen:
		uq.currentOpenedContainer = p
		uq.currentContainerOpened = true
	case *packet.ContainerClose:
		uq.currentOpenedContainer = nil
		uq.currentContainerOpened = false
	}
}
