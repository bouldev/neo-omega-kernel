package minimal_client_entry

import (
	"context"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/neomega/rental_server_impact/access_helper"
	"neo-omega-kernel/neomega/rental_server_impact/info_collect_utils"
	"neo-omega-kernel/nodes"
	"time"
)

const ENTRY_NAME = "omega_minimal_client"

func Entry(args *access_helper.ImpactOption) {
	fmt.Println(i18n.T(i18n.S_neomega_access_point_starting))

	if err := info_collect_utils.ReadUserInfoAndUpdateImpactOptions(args); err != nil {
		panic(err)
	}

	accessOption := access_helper.DefaultOptions()
	accessOption.ImpactOption = args
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ReasonWithPrivilegeStuff = true

	ctx := context.Background()
	node := nodes.NewLocalNode(ctx)
	node = nodes.NewGroup("neo-omega-kernel/neomega-core", node, false)
	omegaCore, err := access_helper.ImpactServer(context.Background(), node, accessOption)
	if err != nil {
		panic(err)
	}

	go func() {
		i := 0
		for {
			i++
			time.Sleep(time.Second)
			ret := omegaCore.GetGameControl().SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s ~~ %v", i)).BlockGetResult()
			fmt.Println(ret)
			fmt.Println(omegaCore.GetMicroUQHolder().GetBotDimension())
		}
	}()
	omegaCore.GetBotAction().DropItemFromHotBar(3)
	panic(<-omegaCore.Dead())
}
