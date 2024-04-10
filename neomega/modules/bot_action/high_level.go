package bot_action

import (
	"context"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/blocks"
	"neo-omega-kernel/neomega/chunks"
	"neo-omega-kernel/neomega/chunks/define"
	"neo-omega-kernel/nodes"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type BotActionHighLevel struct {
	uq                 neomega.MicroUQHolder
	ctrl               neomega.InteractCore
	cmdSender          neomega.CmdSender
	cmdHelper          neomega.CommandHelper
	structureRequester neomega.StructureRequester
	microAction        neomega.BotAction
	pickedItemChan     chan protocol.InventoryAction
	muChan             chan struct{}
	node               nodes.Node
	// asyncNBTBlockPlacer neomega.AsyncNBTBlockPlacer
}

func NewBotActionHighLevel(
	uq neomega.MicroUQHolder, ctrl neomega.InteractCore, react neomega.ReactCore, cmdSender neomega.CmdSender, cmdHelper neomega.CommandHelper, structureRequester neomega.StructureRequester, microAction neomega.BotAction,
	node nodes.Node,
) neomega.BotActionHighLevel {
	muChan := make(chan struct{}, 1)
	muChan <- struct{}{}
	bah := &BotActionHighLevel{
		uq:                 uq,
		ctrl:               ctrl,
		cmdSender:          cmdSender,
		cmdHelper:          cmdHelper,
		structureRequester: structureRequester,
		// asyncNBTBlockPlacer: asyncNBTBlockPlacer,
		microAction: microAction,
		muChan:      muChan,
		node:        node,
	}

	react.SetTypedPacketCallBack(packet.IDInventoryTransaction, func(p packet.Packet) {
		isDroppedItem := false
		pk := p.(*packet.InventoryTransaction)
		for _, _value := range pk.Actions {
			c := bah.pickedItemChan
			if c == nil {
				continue
			}
			// bs, _ := json.Marshal(_value)
			// fmt.Println(string(bs))
			// always slot 1, StackNetworkID 0, when using "give" command
			// protocol.InventoryActionSourceCreative
			// always slot 1, StackNetworkID 0, when item is picked from the world
			// protocol.InventoryActionSourceWorld

			if _value.SourceType == protocol.InventoryActionSourceWorld {
				isDroppedItem = true
			} else if _value.SourceType == protocol.InventoryActionSourceContainer && isDroppedItem {
				value := _value
				select {
				case c <- value:
				default:
				}
			}
		}
	}, false)

	return bah
}

func (o *BotActionHighLevel) occupyBot(timeout time.Duration) (release func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		o.microAction.SleepTick(1)
		return nil, fmt.Errorf("cannot acquire bot (high level)")
	case <-o.muChan:
		if !o.node.TryLock("bot-action-high", time.Second*2) {
			o.muChan <- struct{}{} // give back bot control
			return nil, fmt.Errorf("cannot acquire bot (high level, distribute)")
		}
		stopChan := make(chan struct{})
		go func() {
			for {
				select {
				case <-stopChan:
					return
				case <-time.NewTimer(time.Second).C:
					o.node.ResetLockTime("bot-action-high", time.Second*2)
				}
			}
		}()
		return func() {
			close(stopChan)
			o.node.ResetLockTime("bot-action-high", 0)
			o.microAction.SleepTick(1)
			o.muChan <- struct{}{} // give back bot control
		}, nil
	}
}

func (o *BotActionHighLevel) highLevelEnsureBotNearby(pos define.CubePos, threshold float32) error {
	botPos, _ := o.uq.GetBotPosition()
	if botPos.Sub(mgl32.Vec3{float32(pos.X()), float32(pos.Y()), float32(pos.Z())}).Len() > threshold {
		ret := o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z())).SetTimeout(time.Second * 3).BlockGetResult()
		if ret == nil {
			return fmt.Errorf("cannot move to target pos")
		}
	}
	return nil
}

func (o *BotActionHighLevel) HighLevelEnsureBotNearby(pos define.CubePos, threshold float32) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelEnsureBotNearby(pos, threshold)
}

// if we want the pos to be air when use it, recoverFromAir=true
// e.g. 如果我们希望把某个方块位置临时变为 air, 则 wantAir=true
// 如果希望把某个方块位置临时变为某个非空气方块, 则 wantAir=false
func (o *BotActionHighLevel) HighLevelRemoveSpecificBlockSideEffect(pos define.CubePos, wantAir bool, backupName string) (deferFunc func(), err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return func() {}, err
	}
	defer release()
	return o.highLevelRemoveSpecificBlockSideEffect(pos, wantAir, backupName)
}

func (o *BotActionHighLevel) highLevelRemoveSpecificBlockSideEffect(pos define.CubePos, wantAir bool, backupName string) (deferFunc func(), err error) {
	_, deferFunc, err = o.highLevelGetAndRemoveSpecificBlockSideEffect(pos, wantAir, backupName)
	return deferFunc, err
}

func (o *BotActionHighLevel) highLevelGetAndRemoveSpecificBlockSideEffect(pos define.CubePos, wantAir bool, backupName string) (decodedStructure *neomega.DecodedStructure, deferFunc func(), err error) {
	o.highLevelEnsureBotNearby(pos, 8)
	structure, err := o.structureRequester.RequestStructure(pos, define.CubePos{1, 1, 1}, backupName).SetTimeout(time.Second * 3).BlockGetResult()
	if err != nil {
		return nil, func() {}, err
	}
	decodedStructure, err = structure.Decode()
	if err != nil {
		return nil, func() {}, err
	}
	foreGround, backGround := decodedStructure.BlockOf(define.CubePos{0, 0, 0})
	isAir := false
	if foreGround == blocks.AIR_RUNTIMEID && backGround == blocks.AIR_RUNTIMEID {
		isAir = true
	} else {
		ret := o.cmdHelper.BackupStructureWithGivenNameCmd(pos, define.CubePos{1, 1, 1}, backupName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		if ret == nil {
			return nil, func() {}, fmt.Errorf("cannot backup block for revert")
		}
	}
	deferFunc = func() {
		o.cmdHelper.RevertStructureWithGivenNameCmd(pos, backupName).Send()
		o.microAction.SleepTick(1)
	}
	if !wantAir {
		if isAir {
			deferFunc = func() { o.cmdHelper.SetBlockCmd(pos, "air").Send() }
		}
	} else {
		if isAir {
			deferFunc = func() {}
		}
		o.cmdHelper.SetBlockCmd(pos, "air").AsWebSocket().SendAndGetResponse().BlockGetResult()
	}
	return decodedStructure, deferFunc, nil
}

func (o *BotActionHighLevel) HighLevelPlaceSign(targetPos define.CubePos, text string, lighting bool, signBlock string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceSign(targetPos, text, lighting, signBlock)
}

func (o *BotActionHighLevel) highLevelPlaceSign(targetPos define.CubePos, text string, lighting bool, signBlock string) (err error) {
	victimBlock := targetPos.Add(define.CubePos{1, 0, 0})
	if err = o.highLevelEnsureBotNearby(targetPos, 8); err != nil {
		return err
	}
	ret := o.cmdHelper.ReplaceHotBarItemCmd(0, "oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	if ret == nil {
		return fmt.Errorf("cannot move to target pos")
	}
	deferredAction, err := o.highLevelRemoveSpecificBlockSideEffect(victimBlock, false, "_sign_temp")
	defer deferredAction()

	o.cmdHelper.SetBlockCmd(targetPos, "air").Send()
	o.cmdHelper.SetBlockCmd(victimBlock, "glass").Send()
	blockRTID, found := blocks.BlockStrToRuntimeID("glass")
	if !found {
		return fmt.Errorf("glass runtime id not found")
	}
	// nemcBlockRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(blockRTID)
	// if nemcBlockRuntimeID == chunk.NEMCAirRID {
	// 	return fmt.Errorf("glass nemc runtime id not found")
	// }
	o.microAction.SleepTick(4)
	// o.microAction.SelectHotBar(0)
	o.microAction.UseHotBarItemOnBlock(victimBlock, blockRTID, 4, 0)

	// this is actually a bug of minecraft
	// if old sign is replace by new sign when "writing" (actually is "unfinished" state since nothing is sent to server when writing)
	// data will be put in new sign
	o.cmdHelper.SetBlockCmd(targetPos, signBlock).Send()
	o.microAction.SleepTick(1)
	IgnoreLighting := uint8(0)
	TextIgnoreLegacyBugResolved := uint8(0)
	if lighting {
		IgnoreLighting = uint8(1)
		TextIgnoreLegacyBugResolved = uint8(1)
	}
	o.ctrl.SendPacket(&packet.BlockActorData{
		Position: protocol.BlockPos{int32(targetPos.X()), int32(targetPos.Y()), int32(targetPos.Z())},
		NBTData: map[string]any{
			"x":                           int32(targetPos.X()),
			"y":                           int32(targetPos.Y()),
			"z":                           int32(targetPos.Z()),
			"id":                          "Sign",
			"IgnoreLighting":              IgnoreLighting,
			"PersistFormatting":           uint8(1),
			"TextOwner":                   "",
			"isMovable":                   uint8(1),
			"TextIgnoreLegacyBugResolved": TextIgnoreLegacyBugResolved,
			"SignTextColor":               int32(-16777216),
			"Text":                        text,
		},
	})
	return nil
}

func (o *BotActionHighLevel) HighLevelPlaceCommandBlock(option *neomega.PlaceCommandBlockOption, maxRetry int) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceCommandBlock(option, maxRetry)
}

func (o *BotActionHighLevel) highLevelPlaceCommandBlock(option *neomega.PlaceCommandBlockOption, maxRetry int) error {
	if err := o.highLevelEnsureBotNearby(define.CubePos{option.X, option.Y, option.Z}, 8); err != nil {
		return err
	}
	updateOption := option.GenCommandBlockUpdateFromOption()
	sleepTime := 1
	for maxRetry > 0 {
		maxRetry--
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", option.X, option.Y, option.Z, strings.Replace(option.BlockName, "minecraft:", "", 1), option.BlockState)
		o.cmdSender.SendWebSocketCmdNeedResponse(cmd).SetTimeout(time.Second * 3).BlockGetResult()
		o.ctrl.SendPacket(updateOption)
		time.Sleep(100 * time.Millisecond)
		r, err := o.structureRequester.RequestStructure(define.CubePos{option.X, option.Y, option.Z}, define.CubePos{1, 1, 1}, "_temp").BlockGetResult()
		if err != nil {
		} else {
			d, err := r.Decode()
			if err == nil {
				if len(d.Nbts) > 0 {
					for _, tnbt := range d.Nbts {
						ok := true
						if tnbt["id"].(string) != "CommandBlock" {
							ok = false
						} else if tnbt["Command"].(string) != option.Command {
							ok = false
						} else if tnbt["CustomName"].(string) != option.Name {
							ok = false
						}
						if ok {
							return nil
						}
						break
					}
				}
			}
		}
		sleepTime++
	}
	return fmt.Errorf("cannot successfully place commandblock in given limit")
}

func (o *BotActionHighLevel) HighLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelMoveItemToContainer(pos, moveOperations)
}

func (o *BotActionHighLevel) highLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error {
	if err := o.highLevelEnsureBotNearby(pos, 8); err != nil {
		return err
	}
	structureResponse, err := o.structureRequester.RequestStructure(pos, define.CubePos{1, 1, 1}, "_temp").BlockGetResult()
	if err != nil {
		return err
	}
	structure, err := structureResponse.Decode()
	if err != nil {
		return err
	}
	containerRuntimeID := structure.ForeGround[0]
	// containerNEMCRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(containerRuntimeID)
	if containerRuntimeID == blocks.AIR_RUNTIMEID {
		return fmt.Errorf("block of %v (nemc) not found", containerRuntimeID)
	}
	block, found := blocks.RuntimeIDToBlock(containerRuntimeID)
	if !found {
		panic(fmt.Errorf("block of %v not found", containerRuntimeID))
	}
	_, found = getContainerIDMappingByBlockBaseName(block.Name)
	if !found {
		return fmt.Errorf("block %v is not a supported container", block.Name)
	}
	for _, targetSlot := range moveOperations {
		o.cmdHelper.ReplaceContainerBlockItemCmd(pos, int32(targetSlot), "air").Send()
	}
	deferAction := func() {}
	if strings.Contains(block.Name, "shulker_box") {
		blockerPos := pos
		face := byte(255)
		if len(structure.Nbts[pos]) > 0 {
			if facing_origin, ok := structure.Nbts[pos]["facing"]; ok {
				face, ok = facing_origin.(byte)
				if !ok {
					face = 255
				}
			}
		}
		if face != 255 {
			switch face {
			case 0:
				blockerPos[1] = blockerPos[1] - 1
			case 1:
				blockerPos[1] = blockerPos[1] + 1
			case 2:
				blockerPos[2] = blockerPos[2] - 1
			case 3:
				blockerPos[2] = blockerPos[2] + 1
			case 4:
				blockerPos[0] = blockerPos[0] - 1
			case 5:
				blockerPos[0] = blockerPos[0] + 1
			}
			deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(blockerPos, true, "_temp_container_blocker")
			if err != nil {
				return err
			}
		}
	} else if strings.Contains(block.Name, "chest") {
		o.cmdHelper.BackupStructureWithGivenNameCmd(pos.Add(define.CubePos{0, 1, 0}), define.CubePos{1, 1, 1}, "container_blocker").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(pos.Add(define.CubePos{0, 1, 0}), true, "_temp_container_blocker")
		if err != nil {
			return err
		}
	}
	defer deferAction()
	o.microAction.SleepTick(1)
	return o.microAction.MoveItemFromInventoryToEmptyContainerSlots(pos, containerRuntimeID, block.Name, moveOperations)
}

func (o *BotActionHighLevel) HighLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelRenameItemWithAnvil(pos, slot, newName, autoGenAnvil)
}

func (o *BotActionHighLevel) highLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error) {
	if err := o.highLevelEnsureBotNearby(pos, 8); err != nil {
		return err
	}
	deferActionStand := func() {}
	deferAction := func() {}
	if autoGenAnvil {
		deferActionStand, err = o.highLevelRemoveSpecificBlockSideEffect(pos.Add(define.CubePos{0, -1, 0}), false, "_temp_anvil_stand")
		if err != nil {
			return err
		}
		if ret := o.cmdHelper.SetBlockCmd(pos.Add(define.CubePos{0, -1, 0}), "glass").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil {
			return fmt.Errorf("cannot place anvil for operation")
		}
		deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(pos, false, "_temp_anvil")
		if err != nil {
			return err
		}
		if ret := o.cmdHelper.SetBlockCmd(pos, "anvil").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil {
			return fmt.Errorf("cannot place anvil for operation")
		}
	}
	defer deferActionStand()
	defer deferAction()
	o.microAction.SleepTick(1)
	return o.microAction.UseAnvil(pos, slot, newName)
}

func (o *BotActionHighLevel) HighLevelEnchantItem(slot uint8, enchants map[string]int32) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelEnchantItem(slot, enchants)
}

func (o *BotActionHighLevel) highLevelEnchantItem(slot uint8, enchants map[string]int32) (err error) {
	o.microAction.SelectHotBar(slot)
	o.microAction.SleepTick(1)
	results := make(chan *packet.CommandOutput, len(enchants))
	for nameOrID, level := range enchants {
		o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("enchant @s %v %v", nameOrID, level)).SetTimeout(time.Second * 3).AsyncGetResult(func(output *packet.CommandOutput) {
			results <- output
		})
	}
	for i := 0; i < len(enchants); i++ {
		r := <-results
		if r == nil || r.SuccessCount == 0 {
			err = fmt.Errorf("some enchant command fail")
		}
	}
	return
}

func (o *BotActionHighLevel) highLevelListenItemPicked(ctx context.Context) (actionChan chan protocol.InventoryAction, err error) {
	go func() {
		<-ctx.Done()
		o.pickedItemChan = nil
	}()
	o.pickedItemChan = make(chan protocol.InventoryAction, 64)
	return o.pickedItemChan, nil
}

func (o *BotActionHighLevel) HighLevelListenItemPicked(timeout time.Duration) (actionChan chan protocol.InventoryAction, cancel func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return nil, func() {}, err
	}
	go func() {
		<-ctx.Done()
		o.pickedItemChan = nil
		release()
	}()
	o.pickedItemChan = make(chan protocol.InventoryAction, 64)
	return o.pickedItemChan, cancel, nil
}

func (o *BotActionHighLevel) HighLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlots map[uint8]bool, maxRetriesTotal int) (targetSlotsGetInfo map[uint8]bool, err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return targetSlots, err
	}
	defer release()
	return o.highLevelBlockBreakAndPickInHotBar(pos, recoverBlock, targetSlots, maxRetriesTotal)
}

func (o *BotActionHighLevel) highLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlots map[uint8]bool, maxRetriesTotal int) (targetSlotsGetInfo map[uint8]bool, err error) {
	// split targets
	targetSlotsGetInfo = map[uint8]bool{}
	for k, v := range targetSlots {
		if !v {
			if k > 8 {
				return targetSlots, fmt.Errorf("hot bar slot can only be 0~8")
			}
			targetSlotsGetInfo[k] = v
			o.cmdHelper.ReplaceHotBarItemCmd(int32(k), "air").SendAndGetResponse().BlockGetResult()
		}
	}
	o.microAction.SleepTick(1)
	// merge later
	defer func() {
		for k, v := range targetSlots {
			if v {
				targetSlotsGetInfo[k] = v
			}
		}
	}()
	if len(targetSlotsGetInfo) == 0 {
		return
	}
	if o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z())).SetTimeout(time.Second*3).BlockGetResult() == nil {
		return targetSlotsGetInfo, fmt.Errorf("cannot make bot to target position")
	}
	if cleanResult := o.cmdSender.SendWebSocketCmdNeedResponse("tp @e[type=item,r=9] ~ -100 ~").SetTimeout(time.Second * 3).BlockGetResult(); cleanResult == nil {
		return targetSlotsGetInfo, fmt.Errorf("cannot clean bot nearby items")
	}

	ctx, cancelListen := context.WithCancel(context.Background())
	defer cancelListen()
	actionChan, err := o.highLevelListenItemPicked(ctx)
	if err != nil {
		return targetSlotsGetInfo, err
	}
	currentBlock, recoverAction, err := o.highLevelGetAndRemoveSpecificBlockSideEffect(pos, false, "_temp_break")
	if err != nil {
		return targetSlotsGetInfo, err
	}
	if currentBlock.ForeGround[0] == blocks.AIR_RUNTIMEID && currentBlock.BackGround[0] == blocks.AIR_RUNTIMEID {
		return targetSlotsGetInfo, fmt.Errorf("block is air")
	}
	defer func() {
		if err != nil || recoverBlock {
			recoverAction()
		}
	}()
	totalTimes := len(targetSlotsGetInfo) + maxRetriesTotal
	for tryTime := 0; tryTime < totalTimes; tryTime++ {
		thisTimeOk := false
		pickedSlot := -1
		// move and break immediately
		o.cmdSender.SendWOCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
		o.cmdSender.SendWOCmd(fmt.Sprintf("setblock %v %v %v air 0 destroy", pos.X(), pos.Y(), pos.Z()))
		// listen block picked (wait at most 3s)
		select {
		case <-time.NewTicker(time.Second * 1).C:
			// get no item
		case pickedAction := <-actionChan:
			// item acquired
			pickedSlot = int(pickedAction.InventorySlot)
			// fmt.Println(pickedSlot)
			// we need to check if anything not wanted also dropped into inventory
			o.microAction.SleepTick(5)
			// eat up all unwanted actions
			hasInreleventItem := false
			for {
				notWantedExist := false
				select {
				case <-actionChan:
					notWantedExist = true
					hasInreleventItem = true
				default:
				}
				if !notWantedExist {
					break
				}
			}
			if !hasInreleventItem {
				thisTimeOk = true
			} else {
				return targetSlotsGetInfo, fmt.Errorf("this is not a simple block or a shulker box, but a block with contents, which cannot be get in single slot")
			}
		}

		if thisTimeOk {
			// check if item is in slot we want
			if _, ok := targetSlotsGetInfo[uint8(pickedSlot)]; ok {
				// lucky
				// fmt.Println("get in slot ", pickedSlot)
				targetSlotsGetInfo[uint8(pickedSlot)] = true
			} else {
				// move item
				targetSlot := -1
				for k, v := range targetSlotsGetInfo {
					if !v {
						targetSlot = int(k)
					}
				}
				if targetSlot == -1 {
					panic("programme logic error, should never reach this")
				}
				// fmt.Printf("move %v -> %v\n", pickedSlot, targetSlot)
				if err = o.microAction.MoveItemInsideHotBarOrInventory(uint8(pickedSlot), uint8(targetSlot), 1); err != nil {
					// maybe something block this slot
					o.cmdHelper.ReplaceHotBarItemCmd(int32(targetSlot), "air").SendAndGetResponse().SetTimeout(time.Second).BlockGetResult()
					o.microAction.SleepTick(1)
					if err = o.microAction.MoveItemInsideHotBarOrInventory(uint8(pickedSlot), uint8(targetSlot), 1); err != nil {
						// oh no
						thisTimeOk = false
					} else {
						targetSlotsGetInfo[uint8(targetSlot)] = true
						pickedSlot = targetSlot
					}
				} else {
					targetSlotsGetInfo[uint8(targetSlot)] = true
					pickedSlot = targetSlot
				}
			}
		}

		// fmt.Printf("do time: %v ok: %v slot: %v\n", tryTime, thisTimeOk, pickedSlot)
		allDone := true
		for _, v := range targetSlotsGetInfo {
			if !v {
				allDone = false
				break
			}
		}
		if allDone {
			return targetSlotsGetInfo, nil
		}
		if tryTime == totalTimes-1 {
			break
		}
		recoverAction()
		if !thisTimeOk {
			if cleanResult := o.cmdSender.SendWebSocketCmdNeedResponse("tp @e[type=item,r=9] ~ -100 ~").SetTimeout(time.Second * 3).BlockGetResult(); cleanResult == nil {
				return targetSlotsGetInfo, fmt.Errorf("cannot clean bot nearby items")
			}
			for k, v := range targetSlotsGetInfo {
				if !v {
					o.cmdHelper.ReplaceHotBarItemCmd(int32(k), "air").SendAndGetResponse().BlockGetResult()
				}
			}
		}
		o.microAction.SleepTick(10)
	}
	o.microAction.SleepTick(5)
	return targetSlotsGetInfo, fmt.Errorf("not all slots successfully get block")
}

func (o *BotActionHighLevel) HighLevelWriteBook(slotID uint8, pages []string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelWriteBook(slotID, pages)
}

func (o *BotActionHighLevel) highLevelWriteBook(slotID uint8, pages []string) (err error) {
	rest := o.cmdHelper.ReplaceHotBarItemCmd(int32(slotID), "writable_book").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	if rest == nil {
		return fmt.Errorf("cannot get writable_book in slot")
	}
	o.microAction.SleepTick(1)
	o.microAction.UseHotBarItem(slotID)
	o.microAction.SleepTick(1)
	for page, data := range pages {
		o.ctrl.SendPacket(&packet.BookEdit{
			ActionType:    packet.BookActionReplacePage,
			InventorySlot: slotID,
			Text:          data,
			PageNumber:    uint8(page),
		})
	}
	o.microAction.SleepTick(1)
	return nil
}

func (o *BotActionHighLevel) HighLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	// its wired, we must do this or bot cannot generate book
	o.microAction.SelectHotBar(0)
	return o.highLevelWriteBookAndClose(slotID, pages, bookTitle, bookAuthor)
}

func (o *BotActionHighLevel) highLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error) {
	if err := o.highLevelWriteBook(slotID, pages); err != nil {
		return err
	}
	o.ctrl.SendPacket(&packet.BookEdit{
		ActionType:    packet.BookActionSign,
		InventorySlot: slotID,
		Title:         bookTitle,
		Author:        bookAuthor,
	})
	return nil
}

func (o *BotActionHighLevel) HighLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceItemFrameItem(pos, slotID)
}

func (o *BotActionHighLevel) highLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8) (err error) {
	o.highLevelEnsureBotNearby(pos, 0)
	block, err := o.structureRequester.RequestStructure(pos, define.CubePos{1, 1, 1}, "_t").SetTimeout(time.Second * 3).BlockGetResult()
	if err != nil {
		return err
	}
	decoded, err := block.Decode()
	if err != nil {
		panic(err)
	}
	runtimeID := decoded.ForeGround[0]
	// nemcRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(runtimeID)
	_, states, _ := blocks.RuntimeIDToState(decoded.ForeGround[0])
	face, ok := states["facing_direction"]
	if !ok {
		return fmt.Errorf("facing not found")
	}
	facing, ok := face.(int32)
	if !ok {
		return fmt.Errorf("facing not found")
	}
	o.microAction.SleepTick(1)
	err = o.microAction.UseHotBarItemOnBlock(pos, runtimeID, facing, slotID)
	o.microAction.SleepTick(1)
	return err
}

func (o *BotActionHighLevel) HighLevelSetContainerContent(pos define.CubePos, containerInfo map[uint8]*neomega.ContainerSlotItemStack) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	o.highLevelEnsureBotNearby(pos, 8)
	return o.highLevelSetContainerItems(pos, containerInfo)
}

func (o *BotActionHighLevel) HighLevelGenContainer(pos define.CubePos, containerInfo map[uint8]*neomega.ContainerSlotItemStack, block string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	o.highLevelEnsureBotNearby(pos, 8)
	if ret := o.cmdHelper.SetBlockCmd(pos, block).AsWebSocket().SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult(); ret == nil {
		return fmt.Errorf("cannot set container")
	}
	return o.highLevelSetContainerItems(pos, containerInfo)
}

func (o *BotActionHighLevel) HighLevelMakeItem(item *neomega.Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelMakeItem(item, slotID, anvilPos, nextContainerPos)
}

func (o *BotActionHighLevel) highLevelMakeItem(item *neomega.Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error {
	typeDescription := item.GetTypeDescription()
	if !typeDescription.IsComplexBlock() {
		if typeDescription.KnownItem() == neomega.KnownItemWritableBook {
			if err := o.highLevelWriteBook(uint8(slotID), item.RelatedKnownItemData.Pages); err != nil {
				return err
			}
		} else if typeDescription.KnownItem() == neomega.KnownItemWrittenBook {
			if err := o.highLevelWriteBookAndClose(uint8(slotID), item.RelatedKnownItemData.Pages, item.RelatedKnownItemData.BookName, item.RelatedKnownItemData.BookAuthor); err != nil {
				return err
			}
		} else {
			if ret := o.cmdHelper.ReplaceBotHotBarItemFullCmd(int32(slotID), item.Name, 1, int32(item.Value), item.Components).SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil {
				return fmt.Errorf("cannot put simple block/item in container %v %v %v %v", item.Name, 1, int32(item.Value), item.Components)
			}
		}
		if item.DisplayName != "" {
			deferActionStand, _ := o.highLevelRemoveSpecificBlockSideEffect(anvilPos.Add(define.CubePos{0, -1, 0}), false, "_temp_anvil_stand")
			o.cmdHelper.SetBlockCmd(anvilPos.Add(define.CubePos{0, -1, 0}), "glass").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
			deferAction, _ := o.highLevelRemoveSpecificBlockSideEffect(anvilPos, false, "_temp_anvil")
			o.cmdHelper.SetBlockCmd(anvilPos, "anvil").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
			o.highLevelRenameItemWithAnvil(anvilPos, slotID, item.DisplayName, false)
			deferAction()
			deferActionStand()
		}
		if len(item.Enchants) > 0 {
			if err := o.highLevelEnchantItem(slotID, item.Enchants); err != nil {
				return err
			}
		}
	} else {
		deferActionWorkspace, _ := o.highLevelRemoveSpecificBlockSideEffect(nextContainerPos, false, "_temp_work")
		defer deferActionWorkspace()
		o.cmdHelper.SetBlockCmd(nextContainerPos, fmt.Sprintf("%v %v", item.Name, item.RelatedBlockBedrockStateString)).AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
		if err := o.highLevelSetContainerItems(nextContainerPos, item.RelateComplexBlockData.Container); err != nil {
			return err
		}
		if _, err := o.highLevelBlockBreakAndPickInHotBar(nextContainerPos, false, map[uint8]bool{slotID: false}, 2); err != nil {
			return err
		}
		// give complex block enchant and name
		if len(item.Enchants) > 0 {
			if err := o.highLevelEnchantItem(slotID, item.Enchants); err != nil {
				return err
			}
		}
		if item.DisplayName != "" {
			if err := o.highLevelRenameItemWithAnvil(anvilPos, slotID, item.DisplayName, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *BotActionHighLevel) highLevelSetContainerItems(pos define.CubePos, containerInfo map[uint8]*neomega.ContainerSlotItemStack) (err error) {
	o.highLevelEnsureBotNearby(pos, 8)
	updateErr := func(newErr error) {
		if newErr == nil {
			return
		}
		if err == nil {
			err = newErr
		} else {
			err = fmt.Errorf("%v\n%v", err.Error(), newErr.Error())
		}
	}
	targetContainerPos := pos
	anvilPos := targetContainerPos.Add(define.CubePos{1, 0, -1})
	nextContainerPos := targetContainerPos.Add(define.CubePos{1, 0, 1})
	// put simple block/item in container first
	for slot, stack := range containerInfo {
		if stack.Item.GetTypeDescription().IsSimple() {
			if ret := o.cmdHelper.ReplaceContainerItemFullCmd(targetContainerPos, int32(slot), stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.Components).SendAndGetResponse().BlockGetResult(); ret == nil {
				updateErr(fmt.Errorf("cannot put simple block/item in container %v %v %v %v", stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.Components))
			}
		}
	}
	// put block/item needs only enchant in container
	hotBarSlotID := 0
	slotAndEnchant := map[uint8]*neomega.ContainerSlotItemStack{}
	targetSlots := map[uint8]uint8{}
	flush := func() {
		if len(targetSlots) == 0 {
			return
		}
		// wait for a very short time
		o.microAction.SleepTick(5)
		// enchant & rename

		deferActionStand := func() {}
		deferAction := func() {}

		for _, stack := range slotAndEnchant {
			if stack.Item.DisplayName != "" {
				deferActionStand, _ = o.highLevelRemoveSpecificBlockSideEffect(anvilPos.Add(define.CubePos{0, -1, 0}), false, "_temp_anvil_stand")
				o.cmdHelper.SetBlockCmd(anvilPos.Add(define.CubePos{0, -1, 0}), "glass").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
				deferAction, _ = o.highLevelRemoveSpecificBlockSideEffect(anvilPos, false, "_temp_anvil")
				o.cmdHelper.SetBlockCmd(anvilPos, "anvil").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
				break
			}
		}

		defer deferActionStand()
		defer deferAction()

		for hotBarSlot, stack := range slotAndEnchant {
			if len(stack.Item.Enchants) > 0 {
				updateErr(o.highLevelEnchantItem(hotBarSlot, stack.Item.Enchants))
			}
			if stack.Item.DisplayName != "" {
				updateErr(o.highLevelRenameItemWithAnvil(anvilPos, hotBarSlot, stack.Item.DisplayName, false))
			}
		}
		// swap
		updateErr(o.highLevelMoveItemToContainer(targetContainerPos, targetSlots))
		// reset
		hotBarSlotID = 0
		slotAndEnchant = map[uint8]*neomega.ContainerSlotItemStack{}
		targetSlots = map[uint8]uint8{}
		o.microAction.SleepTick(5)
	}
	for _slot, _stack := range containerInfo {
		slot, stack := _slot, _stack
		typeDescription := stack.Item.GetTypeDescription()
		if typeDescription.NeedHotbar() && !typeDescription.IsComplexBlock() {
			if typeDescription.KnownItem() == neomega.KnownItemWritableBook {
				updateErr(o.highLevelWriteBook(uint8(hotBarSlotID), stack.Item.RelatedKnownItemData.Pages))
			} else if typeDescription.KnownItem() == neomega.KnownItemWrittenBook {
				updateErr(o.highLevelWriteBookAndClose(uint8(hotBarSlotID), stack.Item.RelatedKnownItemData.Pages, stack.Item.RelatedKnownItemData.BookName, stack.Item.RelatedKnownItemData.BookAuthor))
			} else {
				if ret := o.cmdHelper.ReplaceBotHotBarItemFullCmd(int32(hotBarSlotID), stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.Components).SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil {
					updateErr(fmt.Errorf("cannot put simple block/item in container %v %v %v %v", stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.Components))
				}
			}
			slotAndEnchant[uint8(hotBarSlotID)] = stack
			targetSlots[uint8(hotBarSlotID)] = slot
			hotBarSlotID++
			if hotBarSlotID == 8 {
				flush()
			}
		}
	}
	flush()
	for _slot, _stack := range containerInfo {
		slot, stack := _slot, _stack
		if stack.Item.GetTypeDescription().IsComplexBlock() {
			o.microAction.SleepTick(5)
			deferActionWorkspace, _ := o.highLevelRemoveSpecificBlockSideEffect(nextContainerPos, false, "_temp_work")
			defer deferActionWorkspace()
			o.cmdHelper.SetBlockCmd(nextContainerPos, fmt.Sprintf("%v %v", stack.Item.Name, stack.Item.RelatedBlockBedrockStateString)).AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
			o.microAction.SleepTick(5)
			updateErr(o.highLevelSetContainerItems(nextContainerPos, stack.Item.RelateComplexBlockData.Container))
			_, err := o.highLevelBlockBreakAndPickInHotBar(nextContainerPos, false, map[uint8]bool{0: false}, 3)
			updateErr(err)
			// give complex block enchant and name
			if len(stack.Item.Enchants) > 0 {
				updateErr(o.highLevelEnchantItem(0, stack.Item.Enchants))
			}
			if stack.Item.DisplayName != "" {
				updateErr(o.highLevelRenameItemWithAnvil(anvilPos, 0, stack.Item.DisplayName, true))
			}
			updateErr(o.highLevelMoveItemToContainer(targetContainerPos, map[uint8]uint8{0: slot}))
		}
	}
	return
}

func (o *BotActionHighLevel) HighLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelRequestLargeArea(startPos, size, dst, withMove)
}

func (o *BotActionHighLevel) highLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error {
	chunkRangesX := neomega.RangeSplits(startPos.X(), size.X(), 16)
	chunkRangesZ := neomega.RangeSplits(startPos.Z(), size.Z(), 16)
	for _, xRange := range chunkRangesX {
		startX := xRange[0]
		for _, zRange := range chunkRangesZ {
			startZ := zRange[0]
			if withMove {
				o.highLevelEnsureBotNearby(define.CubePos{startX, 320, startZ}, 16)
			}
			var err error
			for i := 0; i < 3; i++ {
				var resp neomega.StructureResponse
				var structure *neomega.DecodedStructure
				if err != nil {
					time.Sleep(time.Second)
				}
				resp, err = o.structureRequester.RequestStructure(define.CubePos{startX, startPos.Y(), startZ}, define.CubePos{xRange[1], size.Y(), zRange[1]}, "_tmp").BlockGetResult()
				if err != nil {
					continue
				}
				structure, err = resp.Decode()
				if err != nil {
					continue
				}
				err = structure.DumpToChunkProvider(dst)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
