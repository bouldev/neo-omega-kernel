package bundle

import (
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/modules/cmd_sender"
	"neo-omega-kernel/neomega/modules/core"
	"neo-omega-kernel/neomega/uqholder"
	"neo-omega-kernel/nodes"
)

func NewEndPointMicroOmega(node nodes.Node) (neomega.MicroOmega, error) {
	reactCore := core.NewEndPointReactCore(node)
	interactCore, err := core.NewEndPointInteractCore(node, reactCore)
	if err != nil {
		return nil, err
	}
	microUQHolder, err := uqholder.NewEndPointMicroUQHolder(node, reactCore)
	if err != nil {
		return nil, err
	}
	cmdSender := cmd_sender.NewEndPointCmdSender(node, reactCore, interactCore)
	unReadyMicroOmega := NewMicroOmega(interactCore, reactCore, microUQHolder, cmdSender, node, false)
	// access must have passed challenge
	unReadyMicroOmega.NotifyChallengePassed()
	return unReadyMicroOmega, nil
}
