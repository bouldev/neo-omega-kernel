package minimal_end_point_entry

import (
	"bufio"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega/bundle"
	"neo-omega-kernel/nodes"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/nodes/underlay_conn"
	"os"
	"strings"
	"time"
)

const ENTRY_NAME = "omega_minimal_end_point"

func Entry(args *Args) {
	var node defines.Node
	// ctx := context.Background()
	{
		client, err := underlay_conn.NewClientFromBasicNet(args.AccessPointAddr, time.Second)
		if err != nil {
			panic(err)
		}
		slave, err := nodes.NewZMQSlaveNode(client)
		if err != nil {
			panic(err)
		}
		node = nodes.NewGroup("neo-omega-kernel/neomega", slave, false)
		node.ListenMessage("reboot", func(msg defines.Values) {
			reason, _ := msg.ToString()
			fmt.Println(reason)
			os.Exit(3)
		}, false)
		if !node.CheckNetTag("access-point") {
			panic(i18n.T(i18n.S_no_access_point_in_network))
		}
		for {
			if node.CheckNetTag("access-point-ready") {
				break
			}
			time.Sleep(time.Second)
		}
	}
	omegaCore, err := bundle.NewEndPointMicroOmega(node)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
	fmt.Println(omegaCore)
	// go func() {
	// 	i := 0
	// 	for {
	// 		i++
	// 		time.Sleep(time.Second)
	// 		ret := omegaCore.GetGameControl().SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s ~~ %v", i)).BlockGetResult()
	// 		fmt.Println(ret)
	// 	}
	// }()

	// go func() {
	// 	i := 0
	// 	for {
	// 		i++
	// 		time.Sleep(time.Second)
	// 		pos, _ := omegaCore.GetMicroUQHolder().GetExtendInfo().GetBotPosition()
	// 		dimension, _ := omegaCore.GetMicroUQHolder().GetBotDimension()
	// 		fmt.Printf("%v %v %v %v %v\n", i, pos.X(), pos.Y(), pos.Z(), dimension)
	// 	}
	// }()

	// omegaCore.GetGameListener().SetTypedPacketCallBack(packet.IDItemStackResponse, func(p packet.Packet) {
	// 	pk := p.(*packet.ItemStackResponse)
	// 	for _, r := range pk.Responses {
	// 		fmt.Printf("status: %v id: %v\n", r.Status, r.RequestID)
	// 		for _, slot := range r.ContainerInfo {
	// 			bs, _ := json.Marshal(slot)
	// 			fmt.Printf("\t%v\n", string(bs))
	// 		}
	// 	}
	// }, true)

	// go func() {
	// 	for {
	// 		picked := <-pickedItemChan
	// 		bs, _ := json.Marshal(picked.NewItem)
	// 		fmt.Printf("type: %v window: %v slot: %v value: %v\n", picked.SourceType, picked.WindowID, picked.InventorySlot, string(bs))
	// 	}
	// }()

	// omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{260, -60, 35}, 1, "planks", 5, 2, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()

	// omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{260, -60, 35}, 4, "iron_sword", 5, 4, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()
	// time.Sleep(time.Second)

	// origBlock := define.CubePos{265, -59, 47}
	// block, err := omegaCore.GetStructureRequester().RequestStructure(origBlock, define.CubePos{1, 1, 1}, "block").BlockGetResult()
	// if err != nil {
	// 	panic(err)
	// }
	// decoded, err := block.Decode()
	// blockName, _ := blocks.RuntimeIDToBlockNameWithStateStr(decoded.ForeGround[0])
	// fmt.Println(blockName)
	// nbtData := decoded.Nbts[origBlock]["Item"]
	// fmt.Println(nbtData)
	// item, err := neomega.GenItemInfoFromItemFrameNBT(nbtData)
	// fmt.Println(item)
	// err = omegaCore.GetBotAction().HighLevelMakeItem(item, 0, origBlock.Add(define.CubePos{1, 0, -1}), origBlock.Add(define.CubePos{1, 0, 1}))
	// fmt.Println("make err", err)
	// err = omegaCore.GetBotAction().HighLevelPlaceItemFrameItem(define.CubePos{259, -60, 54}, 0)
	// fmt.Println("put err", err)
	// read out nbt data of a specific container
	// then re-make it
	// origChest := define.CubePos{264, -60, 46}
	// block, err := omegaCore.GetStructureRequester().RequestStructure(origChest, define.CubePos{1, 1, 1}, "block").BlockGetResult()
	// if err != nil {
	// 	panic(err)
	// }
	// decoded, err := block.Decode()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(decoded)
	// nbtData := decoded.Nbts[origChest]["Items"]
	// containerInfo, _ := neomega.GenContainerItemsInfoFromItemsNbt(nbtData.([]any))
	// fmt.Println(containerInfo)
	// blockName, found := blocks.RuntimeIDToBlockNameWithStateStr(decoded.ForeGround[0])
	// if !found {nodes
	// 	panic(err)
	// }
	// fmt.Println(blockName)
	// err = omegaCore.GetBotAction().HighLevelGenContainer(define.CubePos{260, -60, 46}, containerInfo, blockName)
	// fmt.Println(err)

	// containerBlock.
	// 	omegaCore.B.HighLevelGenContainer()
	// HighLevelSetContainerItems(omegaCore, containerInfo, define.CubePos{257, -60, 36})
	// fmt.Println(containerInfo)

	// os.WriteFile(fmt.Sprintf("%v.nbt", decoded.ForeGround[0]), bs, 0755)
	// // set container item
	// result := omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{264, -60, 46}, 0, "stone", 10, 0, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn: []string{"grass", "stone"},
	// 	CanDestroy: []string{"sand"},
	// }).SendAndGetResponse().BlockGetResult()
	// fmt.Println(result)

	// // make item, enchant, and drop
	// result := omegaCore.GetBotAction().ReplaceBotHotBarItemFullCmd(2, "diamond_sword", 10, 0, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass", "stone"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()
	// fmt.Println(result)
	// err = omegaCore.GetBotAction().HighLevelEnchantItem(2, map[string]int32{
	// 	"9":          2,
	// 	"30":         2,
	// 	"unbreaking": 2,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(2)
	// if err != nil {
	// 	panic(err)
	// }

	// // put item in container
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(4, "oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(5, "dark_oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()

	// err = omegaCore.GetBotAction().HighLevelMoveItemToContainer(define.CubePos{264, -60, 46}, map[uint8]uint8{4: 7, 5: 10})
	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println("move ok")
	// }

	// // rename item use anvil
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// err = omegaCore.GetBotAction().HighLevelRenameItemWithAutoGenAnvil(define.CubePos{260, -60, 41}, 0, "240")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// // drop item from hotbar
	// err = omegaCore.GetBotAction().DropItemFromHotBar(0)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println("drop ok")
	// }

	// //listen item picked
	// pickedItemChan, _, err := omegaCore.GetBotAction().HighLevelListenItemPicked(time.Minute * 10)
	// if err != nil {
	// 	panic(err)
	// }

	// // break and pick block
	// _, err = omegaCore.GetBotAction().HighLevelBlockBreakAndPickInHotBar(define.CubePos{266, -60, 35}, true, map[uint8]bool{5: false, 6: false, 7: false}, 2)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("break and pick ok")
	// }

	// err = omegaCore.GetBotAction().DropItemFromHotBar(5)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(6)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(7)
	// if err != nil {
	// 	panic(err)
	// }

	// place sign
	// omegaCore.GetBotAction().HighLevelPlaceSign(
	// 	define.CubePos{262, -60, 49}, "2401!", true, "standing_sign",
	// )

	// // put command block
	// omegaCore.GetBotAction().HighLevelPlaceCommandBlock(&neomega.PlaceCommandBlockOption{
	// 	X: 263, Y: -60, Z: 49,
	// 	BlockName:    "command_block",
	// 	BlockState:   "0",
	// 	NeedRedStone: true,
	// 	Name:         "240!",
	// 	Command:      "say 240!",
	// }, 3)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			time.Sleep(time.Second / 3)
			fmt.Printf("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "/") {
				omegaCore.GetGameControl().SendWebSocketCmdNeedResponse(line).AsyncGetResult(func(output *packet.CommandOutput) {
					fmt.Println(output)
				})
				continue
			}
			if strings.HasPrefix(line, "player/") {
				omegaCore.GetGameControl().SendPlayerCmdNeedResponse(strings.TrimPrefix(line, "player/")).AsyncGetResult(func(output *packet.CommandOutput) {
					fmt.Println(output)
				})
				continue
			}
			if strings.HasPrefix(line, "#uq.") {
				line = strings.TrimPrefix(line, "#uq.")
				if line == "all_players" {
					for i, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
						name, _ := player.GetUsername()
						uuid, _ := player.GetUUIDString()
						fmt.Printf("%v %v %v\n", i, name, uuid)
					}
				}
				if line == "command_permission_level" {
					for _, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
						name, _ := player.GetUsername()
						level, found := player.GetCommandPermissionLevel()
						fmt.Printf("%v %v %v\n", name, level, found)
					}
				}
				if line == "op_permission_level" {
					for _, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
						name, _ := player.GetUsername()
						level, found := player.GetOPPermissionLevel()
						fmt.Printf("%v %v %v\n", name, level, found)
					}
				}
				continue
			}
		}
	}()

	panic(<-node.WaitClosed())
}
