package bundle

import (
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/minecraft_conn"
	"github.com/OmineDev/neomega-core/neomega/modules/cmd_sender"
	"github.com/OmineDev/neomega-core/neomega/modules/core"
	"github.com/OmineDev/neomega-core/neomega/uqholder"
	"github.com/OmineDev/neomega-core/nodes/defines"
)

func NewAccessPointMicroOmega(node defines.Node, conn minecraft_conn.Conn) neomega.UnReadyMicroOmega {
	interactCore := core.NewAccessPointInteractCore(node, conn)
	reactCore := core.NewAccessPointReactCore(node, conn)
	microUQHolder := uqholder.NewAccessPointMicroUQHolder(node, conn, reactCore)

	cmdSender := cmd_sender.NewAccessPointCmdSender(node, reactCore, interactCore)
	return NewMicroOmega(interactCore, reactCore, microUQHolder, cmdSender, node, true)
}
