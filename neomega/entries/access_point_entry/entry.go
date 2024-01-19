package access_point

import (
	"context"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/rental_server_impact/access_helper"
	"neo-omega-kernel/neomega/rental_server_impact/info_collect_utils"
	"neo-omega-kernel/nodes"
)

const ENTRY_NAME = "omega_access_point"

func Entry(args *Args) {
	fmt.Println(i18n.T(i18n.S_neomega_access_point_starting))
	impactOption := args.ImpactOption
	var err error

	if err := info_collect_utils.ReadUserInfoAndUpdateImpactOptions(impactOption); err != nil {
		panic(err)
	}

	accessOption := access_helper.DefaultOptions()
	accessOption.ImpactOption = args.ImpactOption
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ReasonWithPrivilegeStuff = true

	var omegaCore neomega.MicroOmega
	var node nodes.Node
	ctx := context.Background()
	{
		socket, err := nodes.CreateZMQServerSocket(args.AccessArgs.AccessPointAddr)
		if err != nil {
			panic(err)
		}
		server := nodes.NewSimpleZMQServer(socket)
		master := nodes.NewZMQMasterNode(server)
		node = nodes.NewGroup("neo-omega-kernel/neomega", master, false)
	}
	omegaCore, err = access_helper.ImpactServer(ctx, node, accessOption)
	if err != nil {
		panic(err)
	}
	node.SetTags("access-point-ready")
	node.PublishMessage("reboot", nodes.FromString("reboot to refresh data"))
	fmt.Println(i18n.T(i18n.S_neomega_access_point_ready))

	panic(<-omegaCore.Dead())
}
