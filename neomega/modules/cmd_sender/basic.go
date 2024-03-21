package cmd_sender

import (
	"context"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/py_rpc"
	"neo-omega-kernel/utils/sync_wrapper"

	"github.com/google/uuid"
)

type CmdSenderBasic struct {
	neomega.InteractCore
	cbByUUID *sync_wrapper.SyncKVMap[string, func(*packet.CommandOutput)]
}

// func (c *CmdSender) SendWebSocketCmdOmitResponse(cmd string) {
// 	ud, _ := uuid.NewUUID()
// 	c.SendPacket(c.packCmdWithUUID(cmd, ud, true))
// }

func (c *CmdSenderBasic) onNewCommandOutput(p *packet.CommandOutput) {
	s := p.CommandOrigin.UUID.String()
	cb, ok := c.cbByUUID.GetAndDelete(s)
	if ok {
		cb(p)
	}
}

func (c *CmdSenderBasic) onAICommandEvent(eventName string, eventArgs map[string]any) {
	switch eventName {
	case "AfterExecuteCommandEvent":
		if cmdExecuteResult, _ := eventArgs["executeResult"].(bool); cmdExecuteResult {
			return
		}
		if cmdUUID, ok := eventArgs["uuid"].(string); ok {
			c.cbByUUID.Delete(cmdUUID)
		}
	case "ExecuteCommandOutputEvent":
		cmdMsg, ok := eventArgs["msg"].(string)
		if !ok {
			return
		}
		cmdUUID, ok := eventArgs["uuid"].(string)
		if !ok {
			return
		}
		cb, ok := c.cbByUUID.GetAndDelete(cmdUUID)
		if !ok {
			return
		}
		// CommandOutput packet can be received, but it hold an invalid(random) RequestID
		fakeResp := &packet.CommandOutput{
			CommandOrigin: protocol.CommandOrigin{
				RequestID: cmdUUID,
			},
			OutputMessages: []protocol.CommandOutputMessage{
				{
					Success: true,
					Message: cmdMsg,
				},
			},
		}
		cb(fakeResp)
	}
}

func (c *CmdSenderBasic) onNewPyRpc(p *packet.PyRpc) {
	pkt, ok := p.Value.MakeGo().([]any)
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
	eventArgs, ok := pktData[3].(map[string]any)
	if !ok {
		return
	}
	c.onAICommandEvent(eventName, eventArgs)
}

func NewCmdSenderBasic(reactable neomega.ReactCore, interactable neomega.InteractCore) *CmdSenderBasic {
	c := &CmdSenderBasic{
		InteractCore: interactable,
		cbByUUID:     sync_wrapper.NewSyncKVMap[string, func(*packet.CommandOutput)](),
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

func (c *CmdSenderBasic) SendAICommandNeedResponse(runtimeid string, cmd string) neomega.ResponseHandle {
	ud, _ := uuid.NewUUID()
	pkt := c.packAICmdWithUUID(runtimeid, cmd, ud)
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

func (c *CmdSenderBasic) packCmdWithUUID(cmd string, ud uuid.UUID, ws bool) *packet.CommandRequest {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
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
		Value: py_rpc.FromGo([]any{
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
		}),
	}
	return commandRequest
}
