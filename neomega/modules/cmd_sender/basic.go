package cmd_sender

import (
	"context"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
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
	cb, ok := c.cbByUUID.Get(s)
	if ok {
		cb(p)
	}
}

func NewCmdSenderBasic(reactable neomega.ReactCore, interactable neomega.InteractCore) *CmdSenderBasic {
	c := &CmdSenderBasic{
		InteractCore: interactable,
		cbByUUID:     sync_wrapper.NewSyncKVMap[string, func(*packet.CommandOutput)](),
	}
	reactable.SetTypedPacketCallBack(packet.IDCommandOutput, func(p packet.Packet) {
		// fmt.Println(p)
		c.onNewCommandOutput(p.(*packet.CommandOutput))
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
