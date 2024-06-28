package bot_action

import (
	"context"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/chunks/define"
	"neo-omega-kernel/nodes"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/utils/sync_wrapper"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type AccessPointBotActionWithPersistData struct {
	*BotActionSimple
	uq                               neomega.MicroUQHolder
	listener                         neomega.ReactCore
	ctrl                             neomega.InteractCore
	cmdSender                        neomega.CmdSender
	currentContainerOpenListener     func(*packet.ContainerOpen)
	currentContainerCloseListener    func(*packet.ContainerClose)
	currentItemStackResponseListener func(*packet.ItemStackResponse)
	clientInfo                       *MaintainedGameInfo
	node                             defines.APINode
	muChan                           chan struct{}
}

func NewAccessPointBotActionWithPersistData(
	uq neomega.MicroUQHolder, ctrl neomega.InteractCore, listener neomega.ReactCore, cmdSender neomega.CmdSender,
	node defines.Node,
) neomega.BotAction {
	ba := &AccessPointBotActionWithPersistData{
		BotActionSimple: NewBotActionSimple(uq, ctrl),
		uq:              uq,
		ctrl:            ctrl,
		listener:        listener,
		cmdSender:       cmdSender,
		clientInfo:      NewMaintainedGameInfo(listener),
		node:            nodes.NewGroup("bot_action", node, false),
		muChan:          make(chan struct{}, 1),
	}
	ba.muChan <- struct{}{}
	listener.SetTypedPacketCallBack(packet.IDRespawn, func(p packet.Packet) {
		pkt := p.(*packet.Respawn)
		rtid := uq.GetBotRuntimeID()
		ctrl.SendPacket(&packet.Respawn{
			EntityRuntimeID: rtid,
			Position:        pkt.Position,
			State:           packet.RespawnStateClientReadyToSpawn,
		})
		ctrl.SendPacket(&packet.PlayerAction{
			EntityRuntimeID: rtid,
			ActionType:      protocol.PlayerActionRespawn,
		})
	}, false)
	listener.SetTypedPacketCallBack(packet.IDChangeDimension, func(p packet.Packet) {
		// pkt := p.(*packet.ChangeDimension)
		rtid := uq.GetBotBasicInfo().GetBotRuntimeID()
		ctrl.SendPacket(&packet.PlayerAction{
			EntityRuntimeID: rtid,
			ActionType:      protocol.PlayerActionDimensionChangeDone,
		})
	}, false)
	listener.SetTypedPacketCallBack(packet.IDContainerOpen, func(p packet.Packet) {
		if ba.currentContainerOpenListener == nil {
			return
		}
		listener := ba.currentContainerOpenListener
		ba.currentContainerOpenListener = nil
		listener(p.(*packet.ContainerOpen))
	}, true)
	listener.SetTypedPacketCallBack(packet.IDContainerClose, func(p packet.Packet) {
		if ba.currentContainerCloseListener == nil {
			return
		}
		listener := ba.currentContainerCloseListener
		ba.currentContainerCloseListener = nil
		listener(p.(*packet.ContainerClose))
	}, true)
	listener.SetTypedPacketCallBack(packet.IDItemStackResponse, func(p packet.Packet) {
		// fmt.Println("item stack response")
		if ba.currentItemStackResponseListener == nil {
			return
		}
		listener := ba.currentItemStackResponseListener
		ba.currentItemStackResponseListener = nil
		listener(p.(*packet.ItemStackResponse))
	}, true)
	ba.selectHotBar(0)
	ba.ExposeAPI()
	return ba
}

func (o *AccessPointBotActionWithPersistData) occupyBot(timeout time.Duration) (release func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("cannot acquire bot, bot is busy")
	case <-o.muChan:
		if o.currentContainerOpenListener != nil || o.currentContainerCloseListener != nil || o.currentItemStackResponseListener != nil {
			panic(fmt.Errorf("another operation is already proceeding, but bot is not locked"))
		}
		return func() {
			o.muChan <- struct{}{} // give back bot control
		}, nil
	}
}

func (o *AccessPointBotActionWithPersistData) SelectHotBar(slotID uint8) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.selectHotBar(slotID)
}

func (o *AccessPointBotActionWithPersistData) selectHotBar(slotID uint8) error {
	if slotID > 8 {
		return fmt.Errorf("should be only in 0 ~ 8")
	}
	o.clientInfo.currentSlot = slotID
	o.ctrl.SendPacket(&packet.PlayerHotBar{
		SelectedHotBarSlot: uint32(slotID),
		WindowID:           0,
		SelectHotBarSlot:   true,
	})
	return nil
}

func (o *AccessPointBotActionWithPersistData) ensureHotBar(slotID uint8) error {
	if slotID != o.clientInfo.currentSlot {
		return o.selectHotBar(slotID)
	}
	return nil
}

func (o *AccessPointBotActionWithPersistData) UseHotBarItem(slotID uint8) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	if err = o.ensureHotBar(slotID); err != nil {
		return err
	}
	item, found := o.clientInfo.GetInventorySlot(0, slotID)
	if !found {
		return fmt.Errorf("slot is empty")
	}
	o.ctrl.SendPacket(
		&packet.InventoryTransaction{
			TransactionData: &protocol.UseItemTransactionData{
				ActionType: protocol.UseItemActionClickAir,
				HotBarSlot: int32(slotID),
				HeldItem:   *item,
			},
		},
	)
	return nil
}

func (o *AccessPointBotActionWithPersistData) UseHotBarItemOnBlock(blockPos define.CubePos, blockNEMCRuntimeID uint32, face int32, slotID uint8) (err error) {
	return o.UseHotBarItemOnBlockWithBotOffset(blockPos, define.CubePos{0, 0, 0}, blockNEMCRuntimeID, face, slotID)
}

func (o *AccessPointBotActionWithPersistData) UseHotBarItemOnBlockWithBotOffset(blockPos define.CubePos, botOffset define.CubePos, blockNEMCRuntimeID uint32, face int32, slotID uint8) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	if err = o.ensureHotBar(slotID); err != nil {
		return err
	}
	item, found := o.clientInfo.GetInventorySlot(0, slotID)
	if !found {
		return fmt.Errorf("slot is empty")
	}
	cubePos := protocol.BlockPos{int32(blockPos.X()), int32(blockPos.Y()), int32(blockPos.Z())}
	o.ctrl.SendPacket(&packet.PlayerAction{
		EntityRuntimeID: o.uq.GetBotRuntimeID(),
		ActionType:      protocol.PlayerActionStartItemUseOn,
		BlockPosition:   cubePos,
	})
	botPos := blockPos.Add(botOffset)
	o.ctrl.SendPacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
			Actions:            []protocol.InventoryAction(nil),
			ActionType:         protocol.UseItemActionClickBlock,
			BlockPosition:      cubePos,
			BlockFace:          face,
			HotBarSlot:         int32(slotID),
			HeldItem:           *item,
			BlockRuntimeID:     blockNEMCRuntimeID,
			Position:           mgl32.Vec3{float32(botPos.X()), float32(botPos.Y()), float32(botPos.Z())},
		},
	})
	o.ctrl.SendPacket(&packet.PlayerAction{
		EntityRuntimeID: o.uq.GetBotRuntimeID(),
		ActionType:      protocol.PlayerActionStopItemUseOn,
		BlockPosition:   cubePos,
	})
	return nil
}

func (o *AccessPointBotActionWithPersistData) tapBlockUsingHotBarItem(blockPos define.CubePos, blockNEMCRuntimeID uint32, slotID uint8) (err error) {
	if err = o.ensureHotBar(slotID); err != nil {
		return err
	}
	item, found := o.clientInfo.GetInventorySlot(0, slotID)
	if !found {
		return fmt.Errorf("slot is empty")
	}
	cubePos := protocol.BlockPos{int32(blockPos.X()), int32(blockPos.Y()), int32(blockPos.Z())}
	o.ctrl.SendPacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
			Actions:            []protocol.InventoryAction(nil),
			ActionType:         protocol.UseItemActionClickBlock,
			BlockPosition:      cubePos,
			HotBarSlot:         int32(slotID),
			HeldItem:           *item,
			BlockRuntimeID:     blockNEMCRuntimeID,
		},
	})
	o.ctrl.SendPacket(&packet.PlayerAction{
		EntityRuntimeID: o.uq.GetBotRuntimeID(),
		ActionType:      protocol.PlayerActionStartBuildingBlock,
		BlockPosition:   cubePos,
	})
	return nil
}

func (o *AccessPointBotActionWithPersistData) forceGetRidOfUnrecoverableState(hint string) {
	panic(fmt.Errorf("force get rid of unrecoverable state %v", hint))
}

func (o *AccessPointBotActionWithPersistData) makeEmptyItemInstance() *protocol.ItemInstance {
	return &protocol.ItemInstance{
		StackNetworkID: 0,
		Stack:          protocol.ItemStack{},
	}
}

func (o *AccessPointBotActionWithPersistData) copyItemInstance(instance *protocol.ItemInstance) *protocol.ItemInstance {
	newInstance := &protocol.ItemInstance{
		StackNetworkID: instance.StackNetworkID,
		Stack: protocol.ItemStack{
			ItemType:       instance.Stack.ItemType,
			BlockRuntimeID: instance.Stack.BlockRuntimeID,
			Count:          instance.Stack.Count,
			NBTData:        instance.Stack.NBTData,
			CanBePlacedOn:  instance.Stack.CanBePlacedOn,
			CanBreak:       instance.Stack.CanBreak,
			HasNetworkID:   instance.Stack.HasNetworkID,
		},
	}
	newInstance.Stack.NBTData = make(map[string]any)
	for _k, _v := range instance.Stack.NBTData {
		k, v := _k, _v
		newInstance.Stack.NBTData[k] = v
	}
	newInstance.Stack.CanBePlacedOn = make([]string, len(instance.Stack.CanBePlacedOn))
	for i, v := range instance.Stack.CanBePlacedOn {
		newInstance.Stack.CanBePlacedOn[i] = v
	}
	newInstance.Stack.CanBreak = make([]string, len(instance.Stack.CanBreak))
	for i, v := range instance.Stack.CanBreak {
		newInstance.Stack.CanBreak[i] = v
	}
	return newInstance
}

// 1. 玩家 Inventory(背包) 对应位置必须不为空
// 2. 被移动的位置必须为空
func (o *AccessPointBotActionWithPersistData) MoveItemFromInventoryToEmptyContainerSlots(pos define.CubePos, blockNemcRtid uint32, blockName string, moveOperations map[uint8]uint8) error {
	containerType, found := getContainerIDMappingByBlockBaseName(blockName)
	// fmt.Println(containerType)
	if !found {
		return fmt.Errorf("not a supported container")
	}
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()

	if _, opened := o.uq.GetCurrentOpenedContainer(); opened {
		return fmt.Errorf("another operation is already proceeding")
	}

	defer func() {
		o.currentContainerOpenListener = nil
		o.currentContainerCloseListener = nil
		o.currentItemStackResponseListener = nil
	}()

	openWaiter := make(chan *packet.ContainerOpen, 1)
	closeWaitor := make(chan struct{}, 1)
	itemResponseWaitor := make(chan *packet.ItemStackResponse, 1)

	o.currentContainerOpenListener = func(co *packet.ContainerOpen) { openWaiter <- co }
	o.currentContainerCloseListener = func(co *packet.ContainerClose) { closeWaitor <- struct{}{} }
	o.currentItemStackResponseListener = func(co *packet.ItemStackResponse) { itemResponseWaitor <- co }

	var container *packet.ContainerOpen
	o.tapBlockUsingHotBarItem(pos, blockNemcRtid, 0)
	select {
	case <-time.NewTimer(time.Second * 3).C:
		return fmt.Errorf("open container time out")
	case container = <-openWaiter:
		if container.ContainerPosition.X() != int32(pos.X()) || container.ContainerPosition.Y() != int32(pos.Y()) || container.ContainerPosition.Z() != int32(pos.Z()) {
			return fmt.Errorf("not this container opened")
		}
	}

	defer func() {
		container, found := o.uq.GetCurrentOpenedContainer()
		if found {
			o.ctrl.SendPacket(&packet.ContainerClose{
				WindowID:   container.WindowID,
				ServerSide: false,
			})
			select {
			case <-time.NewTimer(time.Second * 5).C:
				o.forceGetRidOfUnrecoverableState("fail to close container")
			case <-closeWaitor:
				o.SleepTick(1)
			}
		}
	}()

	containerWindow := container.WindowID
	var containerSlots *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]
	containerFound := false

	// wait until the information is updated
	for i := 0; i < 3*20; i++ {
		o.SleepTick(1)
		containerSlots, containerFound = o.clientInfo.GetInventoryWindow(uint32(containerWindow))
		if containerFound {
			break
		}
	}
	if !containerFound {
		return fmt.Errorf("not known information of target window")
	}

	if len(moveOperations) == 0 {
		return nil
	}

	requests := []protocol.ItemStackRequest{}
	containerNewData := map[uint8]*protocol.ItemInstance{}
	inventoryNewData := map[uint8]*protocol.ItemInstance{}

	for _inventorySlot, _containerSlot := range moveOperations {
		inventorySlot, containerSlot := _inventorySlot, _containerSlot
		inventoryItem, found := o.clientInfo.GetInventorySlot(0, inventorySlot)
		if !found || inventoryItem.StackNetworkID == 0 {
			return fmt.Errorf("item on specific inventory slot not found")
		}
		containerItem, found := containerSlots.Get(containerSlot)
		if !found {
			return fmt.Errorf("item on specific container slot not found")
		}
		placeStackRequestAction := &protocol.PlaceStackRequestAction{}
		placeStackRequestAction.Count = uint8(inventoryItem.Stack.Count)
		placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
			ContainerID:    ContainerIDInventory, // player inventory (bag)
			Slot:           inventorySlot,
			StackNetworkID: inventoryItem.StackNetworkID,
		}
		if inventoryItem.StackNetworkID == 0 {
			return fmt.Errorf("cannot move empty inventory slot %v to container slot %v", inventorySlot, containerSlot)
		}
		placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
			ContainerID:    containerType, // if is chest, should be 7
			Slot:           containerSlot,
			StackNetworkID: containerItem.StackNetworkID,
		}
		if containerItem.StackNetworkID != 0 {
			return fmt.Errorf("cannot move empty inventory slot %v to non-empty container slot %v", inventorySlot, containerSlot)
		}
		requestID := o.clientInfo.NextItemRequestID()
		containerNewData[containerSlot] = o.copyItemInstance(inventoryItem)
		inventoryNewData[inventorySlot] = o.copyItemInstance(containerItem)
		requests = append(requests, protocol.ItemStackRequest{
			RequestID: int32(requestID),
			Actions:   []protocol.StackRequestAction{placeStackRequestAction},
		})
	}
	// fmt.Println(packet.ItemStackRequest{
	// 	Requests: requests,
	// })
	o.ctrl.SendPacket(&packet.ItemStackRequest{
		Requests: requests,
	})

	// fmt.Println(inventoryNewData)
	// fmt.Println(containerNewData)
	select {
	case <-time.NewTimer(time.Second * 3).C:
		// fmt.Println("timeout in getting item stack response")
		return fmt.Errorf("timeout in getting item stack response")
	case resps := <-itemResponseWaitor:
		for _, response := range resps.Responses {
			if response.Status == protocol.ItemStackResponseStatusOK {
				containers := response.ContainerInfo
				for _, container := range containers {
					target := containerNewData
					windowID := int(containerWindow)
					if container.ContainerID == ContainerIDInventory {
						target = inventoryNewData
						windowID = 0
					}
					for _, slot := range container.SlotInfo {
						// fmt.Printf("window id: %v slot: %v networkID: %v\n", windowID, slot.Slot, slot.StackNetworkID)
						var itemInstance *protocol.ItemInstance
						if slot.StackNetworkID == 0 {
							itemInstance = o.makeEmptyItemInstance()
						} else {
							itemInstance = o.copyItemInstance(target[slot.Slot])
							itemInstance.StackNetworkID = slot.StackNetworkID
							itemInstance.Stack.Count = uint16(slot.Count)
						}
						o.clientInfo.writeInventorySlot(uint32(windowID), slot.Slot, itemInstance)
					}
				}
			} else {
				err = fmt.Errorf("sever report item stack request fail")
			}
		}
	}
	return err
}

func (o *AccessPointBotActionWithPersistData) UseAnvil(pos define.CubePos, slot uint8, newName string) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()

	if _, opened := o.uq.GetCurrentOpenedContainer(); opened {
		return fmt.Errorf("another operation is already proceeding")
	}

	defer func() {
		o.currentContainerOpenListener = nil
		o.currentContainerCloseListener = nil
		o.currentItemStackResponseListener = nil
	}()

	openWaiter := make(chan *packet.ContainerOpen, 1)
	closeWaitor := make(chan struct{}, 1)
	itemResponseWaitor := make(chan *packet.ItemStackResponse, 1)

	o.currentContainerOpenListener = func(co *packet.ContainerOpen) { openWaiter <- co }
	o.currentContainerCloseListener = func(co *packet.ContainerClose) { closeWaitor <- struct{}{} }
	o.currentItemStackResponseListener = func(co *packet.ItemStackResponse) { itemResponseWaitor <- co }

	// open anvil
	o.ctrl.SendPacket(&packet.PlayerAction{
		EntityRuntimeID: o.uq.GetBotRuntimeID(),
		ActionType:      protocol.PlayerActionStartBuildingBlock,
		BlockPosition:   protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
	})

	var container *packet.ContainerOpen
	select {
	case <-time.NewTimer(time.Second * 3).C:
		return fmt.Errorf("open anvil time out")
	case container = <-openWaiter:
		if container.ContainerPosition.X() != int32(pos.X()) || container.ContainerPosition.Y() != int32(pos.Y()) || container.ContainerPosition.Z() != int32(pos.Z()) {
			return fmt.Errorf("not this anvil opened")
		}
	}

	//close anvil
	defer func() {
		container, found := o.uq.GetCurrentOpenedContainer()
		if found {
			o.ctrl.SendPacket(&packet.ContainerClose{
				WindowID:   container.WindowID,
				ServerSide: false,
			})
			select {
			case <-time.NewTimer(time.Second * 5).C:
				o.forceGetRidOfUnrecoverableState("fail to close anvil")
			case <-closeWaitor:
				o.SleepTick(1)
			}
		}
	}()

	// wait anvil window open
	containerWindow := container.WindowID
	var containerSlots *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]
	containerFound := false

	// wait until the information is updated
	for i := 0; i < 3*20; i++ {
		o.SleepTick(1)
		containerSlots, containerFound = o.clientInfo.GetInventoryWindow(uint32(containerWindow))
		if containerFound {
			break
		}
	}
	if !containerFound {
		return fmt.Errorf("not known information of target window")
	}

	// get inventory
	inventoryItem, found := o.clientInfo.GetInventorySlot(0, slot)
	origInventoryItem := o.copyItemInstance(inventoryItem)
	if !found || inventoryItem.StackNetworkID == 0 {
		return fmt.Errorf("item on specific inventory slot not found")
	}

	// double check anvil status
	_, found = containerSlots.Get(0)
	if !found {
		return fmt.Errorf("specific container slot not found")
	}

	// construct request
	placeStackRequestAction := &protocol.PlaceStackRequestAction{}
	placeStackRequestAction.Count = uint8(inventoryItem.Stack.Count)
	placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    ContainerIDInventory, // player inventory (bag)
		Slot:           slot,
		StackNetworkID: inventoryItem.StackNetworkID,
	}
	if inventoryItem.StackNetworkID == 0 {
		return fmt.Errorf("cannot move empty inventory slot %v to anvil slot %v", slot, 0)
	}
	placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    0, // this is not a container
		Slot:           1,
		StackNetworkID: 0,
	}

	RequestIDPutItemOnAnvil := o.clientInfo.NextItemRequestID()
	RequestIDDoRenameAndTakeIt := o.clientInfo.NextItemRequestID()
	RequestIDGetBackItemWhenFail := o.clientInfo.NextItemRequestID()

	getStackRequestAction := &protocol.PlaceStackRequestAction{}
	getStackRequestAction.Count = uint8(inventoryItem.Stack.Count)
	getStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    ContainerIDInventory, // player inventory (bag)
		Slot:           slot,
		StackNetworkID: 0, // after put item on anvil, this should be 0
	}
	getStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    0x3c,
		Slot:           0x32,
		StackNetworkID: RequestIDDoRenameAndTakeIt,
	}

	getStackWhenFailRequestAction := &protocol.PlaceStackRequestAction{}
	getStackWhenFailRequestAction.Count = uint8(inventoryItem.Stack.Count)
	getStackWhenFailRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    ContainerIDInventory, // player inventory (bag)
		Slot:           slot,
		StackNetworkID: 0, // after put item on anvil, this should be 0
	}
	getStackWhenFailRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    0,
		Slot:           1,
		StackNetworkID: RequestIDPutItemOnAnvil,
	}

	o.ctrl.SendPacket(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			// put item on anvil
			{
				RequestID: RequestIDPutItemOnAnvil,
				Actions: []protocol.StackRequestAction{
					placeStackRequestAction,
				},
			},
			// rename
			{
				RequestID: RequestIDDoRenameAndTakeIt,
				Actions: []protocol.StackRequestAction{
					&protocol.CraftRecipeOptionalStackRequestAction{
						RecipeNetworkID:   0,
						FilterStringIndex: 0,
					},
					&protocol.ConsumeStackRequestAction{
						DestroyStackRequestAction: protocol.DestroyStackRequestAction{
							Count: byte(inventoryItem.Stack.Count),
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    0,
								Slot:           1,
								StackNetworkID: RequestIDPutItemOnAnvil,
							},
						},
					},
					getStackRequestAction,
				},
				FilterStrings: []string{newName},
			},
			{
				RequestID: RequestIDGetBackItemWhenFail,
				Actions: []protocol.StackRequestAction{
					getStackWhenFailRequestAction,
				},
			},
		},
	})

	// post-process and sync data
	select {
	case <-time.NewTimer(time.Second * 3).C:
		// fmt.Println("timeout in getting item stack response")
		return fmt.Errorf("timeout in getting item stack response")
	case resps := <-itemResponseWaitor:
		// for _, response := range resps.Responses {
		// 	bs, _ := json.Marshal(response)
		// 	fmt.Printf("%v\n", string(bs))
		// }

		// 如果不是3个response, 直接让程序崩溃，因为这种情况我们处理不了
		placeItemOnAnvilResponse := resps.Responses[0]
		renameAndGetItemResponse := resps.Responses[1]
		fallbackGetItemResponse := resps.Responses[2]
		if placeItemOnAnvilResponse.RequestID != RequestIDPutItemOnAnvil {
			o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
			o.SleepTick(1)
			return fmt.Errorf("client and server out of sync in maintained info (put item on anvil)")
			// o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info")
		}
		if renameAndGetItemResponse.RequestID != RequestIDDoRenameAndTakeIt {
			o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
			o.SleepTick(1)
			return fmt.Errorf("client and server out of sync in maintained info (take item from anvil)")
			// o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info")
		}

		// 如果命名成功, 这个 response 的 ID 可能变化
		// if fallbackGetItemResponse.RequestID != RequestIDGetBackItemWhenFail {
		// 	panic("client and server out of sync in maintained info")
		// }

		if renameAndGetItemResponse.Status == protocol.ItemStackResponseStatusOK {
			for _, container := range renameAndGetItemResponse.ContainerInfo {
				if container.ContainerID == ContainerIDInventory { // 玩家背包
					for _, slot := range container.SlotInfo {
						var itemInstance *protocol.ItemInstance
						if slot.StackNetworkID == 0 {
							o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
							o.SleepTick(1)
							return fmt.Errorf("client and server out of sync in maintained info (server report rename ok and get is successful, but get nothing)")
							// o.forceGetRidOfUnrecoverableState("server report rename ok and get is successful, but get nothing, out of sync!")
						} else {
							itemInstance = o.copyItemInstance(origInventoryItem)
							itemInstance.StackNetworkID = slot.StackNetworkID
							itemInstance.Stack.Count = uint16(slot.Count)
						}
						o.clientInfo.writeInventorySlot(0, slot.Slot, itemInstance)
						return nil
					}
				}
			}
		}

		// already fail
		// 现在检查是因为什么原因fail了，是第一步就没放上去还是命名失败
		if placeItemOnAnvilResponse.Status != protocol.ItemStackResponseStatusOK {
			// 第一步就出错了，数据没有更改，直接返回即可
			return fmt.Errorf("cannot finish rename, anvil is not empty")
		}

		// 因为命名失败导致第二步错误，现在检查是否正确拿到东西了
		if fallbackGetItemResponse.RequestID != RequestIDGetBackItemWhenFail {
			// o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info")
			o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
			o.SleepTick(1)
			return fmt.Errorf("client and server out of sync in maintained info (take back item from anvil)")
		}

		if fallbackGetItemResponse.Status == protocol.ItemStackResponseStatusOK {
			for _, container := range fallbackGetItemResponse.ContainerInfo {
				if container.ContainerID == ContainerIDInventory { // 玩家背包
					for _, slot := range container.SlotInfo {
						var itemInstance *protocol.ItemInstance
						if slot.StackNetworkID == 0 {
							// o.forceGetRidOfUnrecoverableState("server report rename fail and get is successful, but get nothing, out of sync!")
							o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
							o.SleepTick(1)
							return fmt.Errorf("client and server out of sync in maintained info (server report rename fail and get is successful, but get nothing)")
						} else {
							itemInstance = o.copyItemInstance(origInventoryItem)
							itemInstance.StackNetworkID = slot.StackNetworkID
							itemInstance.Stack.Count = uint16(slot.Count)
						}
						o.clientInfo.writeInventorySlot(0, slot.Slot, itemInstance)
						return fmt.Errorf("cannot finish rename, name is invalid")
					}
				}
			}
		} else {
			o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
			o.SleepTick(1)
			return fmt.Errorf("client and server out of sync in maintained info (server report rename fail and cannot get item back)")
			// o.forceGetRidOfUnrecoverableState("server report rename fail and cannot get item back, out of sync!")
		}
	}
	return nil
}

func (o *AccessPointBotActionWithPersistData) DropItemFromHotBar(slot uint8) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()

	defer func() {
		o.currentItemStackResponseListener = nil
	}()

	itemResponseWaitor := make(chan *packet.ItemStackResponse, 1)
	o.currentItemStackResponseListener = func(co *packet.ItemStackResponse) { itemResponseWaitor <- co }

	err = o.ensureHotBar(slot)
	if err != nil {
		return err
	}

	inventoryItem, found := o.clientInfo.GetInventorySlot(0, slot)
	if !found || inventoryItem.StackNetworkID == 0 {
		return fmt.Errorf("item on specific inventory slot not found")
	}

	dropItemRequestID := o.clientInfo.NextItemRequestID()
	o.ctrl.SendPacket(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: dropItemRequestID,
				Actions: []protocol.StackRequestAction{
					&protocol.DropStackRequestAction{
						Count: byte(inventoryItem.Stack.Count),
						Source: protocol.StackRequestSlotInfo{
							ContainerID:    ContainerIDHotBar,
							Slot:           byte(slot),
							StackNetworkID: inventoryItem.StackNetworkID,
						},
						Randomly: false,
					},
				},
			},
		},
	})
	// post-process and sync data
	select {
	case <-time.NewTimer(time.Second * 3).C:
		// fmt.Println("timeout in getting item stack response")
		return fmt.Errorf("timeout in getting item stack response")
	case resps := <-itemResponseWaitor:
		dropItemResponse := resps.Responses[0]
		if dropItemResponse.RequestID != dropItemRequestID {
			// o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info")
			o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
			o.SleepTick(1)
			return fmt.Errorf("client and server out of sync in maintained info (drop item)")
		}
		for _, container := range dropItemResponse.ContainerInfo {
			if container.ContainerID == ContainerIDHotBar { // 玩家快捷物品栏
				for _, slot := range container.SlotInfo {
					if slot.StackNetworkID == 0 {
						o.clientInfo.writeInventorySlot(0, slot.Slot, o.makeEmptyItemInstance())
					} else {
						// 明明应该丢出物品的，却告诉背包里有新东西，我们肯定无法知道这个新东西是什么
						// o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info, want drop item, but get item")
						o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("replaceitem entity @s slot.hotbar %v air", slot)).SetTimeout(time.Second * 3).BlockGetResult()
						o.SleepTick(1)
						return fmt.Errorf("client and server out of sync in maintained info (want drop item, but get item)")
					}
				}
			}
		}
	}
	return nil
}

func (o *AccessPointBotActionWithPersistData) MoveItemInsideHotBarOrInventory(sourceSlot, targetSlot uint8, count uint8) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()

	defer func() {
		o.currentContainerOpenListener = nil
		o.currentContainerCloseListener = nil
		o.currentItemStackResponseListener = nil
	}()

	openWaiter := make(chan *packet.ContainerOpen, 1)
	closeWaitor := make(chan struct{}, 1)
	itemResponseWaitor := make(chan *packet.ItemStackResponse, 1)

	o.currentContainerOpenListener = func(co *packet.ContainerOpen) { openWaiter <- co }
	o.currentContainerCloseListener = func(co *packet.ContainerClose) { closeWaitor <- struct{}{} }
	o.currentItemStackResponseListener = func(co *packet.ItemStackResponse) { itemResponseWaitor <- co }

	o.ctrl.SendPacket(&packet.Interact{
		ActionType:            packet.InteractActionOpenInventory,
		TargetEntityRuntimeID: o.uq.GetBotRuntimeID(),
	})

	// var container *packet.ContainerOpen
	select {
	case <-time.NewTimer(time.Second * 3).C:
		return fmt.Errorf("open inventory (bot bag) time out")
	case <-openWaiter:
		// fmt.Println(container)
		// WindowID:2
		// ContainerIDType 255
		// Pos: bot pos
		// EntityRuntimeID -1
	}

	defer func() {
		container, found := o.uq.GetCurrentOpenedContainer()
		if found {
			o.ctrl.SendPacket(&packet.ContainerClose{
				WindowID:   container.WindowID,
				ServerSide: false,
			})
			select {
			case <-time.NewTimer(time.Second * 5).C:
				o.forceGetRidOfUnrecoverableState("fail to close inventory")
			case <-closeWaitor:
				o.SleepTick(1)
			}
		}
	}()

	RequestIDMoveItem := o.clientInfo.NextItemRequestID()

	inventoryItem, found := o.clientInfo.GetInventorySlot(0, sourceSlot)
	if !found || inventoryItem.StackNetworkID == 0 {
		return fmt.Errorf("item on specific inventory slot not found")
	}
	targetSlotItem, found := o.clientInfo.GetInventorySlot(0, targetSlot)
	if !found || targetSlotItem.StackNetworkID != 0 {
		return fmt.Errorf("specific target slot is not empty")
	}
	moveStackRequestAction := &protocol.PlaceStackRequestAction{}
	if count > uint8(inventoryItem.Stack.Count) {
		count = uint8(inventoryItem.Stack.Count)
	}
	moveStackRequestAction.Count = count
	moveStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    ContainerIDInventory, // player inventory (bag)
		Slot:           sourceSlot,
		StackNetworkID: inventoryItem.StackNetworkID,
	}
	moveStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    ContainerIDInventory, // player inventory (bag)
		Slot:           targetSlot,
		StackNetworkID: 0,
	}
	o.ctrl.SendPacket(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: RequestIDMoveItem,
				Actions: []protocol.StackRequestAction{
					moveStackRequestAction,
				},
			},
		},
	})
	// post-process and sync data
	select {
	case <-time.NewTimer(time.Second * 3).C:
		// fmt.Println("timeout in getting item stack response")
		return fmt.Errorf("timeout in getting item stack response")
	case resps := <-itemResponseWaitor:
		moveItemResponse := resps.Responses[0]
		if moveItemResponse.RequestID != RequestIDMoveItem {
			o.forceGetRidOfUnrecoverableState("client and server out of sync in maintained info")
		}
		// 必然是一个响应，如果不是一个，我们也处理不了这个错误
		response := resps.Responses[0]
		if response.Status == protocol.ItemStackResponseStatusOK {
			origItem := o.copyItemInstance(inventoryItem)
			for _, container := range response.ContainerInfo {
				if container.ContainerID != ContainerIDInventory {
					o.forceGetRidOfUnrecoverableState("operate inventory but server report non-inventory container")
				}
				for _, slot := range container.SlotInfo {
					if slot.StackNetworkID == 0 {
						o.clientInfo.writeInventorySlot(0, slot.Slot, o.makeEmptyItemInstance())
					} else {
						newItem := o.copyItemInstance(origItem)
						newItem.Stack.Count = uint16(slot.Count)
						newItem.StackNetworkID = slot.StackNetworkID
						o.clientInfo.writeInventorySlot(0, slot.Slot, newItem)
					}
				}
			}
			return nil
		} else {
			return fmt.Errorf("fail to move item inside player inventory or hotbar")
		}
	}
}
