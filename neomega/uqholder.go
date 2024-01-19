package neomega

import (
	"neo-omega-kernel/minecraft/protocol/packet"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

type UQInfoHolderEntry interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	UpdateFromPacket(packet packet.Packet)
}

type BotBasicInfoHolder interface {
	GetBotName() string
	GetBotRuntimeID() uint64
	GetBotUniqueID() int64
	GetBotIdentity() string
	GetBotUUIDStr() string
	UQInfoHolderEntry
}

type PlayerUQReader interface {
	GetUUID() (id uuid.UUID, found bool)
	GetUUIDString() (id string, found bool)
	GetEntityUniqueID() (id int64, found bool)
	GetLoginTime() (time time.Time, found bool)
	GetUsername() (name string, found bool)
	GetPlatformChatID() (id string, found bool)
	GetBuildPlatform() (platform int32, found bool)
	GetSkinID() (id string, found bool)
	GetPropertiesFlag() (flag uint32, found bool)
	GetCommandPermissionLevel() (level uint32, found bool)
	GetActionPermissions() (permissions uint32, found bool)
	GetAbilityString() (adventureFlagsMap, actionPermissionMap map[string]bool, found bool)
	GetOPPermissionLevel() (level uint32, found bool)
	GetCustomStoredPermissions() (permissions uint32, found bool)
	GetDeviceID() (id string, found bool)
	GetEntityRuntimeID() (id uint64, found bool)
	GetEntityMetadata() (entityMetadata map[uint32]any, found bool)
	IsOP() (op bool, found bool)
	StillOnline() bool
}

type PlayersInfoHolder interface {
	GetAllOnlinePlayers() []PlayerUQReader
	GetPlayerByUUID(uuid.UUID) (player PlayerUQReader, found bool)
	GetPlayerByUUIDString(uuidStr string) (player PlayerUQReader, found bool)
	GetPlayerByUniqueID(uniqueID int64) (player PlayerUQReader, found bool)
	GetPlayerByName(name string) (player PlayerUQReader, found bool)
	UQInfoHolderEntry
}

type GameRule struct {
	CanBeModifiedByPlayer bool
	Value                 string
}

type ExtendInfo interface {
	GetCompressThreshold() (compressThreshold uint16, found bool)
	GetWorldGameMode() (worldGameMode int32, found bool)
	GetWorldDifficulty() (worldDifficulty uint32, found bool)
	GetTime() (time int32, found bool)
	GetDayTime() (dayTime int32, found bool)
	GetDayTimePercent() (dayTimePercent float32, found bool)
	GetGameRules() (gameRules map[string]*GameRule, found bool)
	GetCurrentTick() (currentTick int64, found bool)
	GetSyncRatio() (ratio float32, known bool)
	GetCurrentOpenedContainer() (container *packet.ContainerOpen, open bool)
	GetBotDimension() (dimension int32, found bool)
	GetBotPosition() (pos mgl32.Vec3, outOfSyncTick int64)
	UQInfoHolderEntry
}

// type NetWorkData interface {
// 	NetAsyncSetData(key string, value []byte, cb func(error))
// 	NetBlockSetData(key string, value []byte) error
// 	NetAsyncGetData(key string, cb func(data []byte, found bool))
// 	NetBlockGetData(key string) (value []byte, found bool)
// }

type MicroUQHolder interface {
	GetBotBasicInfo() BotBasicInfoHolder
	GetPlayersInfo() PlayersInfoHolder
	GetExtendInfo() ExtendInfo
	BotBasicInfoHolder
	PlayersInfoHolder
	ExtendInfo
	UQInfoHolderEntry
}

// type PlayerUQsHolder interface {
// 	GetPlayerUQByName(name string) (uq PlayerUQReader, found bool)
// 	GetPlayerUQByUUID(ud uuid.UUID) (uq PlayerUQReader, found bool)
// 	GetBot() (botUQ PlayerUQReader)
// }

// type PlayerUQReader interface {
// 	IsBot() bool
// 	GetPlayerName() string
// }

// type PlayerUQ interface {
// 	PlayerUQReader
// }
