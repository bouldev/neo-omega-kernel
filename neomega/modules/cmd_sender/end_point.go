package cmd_sender

import (
	"context"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/nodes"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(sender neomega.CmdSender) {}(&EndPointCmdSender{})
	}
}

type EndPointCmdSender struct {
	*CmdSenderBasic
	node nodes.APINode
}

func NewEndPointCmdSender(node nodes.APINode, reactable neomega.ReactCore, interactable neomega.InteractCore) neomega.CmdSender {
	c := &EndPointCmdSender{
		CmdSenderBasic: NewCmdSenderBasic(reactable, interactable),
		node:           node,
	}
	return c
}

func (c *EndPointCmdSender) SendPlayerCmdNeedResponse(cmd string) neomega.ResponseHandle {
	ud, _ := uuid.NewUUID()
	args := nodes.FromString(cmd).Extend(nodes.FromUUID(ud))
	deferredAction := func() {
		c.node.CallOmitResponse("send-player-command", args)
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
