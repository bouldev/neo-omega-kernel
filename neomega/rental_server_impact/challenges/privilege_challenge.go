package challenges

import (
	"context"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"time"
)

type OperatorChallenge struct {
	neomega.MicroOmega
	hasOpPrivilege  bool
	cheatOn         bool
	lostPrivilegeCB func()
}

func NewOperatorChallenge(omega neomega.MicroOmega, lostPrivilegeCallBack func()) *OperatorChallenge {
	if lostPrivilegeCallBack == nil {
		lostPrivilegeCallBack = func() {
			panic(fmt.Errorf(i18n.T(i18n.S_bot_op_privilege_removed)))
		}
	}
	helper := &OperatorChallenge{
		MicroOmega:      omega,
		lostPrivilegeCB: lostPrivilegeCallBack,
	}
	omega.GetGameListener().AddNewNoBlockAndDetachablePacketCallBack(
		map[uint32]bool{
			packet.IDAddPlayer:       true,
			packet.IDUpdateAbilities: true,
		},
		func(pk packet.Packet) error {
			switch pk.ID() {
			case packet.IDAddPlayer:
				pkt := pk.(*packet.AddPlayer)
				helper.onPermissionChange(pkt.AbilityData)
			case packet.IDUpdateAbilities:
				pkt := pk.(*packet.UpdateAbilities)
				helper.onPermissionChange(pkt.AbilityData)
			}
			return nil
		},
	)
	omega.GetGameListener().SetTypedPacketCallBack(packet.IDSetCommandsEnabled, helper.onSetCommandEnabledPacket, false)
	return helper
}

func (o *OperatorChallenge) onPermissionChange(abilityData protocol.AbilityData) {
	if o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID() == abilityData.EntityUniqueID {
		if abilityData.CommandPermissions >= packet.CommandPermissionLevelHost {
			o.hasOpPrivilege = true
		} else {
			if o.hasOpPrivilege {
				o.lostPrivilegeCB()
			}
			o.hasOpPrivilege = false
		}
	}
}

func (o *OperatorChallenge) onSetCommandEnabledPacket(pk packet.Packet) {
	p := pk.(*packet.SetCommandsEnabled)
	o.cheatOn = p.Enabled
	if !o.cheatOn && o.hasOpPrivilege {
		o.GetGameControl().SendWOCmd("changesetting allow-cheats true")
		o.cheatOn = true
	}
}

func (o *OperatorChallenge) WaitForPrivilege(ctx context.Context) (err error) {
	for !o.hasOpPrivilege {
		time.Sleep(1 * time.Second)
		o.GetGameControl().SendWOCmd("changesetting allow-cheats true")
		if o.GetGameControl().SendWebSocketCmdNeedResponse("tp @s ~~~").BlockGetResult().SuccessCount > 0 {
			o.hasOpPrivilege = true
			o.cheatOn = true
		}
		if ctx.Err() != nil {
			return fmt.Errorf(i18n.T(i18n.S_bot_operator_privilege_timeout))
		}
		if !o.hasOpPrivilege {
			// o.GetGameControl().BotSay(i18n.T(i18n.S_please_grant_bot_operator_privilege))
			fmt.Println(i18n.T(i18n.S_please_grant_bot_operator_privilege))
		}
	}
	fmt.Println(i18n.T(i18n.S_bot_operator_privilege_granted))
	return nil
}
