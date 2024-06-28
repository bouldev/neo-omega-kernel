package uqholder

import (
	"neo-omega-kernel/neomega"
	"time"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(neomega.PlayerUQReader) {}(&Player{})
	}
}

type Player struct {
	UUID                uuid.UUID
	knownUUID           bool
	EntityUniqueID      int64
	knownEntityUniqueID bool
	NeteaseUID          int64
	knownNeteaseUID     bool
	LoginTime           time.Time
	knownLoginTime      bool
	Username            string
	knownUsername       bool
	PlatformChatID      string
	knownPlatformChatID bool
	BuildPlatform       int32
	knownBuildPlatform  bool
	SkinID              string
	knownSkinID         bool
	// PropertiesFlag               uint32
	// knownPropertiesFlag          bool
	// CommandPermissionLevel       uint32
	// knownCommandPermissionLevel  bool
	// ActionPermissions            uint32
	// knownActionPermissions       bool
	// OPPermissionLevel            uint32
	// knownOPPermissionLevel       bool
	// CustomStoredPermissions      uint32
	// knownCustomStoredPermissions bool
	knowAbilitiesAndStatus bool
	canBuild               bool
	canMine                bool
	canDoorsAndSwitches    bool
	canOpenContainers      bool
	canAttackPlayers       bool
	canAttackMobs          bool
	canOperatorCommands    bool
	canTeleport            bool
	statusInvulnerable     bool
	statusFlying           bool
	statusMayFly           bool
	// StatusInstantBuild      bool
	// StatusLightning         bool
	// StatusMuted             bool
	// StatusWorldBuilder      bool
	// StatusNoClip            bool
	// StatusPrivilegedBuilder bool
	// StatusCount             bool

	DeviceID             string
	knownDeviceID        bool
	EntityRuntimeID      uint64
	knownEntityRuntimeID bool
	EntityMetadata       map[uint32]any
	knownEntityMetadata  bool
	Online               bool
}

func NewPlayerUQHolder() *Player {
	return &Player{
		Online: true,
	}
}

func (p *Player) StillOnline() bool {
	return p.Online
}

func (p *Player) GetUUID() (id uuid.UUID, found bool) {
	if p == nil {
		return
	}
	return p.UUID, p.knownUUID
}

func (p *Player) GetUUIDString() (id string, found bool) {
	if p == nil {
		return
	}
	return p.UUID.String(), p.knownUUID
}

func (p *Player) setUUID(id uuid.UUID) {
	p.UUID = id
	p.knownUUID = true
}

func (p *Player) GetEntityUniqueID() (id int64, found bool) {
	if p == nil {
		return
	}
	return p.EntityUniqueID, p.knownEntityUniqueID
}

func (p *Player) setEntityUniqueID(id int64) {
	p.EntityUniqueID = id
	p.knownEntityUniqueID = true
}

func (p *Player) GetNeteaseUID() (id int64, found bool) {
	if p == nil {
		return
	}
	return p.NeteaseUID, p.knownNeteaseUID
}

func (p *Player) setNeteaseUID(id int64) {
	p.NeteaseUID = id
	p.knownNeteaseUID = true
}

func (p *Player) GetLoginTime() (t time.Time, found bool) {
	if p == nil {
		return
	}
	return p.LoginTime, p.knownLoginTime
}

func (p *Player) setLoginTime(t time.Time) {
	p.LoginTime = t
	p.knownLoginTime = true
}

func (p *Player) GetUsername() (name string, found bool) {
	if p == nil {
		return
	}
	return p.Username, p.knownUsername
}

func (p *Player) setUsername(name string) {
	p.Username = name
	p.knownUsername = true
}

func (p *Player) GetPlatformChatID() (id string, found bool) {
	if p == nil {
		return
	}
	return p.PlatformChatID, p.knownPlatformChatID
}

func (p *Player) setPlatformChatID(id string) {
	p.PlatformChatID = id
	p.knownPlatformChatID = true
}

func (p *Player) GetBuildPlatform() (platform int32, found bool) {
	if p == nil {
		return
	}
	return p.BuildPlatform, p.knownBuildPlatform
}

func (p *Player) setBuildPlatform(platform int32) {
	p.BuildPlatform = platform
	p.knownBuildPlatform = true
}

func (p *Player) GetSkinID() (id string, found bool) {
	if p == nil {
		return
	}
	return p.SkinID, p.knownSkinID
}

func (p *Player) setSkinID(id string) {
	p.SkinID = id
	p.knownSkinID = true
}

// func (p *Player) GetPropertiesFlag() (flag uint32, found bool) {
// 	if p == nil {
// 		return
// 	}
// 	return p.PropertiesFlag, p.knownPropertiesFlag
// }

// func (p *Player) setPropertiesFlag(flag uint32) {
// 	p.PropertiesFlag = flag
// 	p.knownPropertiesFlag = true
// }

// func (p *Player) GetCommandPermissionLevel() (level uint32, found bool) {
// 	if p == nil {
// 		return
// 	}
// 	return p.CommandPermissionLevel, p.knownCommandPermissionLevel
// }

// func (p *Player) setCommandPermissionLevel(level uint32) {
// 	p.CommandPermissionLevel = level
// 	p.knownCommandPermissionLevel = true
// }

// func (p *Player) GetActionPermissions() (permissions uint32, found bool) {
// 	if p == nil {
// 		return
// 	}
// 	return p.ActionPermissions, p.knownActionPermissions
// }

// func (p *Player) setActionPermissions(permissions uint32) {
// 	p.ActionPermissions = permissions
// 	p.knownActionPermissions = true
// }

// func (p *Player) GetOPPermissionLevel() (level uint32, found bool) {
// 	if p == nil {
// 		return
// 	}
// 	return p.OPPermissionLevel, p.knownOPPermissionLevel
// }

// func (p *Player) IsOP() (op bool, found bool) {
// 	if p == nil {
// 		return false, false
// 	}
// 	if p.knownOPPermissionLevel {
// 		if p.OPPermissionLevel == packet.PermissionLevelOperator {
// 			return true, true
// 		}
// 		if p.OPPermissionLevel < packet.PermissionLevelOperator {
// 			return false, true
// 		}
// 	}
// 	permission, found := p.GetCommandPermissionLevel()
// 	if found {
// 		if permission >= packet.CommandPermissionLevelHost {
// 			return true, true
// 		} else {
// 			return false, true
// 		}
// 	}

// 	permission, found = p.GetActionPermissions()
// 	isOP := (permission & packet.ActionPermissionOperator) != 0
// 	return isOP, found
// }

// func (p *Player) setOPPermissionLevel(level uint32) {
// 	p.OPPermissionLevel = level
// 	p.knownOPPermissionLevel = true
// }

// func (p *Player) GetCustomStoredPermissions() (permissions uint32, found bool) {
// 	if p == nil {
// 		return
// 	}
// 	return p.CustomStoredPermissions, p.knownCustomStoredPermissions
// }

// func (p *Player) setCustomStoredPermissions(permissions uint32) {
// 	p.CustomStoredPermissions = permissions
// 	p.knownCustomStoredPermissions = true
// }

func (p *Player) IsOP() (op bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canOperatorCommands, true

	}
	return false, false
}

func (p *Player) GetDeviceID() (id string, found bool) {
	return p.DeviceID, p.knownDeviceID
}

func (p *Player) setDeviceID(id string) {
	p.DeviceID = id
	p.knownDeviceID = true
}

func (p *Player) GetEntityRuntimeID() (id uint64, found bool) {
	return p.EntityRuntimeID, p.knownEntityRuntimeID
}

func (p *Player) setEntityRuntimeID(id uint64) {
	p.EntityRuntimeID = id
	p.knownEntityRuntimeID = true
}

func (p *Player) GetEntityMetadata() (entityMetadata map[uint32]any, found bool) {
	return p.EntityMetadata, p.knownEntityMetadata
}

func (p *Player) setEntityMetadata(entityMetadata map[uint32]any) {
	p.EntityMetadata = entityMetadata
	p.knownEntityMetadata = true
}

func (p *Player) CanBuild() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canBuild, true

	}
	return false, false
}

func (p *Player) CanMine() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canMine, true

	}
	return false, false
}

func (p *Player) CanDoorsAndSwitches() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.canDoorsAndSwitches {
		return p.canBuild, true

	}
	return false, false
}

func (p *Player) CanOpenContainers() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canOpenContainers, true

	}
	return false, false
}

func (p *Player) CanAttackPlayers() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canAttackPlayers, true

	}
	return false, false
}

func (p *Player) CanAttackMobs() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canAttackMobs, true

	}
	return false, false
}

func (p *Player) CanOperatorCommands() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canOperatorCommands, true

	}
	return false, false
}

func (p *Player) CanTeleport() (hasAbility bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.canTeleport, true

	}
	return false, false
}

func (p *Player) StatusInvulnerable() (hasStatus bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.statusInvulnerable, true

	}
	return false, false
}

func (p *Player) StatusFlying() (hasStatus bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.statusFlying, true

	}
	return false, false
}

func (p *Player) StatusMayFly() (hasStatus bool, found bool) {
	if p == nil {
		return false, false
	}
	if p.knowAbilitiesAndStatus {
		return p.statusMayFly, true

	}
	return false, false
}

// var AdventureFlagMap = map[string]uint32{
// 	"AdventureFlagWorldImmutable":        packet.AdventureFlagWorldImmutable,
// 	"AdventureSettingsFlagsNoPvM":        packet.AdventureSettingsFlagsNoPvM,
// 	"AdventureSettingsFlagsNoMvP":        packet.AdventureSettingsFlagsNoMvP,
// 	"AdventureSettingsFlagsUnused":       packet.AdventureSettingsFlagsUnused,
// 	"AdventureSettingsFlagsShowNameTags": packet.AdventureSettingsFlagsShowNameTags,
// 	"AdventureFlagAutoJump":              packet.AdventureFlagAutoJump,
// 	"AdventureFlagAllowFlight":           packet.AdventureFlagAllowFlight,
// 	"AdventureFlagNoClip":                packet.AdventureFlagNoClip,
// 	"AdventureFlagWorldBuilder":          packet.AdventureFlagWorldBuilder,
// 	"AdventureFlagFlying":                packet.AdventureFlagFlying,
// 	"AdventureFlagMuted":                 packet.AdventureFlagMuted,
// }

// var ActionPermissionMap = map[string]uint32{
// 	"ActionPermissionMine":             packet.ActionPermissionMine,
// 	"ActionPermissionDoorsAndSwitches": packet.ActionPermissionDoorsAndSwitches,
// 	"ActionPermissionOpenContainers":   packet.ActionPermissionOpenContainers,
// 	"ActionPermissionAttackPlayers":    packet.ActionPermissionAttackPlayers,
// 	"ActionPermissionAttackMobs":       packet.ActionPermissionAttackMobs,
// 	"ActionPermissionOperator":         packet.ActionPermissionOperator,
// 	"ActionPermissionTeleport":         packet.ActionPermissionTeleport,
// 	"ActionPermissionBuild":            packet.ActionPermissionBuild,
// 	"ActionPermissionDefault":          packet.ActionPermissionDefault,
// }

// func (p *Player) GetAbilityString() (adventureFlagsMap, actionPermissionMap map[string]bool, found bool) {
// 	adventureFlagsMap = make(map[string]bool)
// 	actionPermissionMap = make(map[string]bool)
// 	adventrueFlags, ok := p.GetPropertiesFlag()
// 	if !ok {
// 		return
// 	}
// 	actionFlags, ok := p.GetActionPermissions()
// 	if !ok {
// 		return
// 	}

// 	for flagName, flagValue := range AdventureFlagMap {
// 		adventureFlagsMap[flagName] = (adventrueFlags & flagValue) != 0
// 	}
// 	for flagName, flagValue := range ActionPermissionMap {
// 		actionPermissionMap[flagName] = (actionFlags & flagValue) != 0
// 	}

// 	return adventureFlagsMap, actionPermissionMap, true
// }
