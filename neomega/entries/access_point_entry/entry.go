package access_point

import (
	"context"
	"fmt"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/access_helper"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/info_collect_utils"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
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
	var node defines.Node
	ctx := context.Background()
	{
		server, err := underlay_conn.NewServerFromBasicNet(args.AccessArgs.AccessPointAddr)
		if err != nil {
			panic(err)
		}
		master := nodes.NewZMQMasterNode(server)
		node = nodes.NewGroup("github.com/OmineDev/neomega-core/neomega", master, false)
	}
	omegaCore, err = access_helper.ImpactServer(ctx, node, accessOption)
	if err != nil {
		panic(err)
	}
	node.SetTags("access-point-ready")
	node.PublishMessage("reboot", defines.FromString("reboot to refresh data"))
	fmt.Println(i18n.T(i18n.S_neomega_access_point_ready))

	panic(<-omegaCore.Dead())
}
