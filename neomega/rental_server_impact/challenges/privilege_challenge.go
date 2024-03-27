package challenges

import (
	"context"
	"fmt"
	"neo-omega-kernel/i18n"
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
	omega.GetGameListener().SetTypedPacketCallBack(packet.IDAdventureSettings, helper.onAdventurePacket, false)
	omega.GetGameListener().SetTypedPacketCallBack(packet.IDSetCommandsEnabled, helper.onSetCommandEnabledPacket, false)
	return helper
}

func (o *OperatorChallenge) onAdventurePacket(pk packet.Packet) {
	p := pk.(*packet.AdventureSettings)
	if o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID() == p.PlayerUniqueID {
		if p.PermissionLevel >= packet.PermissionLevelOperator {
			o.hasOpPrivilege = true
			fmt.Println(i18n.T(i18n.S_bot_operator_privilege_granted))
		} else {
			if o.hasOpPrivilege {
				o.lostPrivilegeCB()
			}
			fmt.Println(i18n.T(i18n.S_please_grant_bot_operator_privilege))
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
	time.Sleep(1 * time.Second)
	for !o.hasOpPrivilege {
		o.GetGameControl().SendWOCmd("changesetting allow-cheats true")
		o.GetGameControl().SendWebSocketCmdNeedResponse("tp @s ~~~").AsyncGetResult(func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.hasOpPrivilege = true
				o.cheatOn = true
			}
		})
		if ctx.Err() != nil {
			return fmt.Errorf(i18n.T(i18n.S_bot_operator_privilege_timeout))
		}
		time.Sleep(1 * time.Second)
		if !o.hasOpPrivilege {
			// o.GetGameControl().BotSay(i18n.T(i18n.S_please_grant_bot_operator_privilege))
			fmt.Println(i18n.T(i18n.S_please_grant_bot_operator_privilege))
		}
	}
	fmt.Println(i18n.T(i18n.S_bot_operator_privilege_granted))
	return nil
}
