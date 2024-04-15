package cmd_sender

import (
	"context"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/utils/sync_wrapper"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type CmdSenderBasic struct {
	neomega.InteractCore
	cbByUUID          *sync_wrapper.SyncKVMap[string, func(*packet.CommandOutput)]
	aiCmdResultByUUID *sync_wrapper.SyncKVMap[string, *aiCommandOnGoingOutputs]
}

// func (c *CmdSender) SendWebSocketCmdOmitResponse(cmd string) {
// 	ud, _ := uuid.NewUUID()
// 	c.SendPacket(c.packCmdWithUUID(cmd, ud, true))
// }

func (c *CmdSenderBasic) onNewCommandOutput(p *packet.CommandOutput) {
	s := p.CommandOrigin.UUID.String()
	cb, ok := c.cbByUUID.Get(s)
	if ok {
		cb(p)
	}
}

type aiCommandOnGoingOutputs struct {
	outputs []string
	mu      sync.Mutex
}

func newAiCommandOnGoingOutputs() *aiCommandOnGoingOutputs {
	return &aiCommandOnGoingOutputs{
		outputs: make([]string, 0, 1),
		mu:      sync.Mutex{},
	}
}

func (o *aiCommandOnGoingOutputs) appendOutput(msg string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	inferredConsoleTranslatedOutputMsg := strings.TrimPrefix(msg, "命令输出：")
	o.outputs = append(o.outputs, inferredConsoleTranslatedOutputMsg)
}

func (o *aiCommandOnGoingOutputs) genInferredConsoleOutput(executeResult bool, origUUIDStr string) *packet.CommandOutput {
	outputMsgs := []protocol.CommandOutputMessage{}
	if o != nil {
		for _, msg := range o.outputs {
			outputMsgs = append(outputMsgs, protocol.CommandOutputMessage{
				Success: executeResult,
				Message: msg,
			})
		}
	}
	successCount := uint32(0) // seems related infer not provided
	if executeResult {
		successCount = 1
	}
	return &packet.CommandOutput{
		CommandOrigin: protocol.CommandOrigin{
			UUID:      uuid.MustParse(origUUIDStr),
			RequestID: "server-agent-in-ai-command",
		},
		SuccessCount:   successCount,
		OutputMessages: outputMsgs,
	}
}

func (c *CmdSenderBasic) onAICommandEvent(eventName string, eventArgs map[any]any) {
	switch eventName {
	case "AfterExecuteCommandEvent":
		// fmt.Println("AfterExecuteCommandEvent", eventArgs)
		executeResult, ok := eventArgs["executeResult"]
		if !ok {
			return
		}
		cmdExecuteResult := executeResult.(bool)
		uid, ok := eventArgs["uuid"]
		if !ok {
			return
		}
		cmdUUID := uid.(string)
		cb, ok := c.cbByUUID.GetAndDelete(cmdUUID)
		if !ok {
			return
		}
		// We can ensure all the "ExecuteCommandOutputEvent" are sent before "AfterExecuteCommandEvent"
		outputs, _ := c.aiCmdResultByUUID.GetAndDelete(cmdUUID)
		// CommandOutput packet can be received, but it hold an invalid(random) UUID
		inferredConsoleResp := outputs.genInferredConsoleOutput(cmdExecuteResult, cmdUUID)
		cb(inferredConsoleResp)
	case "ExecuteCommandOutputEvent":
		// fmt.Println("ExecuteCommandOutputEvent", eventArgs)
		msg, ok := eventArgs["msg"]
		if !ok {
			return
		}
		uuid, ok := eventArgs["uuid"]
		if !ok {
			return
		}
		// Cache result of cmd because it may have multiple outputs
		cmdUUID := uuid.(string)
		existOutputs, ok := c.aiCmdResultByUUID.Get(cmdUUID)
		if !ok {
			existOutputs = newAiCommandOnGoingOutputs()
			existOutputs, _ = c.aiCmdResultByUUID.GetOrSet(cmdUUID, existOutputs)
		}
		existOutputs.appendOutput(msg.(string))
	}
}

func (c *CmdSenderBasic) onNewPyRpc(p *packet.PyRpc) {
	pkt, ok := p.Value.([]any)
	if !ok || len(pkt) != 3 {
		return
	}
	if pktType, ok := pkt[0].(string); !ok || pktType != "ModEventS2C" {
		return
	}
	pktData, ok := pkt[1].([]any)
	if !ok || len(pktData) != 4 {
		return
	}
	if systemName, ok := pktData[1].(string); !ok || systemName != "aiCommand" {
		return
	}
	eventName, ok := pktData[2].(string)
	if !ok {
		return
	}
	eventArgs, ok := pktData[3].(map[any]any)
	if !ok {
		return
	}
	c.onAICommandEvent(eventName, eventArgs)
}

func NewCmdSenderBasic(reactable neomega.ReactCore, interactable neomega.InteractCore) *CmdSenderBasic {
	c := &CmdSenderBasic{
		InteractCore:      interactable,
		cbByUUID:          sync_wrapper.NewSyncKVMap[string, func(*packet.CommandOutput)](),
		aiCmdResultByUUID: sync_wrapper.NewSyncKVMap[string, *aiCommandOnGoingOutputs](),
	}
	reactable.SetTypedPacketCallBack(packet.IDCommandOutput, func(p packet.Packet) {
		c.onNewCommandOutput(p.(*packet.CommandOutput))
	}, false)
	reactable.SetTypedPacketCallBack(packet.IDPyRpc, func(p packet.Packet) {
		c.onNewPyRpc(p.(*packet.PyRpc))
	}, false)
	return c
}

func (c *CmdSenderBasic) SendWebSocketCmdOmitResponse(cmd string) {
	ud, _ := uuid.NewUUID()
	c.SendPacket(c.packCmdWithUUID(cmd, ud, true))
}

// func (c *CmdSender) SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool) {
// 	c.SendPacket(c.packCmdWithUUID(cmd, ud, ws))
// }

func (c *CmdSenderBasic) SendWOCmd(cmd string) {
	c.SendPacket(&packet.SettingsCommand{
		CommandLine:    cmd,
		SuppressOutput: true,
	})
}

func (c *CmdSenderBasic) SendPlayerCmdOmitResponse(cmd string) {
	ud, _ := uuid.NewUUID()
	pkt := c.packCmdWithUUID(cmd, ud, false)
	c.SendPacket(pkt)
}

func (c *CmdSenderBasic) SendWebSocketCmdNeedResponse(cmd string) neomega.ResponseHandle {
	ud, _ := uuid.NewUUID()
	pkt := c.packCmdWithUUID(cmd, ud, true)
	deferredAction := func() {
		c.SendPacket(pkt)
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

func (c *CmdSenderBasic) SendAICommandOmitResponse(runtimeid string, cmd string) {
	ud, _ := uuid.NewUUID()
	pkt := c.packAICmdWithUUID(runtimeid, cmd, ud)
	c.SendPacket(pkt)
}

func (c *CmdSenderBasic) packCmdWithUUID(cmd string, ud uuid.UUID, ws bool) *packet.CommandRequest {
	requestId := uuid.New()
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginAutomationPlayer,
		UUID:           ud,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	if !ws {
		origin.Origin = protocol.CommandOriginPlayer
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   cmd,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return commandRequest
}

func (c *CmdSenderBasic) packAICmdWithUUID(runtimeid string, cmd string, ud uuid.UUID) *packet.PyRpc {
	commandRequest := &packet.PyRpc{
		Value: []any{
			"ModEventC2S",
			[]any{
				"Minecraft",
				"aiCommand",
				"ExecuteCommandEvent",
				map[string]any{
					"playerId": runtimeid,
					"cmd":      cmd,
					"uuid":     ud.String(),
				},
			},
			nil,
		},
	}
	return commandRequest
}
