package cmd_sender

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/nodes/defines"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(sender neomega.CmdSender) {}(&AccessPointCmdSender{})
	}
}

type needFeedBackPacket struct {
	do     func()
	waitor chan struct{}
}

type AccessPointCmdSender struct {
	*CmdSenderBasic
	expectedCmdFeedback     bool
	currentCmdFeedback      bool
	cmdFeedbackOffSent      bool
	commandFeedbackCtrlLock sync.Mutex
	needFeedBackPackets     []*needFeedBackPacket
}

func NewAccessPointCmdSender(node defines.APINode, reactable neomega.ReactCore, interactable neomega.InteractCore) neomega.CmdSender {
	c := &AccessPointCmdSender{
		CmdSenderBasic:          NewCmdSenderBasic(reactable, interactable),
		expectedCmdFeedback:     false,
		currentCmdFeedback:      false,
		cmdFeedbackOffSent:      false,
		commandFeedbackCtrlLock: sync.Mutex{},
		needFeedBackPackets:     make([]*needFeedBackPacket, 0),
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second*10)
	fmt.Printf(i18n.T(i18n.S_detecting_send_command_feedback_mode) + "\n")
	reactable.SetTypedPacketCallBack(packet.IDGameRulesChanged, func(p packet.Packet) {
		for _, rule := range p.(*packet.GameRulesChanged).GameRules {
			if rule.Name == "sendcommandfeedback" {
				feedbackOn := rule.Value.(bool)
				if ctx.Err() == nil {
					c.expectedCmdFeedback = feedbackOn
					done()
				}
				if feedbackOn {
					c.onCommandFeedbackOn()
				} else {
					c.onCommandFeedBackOff()
				}
			}
		}
	}, false)
	go func() {
		<-ctx.Done()
		fmt.Printf(i18n.T(i18n.S_send_commandfeedback_is_set_to_be)+"\n", c.expectedCmdFeedback)
	}()
	node.ExposeAPI("send-player-command", func(args defines.Values) (result defines.Values, err error) {
		cmd, err := args.ToString()
		if err != nil {
			return
		}
		args = args.ConsumeHead()
		ud, err := args.ToUUID()
		if err != nil {
			return
		}
		pkt := c.packCmdWithUUID(cmd, ud, false)
		do := func() {
			c.SendPacket(pkt)
		}
		c.launchOrDeferPlayerCommand(do)
		return defines.Empty, nil
	}, false)
	node.ExposeAPI("send-ai-command", func(args defines.Values) (result defines.Values, err error) {
		// fmt.Println("ai command")
		runtimeid, err := args.ToString()
		if err != nil {
			return
		}
		args = args.ConsumeHead()
		cmd, err := args.ToString()
		if err != nil {
			return
		}
		args = args.ConsumeHead()
		ud, err := args.ToUUID()
		if err != nil {
			return
		}
		pkt := c.packAICmdWithUUID(runtimeid, cmd, ud)
		do := func() {
			c.SendPacket(pkt)
		}
		c.launchOrDeferPlayerCommand(do)
		return defines.Empty, nil
	}, false)
	// deduce command feed back
	// no we cannot do this since bot currently not pass the challenge
	// go func() {
	// 	fmt.Println("deducing command feed back status")
	// 	output := c.SendWebSocketCmdNeedResponse("gamerule sendcommandfeedback").BlockGetResult()
	// 	fmt.Println(output.DataSet)
	// 	if strings.Contains(output.DataSet, "false") {
	// 		fmt.Println("sendcommandfeedback is deduced to be false")
	// 		c.currentCmdFeedback = false
	// 		c.expectedCmdFeedback = false
	// 	} else {
	// 		fmt.Println("sendcommandfeedback is deduced to be true")
	// 		c.currentCmdFeedback = true
	// 		c.expectedCmdFeedback = true
	// 	}
	// 	fmt.Println("deducing command feed back status done")
	// }()

	// if c.expectedCmdFeedback {
	// 	c.SendWOCmd("gamerule sendcommandfeedback true")
	// } else {
	// 	c.SendWOCmd("gamerule sendcommandfeedback false")
	// }

	return c
}

func (c *AccessPointCmdSender) SendPlayerCmdNeedResponse(cmd string) neomega.ResponseHandle {
	ud, _ := uuid.NewUUID()
	pkt := c.packCmdWithUUID(cmd, ud, false)
	deferredAction := func() {
		do := func() {
			c.SendPacket(pkt)
		}
		c.launchOrDeferPlayerCommand(do)
	}
	return &CmdResponseHandle{
		deferredActon:         deferredAction,
		timeoutSpecificResult: nil,
		terminated:            false,
		uuidStr:               ud.String(),
		cbByUUID:              c.cbByUUID,
		ctx:                   context.Background(),
	}
}

func (c *AccessPointCmdSender) SendAICommandNeedResponse(runtimeid string, cmd string) neomega.ResponseHandle {
	ud, _ := uuid.NewUUID()
	pkt := c.packAICmdWithUUID(runtimeid, cmd, ud)
	deferredAction := func() {
		do := func() {
			c.SendPacket(pkt)
		}
		c.launchOrDeferPlayerCommand(do)
	}
	return &CmdResponseHandle{
		deferredActon:         deferredAction,
		timeoutSpecificResult: nil,
		terminated:            false,
		uuidStr:               ud.String(),
		cbByUUID:              c.cbByUUID,
		ctx:                   context.Background(),
	}
}

func (c *AccessPointCmdSender) launchOrDeferPlayerCommand(do func()) {
	c.commandFeedbackCtrlLock.Lock()
	defer c.commandFeedbackCtrlLock.Unlock()
	if !c.cmdFeedbackOffSent {
		// the major problem is, whether sendcommandfeedback is on and user expect sendcommandfeedback on? If both, just send
		if c.expectedCmdFeedback && c.currentCmdFeedback {
			c.SendWOCmd("gamerule sendcommandfeedback true")
			do()
			return
		}
		// and if not, there exists two problems:
		// 1. we need to turn on commandfeedback, and then turn off
		// 2. the effect on server and what we receive is not sync
		if !c.expectedCmdFeedback && c.currentCmdFeedback {
			c.SendWOCmd("gamerule sendcommandfeedback true")
			do()
			return
		}
	}

	// note that above two conditions equals to if c.currentCmdFeedBack {c.SendPacket(pkt)}

	waitor := make(chan struct{})
	if c.cmdFeedbackOffSent {
		// server and client status not sync, just put packets in pending queue
		c.needFeedBackPackets = append(c.needFeedBackPackets, &needFeedBackPacket{
			do, waitor,
		})
	} else {
		// if !c.currentCmdFeedback {
		// fmt.Println("send sendcommandfeedback true")
		c.SendWOCmd("gamerule sendcommandfeedback true")

		c.needFeedBackPackets = append(c.needFeedBackPackets, &needFeedBackPacket{
			do, waitor,
		})
		// }
	}

	go func() {
		select {
		case <-waitor:
		case <-time.NewTimer(time.Millisecond * 50 * 4).C:
			// for some unknown reason, in some server just cannot receive sendcommandfeedback true notify, so we emit these pending packet anyway
			c.emitNeedFeedBackPackets(true)
		}
	}()
}

func (c *AccessPointCmdSender) emitNeedFeedBackPackets(force bool) {
	c.SendWOCmd("gamerule sendcommandfeedback true")
	if force || (!c.cmdFeedbackOffSent) {
		pkts := c.needFeedBackPackets
		c.needFeedBackPackets = make([]*needFeedBackPacket, 0)
		for _, p := range pkts {
			p.do()
			close(p.waitor)
		}
	}
}

func (c *AccessPointCmdSender) onCommandFeedbackOn() {
	c.commandFeedbackCtrlLock.Lock()
	defer c.commandFeedbackCtrlLock.Unlock()
	// fmt.Println("recv sendcommandfeedback true")
	c.currentCmdFeedback = true
	c.emitNeedFeedBackPackets(false)
	if !c.expectedCmdFeedback {
		go func() {
			time.Sleep(time.Millisecond * 50 * 4) //wait 4 ticks
			c.commandFeedbackCtrlLock.Lock()
			defer c.commandFeedbackCtrlLock.Unlock()
			if c.currentCmdFeedback && !c.cmdFeedbackOffSent {
				// fmt.Println("send sendcommandfeedback false")
				c.SendWOCmd("gamerule sendcommandfeedback false")
				c.cmdFeedbackOffSent = true
			}
		}()
	}
}

func (c *AccessPointCmdSender) onCommandFeedBackOff() {
	c.commandFeedbackCtrlLock.Lock()
	defer c.commandFeedbackCtrlLock.Unlock()
	// fmt.Println("recv sendcommandfeedback false")
	c.cmdFeedbackOffSent = false
	c.currentCmdFeedback = false
	if (c.expectedCmdFeedback || len(c.needFeedBackPackets) > 0) && !c.cmdFeedbackOffSent {
		// fmt.Println("send sendcommandfeedback true")
		c.SendWOCmd("gamerule sendcommandfeedback true")
	}
}
