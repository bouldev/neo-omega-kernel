package player_interact

import (
	"encoding/json"
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/uqholder"
	"strings"
)

type PlayerKit struct {
	neomega.PlayerUQReader
	userName string
	i        *PlayerInteract
}

// 	Say(msg string)
// RawSay(msg string)
// ActionBar(msg string)
// Title(msg string)
// SubTitle(msg string)
// InterceptJustNextInput(cb func(chat *GameChat))

func (p *PlayerKit) Say(msg string) {
	p.i.info.SayTo(p.userName, msg)
}

func (p *PlayerKit) RawSay(msg string) {
	p.i.info.RawSayTo(p.userName, msg)
}

func (p *PlayerKit) ActionBar(msg string) {
	p.i.info.ActionBarTo(p.userName, msg)
}

func (p *PlayerKit) Title(msg string) {
	p.i.info.TitleTo(p.userName, msg)
}

func (p *PlayerKit) SubTitle(subTitle string, title string) {
	p.i.info.SubTitleTo(p.userName, subTitle, title)
}

func (p *PlayerKit) InterceptJustNextInput(cb func(chat *neomega.GameChat)) {
	p.i.InterceptJustNextInput(p.userName, cb)
}

func (p *PlayerKit) CheckCondition(onResult func(bool), conditions ...string) {
	condstr := strings.Join(conditions, ",")
	if condstr != "" {
		condstr = "," + condstr
	}
	p.i.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("testfor @a[name=\"%s\"%s]", p.userName, condstr)).AsyncGetResult(func(output *packet.CommandOutput) {
		ok := output.SuccessCount != 0
		go onResult(ok)
	})
}

func (p *PlayerKit) Query(onResult func([]neomega.QueryResult), conditions ...string) {
	condstr := strings.Join(conditions, ",")
	if condstr != "" {
		condstr = "," + condstr
	}
	var QueryResults []neomega.QueryResult

	p.i.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("querytarget @a[name=\"%s\"%s]", p.userName, condstr)).AsyncGetResult(func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			for _, v := range output.OutputMessages {
				for _, j := range v.Parameters {
					err := json.Unmarshal([]byte(j), &QueryResults)
					if err != nil {
						go onResult(nil)
					}
					go onResult(QueryResults)
				}
			}

		}
	})
}

func (p *PlayerKit) GetName() (name string, found bool) {
	return p.userName, p.PlayerUQReader.StillOnline()
}

func (p *PlayerKit) SetAbility(flagsUpdateMap, actionUpdateMap map[uint32]bool) bool {
	// 获取现有的数据
	flags, ok := p.GetPropertiesFlag()
	if !ok {
		return false
	}
	actionPermissions, ok := p.GetActionPermissions()
	if !ok {
		return false
	}
	// 函数, 对 flags 等进行位修改
	setBit := func(value *uint32, targetBit uint32, enable bool) {
		if enable {
			*value |= targetBit
		} else {
			*value &^= targetBit
		}
	}
	// 在现有数据的基础上进行修改
	for targetBit, enable := range flagsUpdateMap {
		setBit(&flags, targetBit, enable)
	}
	for targetBit, enable := range actionUpdateMap {
		setBit(&actionPermissions, targetBit, enable)
	}
	return p.sendAdventureSettingsPacket(flags, actionPermissions)
}

// GetAbilityString() (adventureFlagsMap, actionPermissionMap map[string]bool, found bool)
func (p *PlayerKit) SetAbilityString(adventureFlagsUpdateMap, actionPermissionUpdateMap map[string]bool) (sent bool) {
	adventureFlagsUpdateMapUint32 := make(map[uint32]bool)
	actionPermissionUpdateMapUint32 := make(map[uint32]bool)
	for flagName, flagValue := range uqholder.ActionPermissionMap {
		if enable, ok := actionPermissionUpdateMap[flagName]; ok {
			actionPermissionUpdateMapUint32[flagValue] = enable
		}
	}
	for flagName, flagValue := range uqholder.AdventureFlagMap {
		if enable, ok := adventureFlagsUpdateMap[flagName]; ok {
			adventureFlagsUpdateMapUint32[flagValue] = enable
		}
	}
	return p.SetAbility(adventureFlagsUpdateMapUint32, actionPermissionUpdateMapUint32)
}

// 发送 AdventureSettings 数据包, 且自动设置部分参数
func (p *PlayerKit) sendAdventureSettingsPacket(flags, actionPermissions uint32) bool {
	var commandPermissionLevel uint32
	var permissionLevel uint32
	// 获取 playerUniqueID
	playerUniqueID, ok := p.GetEntityUniqueID()
	if !ok {
		return false
	}
	// 租赁服OP命令执行权限为Host(3级)
	if actionPermissions&packet.ActionPermissionOperator != 0 {
		commandPermissionLevel = packet.CommandPermissionLevelHost
	}
	// 根据 actionPermissions 来决定 permissionLevel (访客, 成员, 操作员, 自定义)
	switch actionPermissions {
	case 447:
		permissionLevel = packet.PermissionLevelOperator
	case 287:
		permissionLevel = packet.PermissionLevelMember
	default:
		if actionPermissions != 0 {
			permissionLevel = packet.PermissionLevelCustom
		}
	}
	// 构建数据包
	packet := &packet.AdventureSettings{
		Flags:                  flags,
		CommandPermissionLevel: commandPermissionLevel,
		ActionPermissions:      actionPermissions,
		PermissionLevel:        permissionLevel,
		PlayerUniqueID:         playerUniqueID,
	}
	// 发送数据包
	p.i.gameIntractable.SendPacket(packet)
	// 在服务器响应修改前(没有变化或者修改失败时不会响应), 先进行一次手动修改
	p.i.playersUQ.UpdateFromPacket(packet)
	return true
}
