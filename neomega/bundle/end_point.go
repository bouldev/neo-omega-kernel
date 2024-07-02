package bundle

import (
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/modules/cmd_sender"
	"github.com/OmineDev/neomega-core/neomega/modules/core"
	"github.com/OmineDev/neomega-core/neomega/uqholder"
	"github.com/OmineDev/neomega-core/nodes/defines"
)

func NewEndPointMicroOmega(node defines.Node) (neomega.MicroOmega, error) {
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
