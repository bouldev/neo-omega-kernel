package bundle

import (
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/modules/block/placer"
	"neo-omega-kernel/neomega/modules/bot_action"
	"neo-omega-kernel/neomega/modules/info_sender"
	"neo-omega-kernel/neomega/modules/player_interact"
	"neo-omega-kernel/neomega/modules/structure"
	"neo-omega-kernel/nodes"
	"sync"
	"time"
)

func init() {
	if false {
		func(omega neomega.MicroOmega) {}(&MicroOmega{})
	}
}

type MicroOmega struct {
	neomega.ReactCore
	neomega.InteractCore
	neomega.InfoSender
	neomega.CmdSender
	neomega.MicroUQHolder
	neomega.BlockPlacer
	neomega.PlayerInteract
	neomega.StructureRequester
	neomega.CommandHelper
	neomega.BotAction
	neomega.BotActionHighLevel
	deferredActions []struct {
		cb   func()
		name string
	}
	mu sync.Mutex
}

func NewMicroOmega(
	interactCore neomega.InteractCore,
	reactCore neomega.UnStartedReactCore,
	microUQHolder neomega.MicroUQHolder,
	cmdSender neomega.CmdSender,
	node nodes.Node,
	isAccessPoint bool,
) neomega.UnReadyMicroOmega {
	infoSender := info_sender.NewInfoSender(interactCore, cmdSender, microUQHolder.GetBotBasicInfo())
	playerInteract := player_interact.NewPlayerInteract(reactCore, microUQHolder.GetPlayersInfo(), microUQHolder.GetBotBasicInfo(), cmdSender, infoSender, interactCore)
	asyncNbtBlockPlacer := placer.NewAsyncNbtBlockPlacer(reactCore, cmdSender, interactCore)
	structureRequester := structure.NewStructureRequester(interactCore, reactCore, microUQHolder)
	cmdHelper := bot_action.NewCommandHelper(cmdSender, microUQHolder)
	var botAction neomega.BotAction
	if isAccessPoint {
		botAction = bot_action.NewAccessPointBotActionWithPersistData(microUQHolder, interactCore, reactCore, cmdSender, node)
	} else {
		botAction = bot_action.NewEndPointBotAction(node, microUQHolder, interactCore)
	}

	botActionHighLevel := bot_action.NewBotActionHighLevel(microUQHolder, interactCore, reactCore, cmdSender, cmdHelper, structureRequester, asyncNbtBlockPlacer, botAction, node)
	omega := &MicroOmega{
		reactCore,
		interactCore,
		infoSender,
		cmdSender,
		microUQHolder,
		asyncNbtBlockPlacer,
		playerInteract,
		structureRequester,
		cmdHelper,
		botAction,
		botActionHighLevel,
		make([]struct {
			cb   func()
			name string
		}, 0),
		sync.Mutex{},
	}

	if isAccessPoint {
		omega.PostponeActionsAfterChallengePassed("request tick update schedule", func() {
			go func() {
				for {
					clientTick := 0
					if tick, found := omega.GetMicroUQHolder().GetExtendInfo().GetCurrentTick(); found {
						clientTick = int(tick)
					}
					omega.GetGameControl().SendPacket(&packet.TickSync{
						ClientRequestTimestamp: int64(clientTick),
					})
					time.Sleep(time.Second * 5)
				}
			}()
		})
	}

	omega.PostponeActionsAfterChallengePassed("dial tick every 1/20 second", func() {
		go func() {
			startTime := time.Now()
			tickAdd := int64(0)
			for {
				// sleep in some platform (yes, you, windows!) is not very accurate
				tickToAdd := (time.Now().Sub(startTime).Milliseconds() / 50) - tickAdd
				if tickToAdd > 0 {
					tickAdd += tickToAdd
					if tick, found := omega.GetMicroUQHolder().GetExtendInfo().GetCurrentTick(); found {
						omega.GetMicroUQHolder().GetExtendInfo().UpdateFromPacket(&packet.TickSync{
							ClientRequestTimestamp:   0,
							ServerReceptionTimestamp: tick + tickToAdd,
						})
					}
				}
				time.Sleep(time.Second / 20)
			}
		}()
	})
	reactCore.Start()
	return omega
}

func (o *MicroOmega) GetGameControl() neomega.GameCtrl {
	return o
}

func (o *MicroOmega) GetGameListener() neomega.PacketDispatcher {
	return o
}

func (o *MicroOmega) GetPlayerInteract() neomega.PlayerInteract {
	return o
}

func (o *MicroOmega) GetMicroUQHolder() neomega.MicroUQHolder {
	return o
}

func (o *MicroOmega) GetStructureRequester() neomega.StructureRequester {
	return o
}

func (o *MicroOmega) GetBotAction() neomega.BotActionComplex {
	return o
}

func (o *MicroOmega) NotifyChallengePassed() {
	for _, action := range o.deferredActions {
		fmt.Printf(i18n.T(i18n.S_starting_post_challenge_actions), action.name)
		action.cb()
	}
}

func (o *MicroOmega) PostponeActionsAfterChallengePassed(name string, action func()) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.deferredActions = append(o.deferredActions, struct {
		cb   func()
		name string
	}{action, name})
}
