package bundle

import (
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/minecraft_conn"
	"neo-omega-kernel/neomega/modules/cmd_sender"
	"neo-omega-kernel/neomega/modules/core"
	"neo-omega-kernel/neomega/uqholder"
	"neo-omega-kernel/nodes"
)

func NewAccessPointMicroOmega(node nodes.Node, conn minecraft_conn.Conn) neomega.UnReadyMicroOmega {
	interactCore := core.NewAccessPointInteractCore(node, conn)
	reactCore := core.NewAccessPointReactCore(node, conn)
	microUQHolder := uqholder.NewAccessPointMicroUQHolder(node, conn, reactCore)

	cmdSender := cmd_sender.NewAccessPointCmdSender(node, reactCore, interactCore)
	return NewMicroOmega(interactCore, reactCore, microUQHolder, cmdSender, node, true)
}
