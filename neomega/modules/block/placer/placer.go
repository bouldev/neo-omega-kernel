package placer

import (
	"context"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/neomega/mirror/define"
	"strings"
	"sync"
	"time"
)

func init() {
	if false {
		func(neomega.BlockPlacer) {}(&BlockPlacer{})
	}
}

type BlockPlacer struct {
	onBlockActorCbs map[define.CubePos]func(define.CubePos, *packet.BlockActorData)
	blockActorLock  sync.Mutex
	neomega.CmdSender
	neomega.GameIntractable
}

func NewAsyncNbtBlockPlacer(reactable neomega.ReactCore, cmdSender neomega.CmdSender, packetSender neomega.GameIntractable) neomega.BlockPlacer {
	c := &BlockPlacer{
		onBlockActorCbs: make(map[define.CubePos]func(define.CubePos, *packet.BlockActorData)),
		blockActorLock:  sync.Mutex{},
		CmdSender:       cmdSender,
		GameIntractable: packetSender,
	}

	reactable.SetTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
		c.onBlockActor(p.(*packet.BlockActorData))
	}, false)
	return c
}

func (c *BlockPlacer) onBlockActor(p *packet.BlockActorData) {
	pos := define.CubePos{int(p.Position.X()), int(p.Position.Y()), int(p.Position.Z())}
	c.blockActorLock.Lock()
	if cb, found := c.onBlockActorCbs[pos]; found {
		delete(c.onBlockActorCbs, pos)
		c.blockActorLock.Unlock()
		cb(pos, p)
	} else {
		c.blockActorLock.Unlock()
	}
}

func unpackCommandBlockTag(tag []byte) map[string]interface{} {
	return nil
}

func (g *BlockPlacer) GenCommandBlockUpdateFromNbt(pos define.CubePos, blockName string, blockState map[string]interface{}, nbt map[string]interface{}) (cfg *packet.CommandBlockUpdate, err error) {
	tag, exist := nbt["__tag"]
	if exist {
		data := []byte(tag.(string))
		nbt = unpackCommandBlockTag(data)
	}
	if nbt == nil {
		return nil, nil
	}
	var mode uint32
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot gen block update %v", r)
			cfg = nil
		}
	}()
	if blockName == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if blockName == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if blockName == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		return nil, fmt.Errorf("not command block")
	}
	cmd, _ := nbt["Command"].(string)
	cusname, _ := nbt["CustomName"].(string)
	exeft, _ := nbt["ExecuteOnFirstTick"].(uint8)
	tickdelay, _ := nbt["TickDelay"].(int32)     //*/
	aut, _ := nbt["auto"].(uint8)                //!needrestone
	trackoutput, _ := nbt["TrackOutput"].(uint8) //
	lo, _ := nbt["LastOutput"].(string)
	conditionalmode, ok := nbt["conditionalMode"].(uint8)
	if !ok {
		conditionalmode = blockState["conditional_bit"].(uint8)
	}
	//conditionalmode := nbt["conditionalMode"].(uint8)
	var exeftb bool
	if exeft == 0 {
		exeftb = false
	} else {
		exeftb = true
	}
	var tob bool
	if trackoutput == 1 {
		tob = true
	} else {
		tob = false
	}
	var nrb bool
	if aut == 1 {
		nrb = false
		//REVERSED!!
	} else {
		nrb = true
	}
	var conb bool
	if conditionalmode == 1 {
		conb = true
	} else {
		conb = false
	}
	return &packet.CommandBlockUpdate{
		Block:              true,
		Position:           protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
		Mode:               mode,
		NeedsRedstone:      nrb,
		Conditional:        conb,
		Command:            cmd,
		LastOutput:         lo,
		Name:               cusname,
		TickDelay:          tickdelay,
		ExecuteOnFirstTick: exeftb,
		ShouldTrackOutput:  tob,
	}, nil
}

func (g *BlockPlacer) GenCommandBlockUpdateFromOption(opt *neomega.PlaceCommandBlockOption) *packet.CommandBlockUpdate {
	var mode uint32
	if opt.BlockName == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if opt.BlockName == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if opt.BlockName == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		opt.BlockName = "command_block"
		mode = packet.CommandBlockImpulse
	}
	return &packet.CommandBlockUpdate{
		Block:              true,
		Position:           protocol.BlockPos{int32(opt.X), int32(opt.Y), int32(opt.Z)},
		Mode:               mode,
		NeedsRedstone:      opt.NeedRedStone,
		Conditional:        opt.Conditional,
		Command:            opt.Command,
		LastOutput:         "",
		Name:               opt.Name,
		TickDelay:          int32(opt.TickDelay),
		ExecuteOnFirstTick: opt.ExecuteOnFirstTick,
		ShouldTrackOutput:  opt.ShouldTrackOutput,
	}
}

func (g *BlockPlacer) AsyncPlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string,
	withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
	onDone func(done bool), timeOut time.Duration) {
	commandBlockPlacedCtx, commandBlockPlaced := context.WithCancel(context.Background())
	ctx, done := context.WithCancel(context.Background())
	go func() {
		select {
		case <-time.NewTimer(timeOut).C:
			onDone(false)
		case <-ctx.Done():
			onDone(true)
		}
	}()
	var checkAndSendUpdate func(cp define.CubePos, bad *packet.BlockActorData)
	checkAndSendUpdate = func(cp define.CubePos, bad *packet.BlockActorData) {
		nbt := bad.NBTData
		ok := true
		if nbt["id"].(string) == "CommandBlock" {
			commandBlockPlaced()
		}
		if nbt["id"].(string) != "CommandBlock" {
			ok = false
		} else if nbt["Command"].(string) != updatePacket.Command {
			ok = false
		} else if nbt["CustomName"].(string) != updatePacket.Name {
			ok = false
		}
		if ok {
			done()
			return
		}
		if ctx.Err() == nil {
			g.blockActorLock.Lock()
			g.onBlockActorCbs[pos] = checkAndSendUpdate
			g.blockActorLock.Unlock()
			g.SendPacket(updatePacket)
		}
	}
	g.blockActorLock.Lock()
	g.onBlockActorCbs[pos] = checkAndSendUpdate
	g.blockActorLock.Unlock()
	if withMove && ctx.Err() == nil {
		g.SendWebSocketCmdOmitResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
		time.Sleep(6 * 50 * time.Millisecond)
	}
	if withAirPrePlace && commandBlockPlacedCtx.Err() == nil && ctx.Err() == nil {
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
		g.SendWebSocketCmdOmitResponse(cmd)
	}
	if commandBlockPlacedCtx.Err() == nil && ctx.Err() == nil {
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(commandBlockName, "minecraft:", "", 1), blockDataOrStateStr)
		g.SendWebSocketCmdOmitResponse(cmd)
	}
	select {
	case <-commandBlockPlacedCtx.Done():
		g.SendPacket(updatePacket)
	case <-time.NewTimer(50 * time.Millisecond).C:
		g.SendPacket(updatePacket)
	case <-ctx.Done():

	}
	//
	//g.SendWebSocketCmdNeedResponse(cmd).AsyncGetResult(func(output *packet.CommandOutput) {
	//	time.Sleep(300 * time.Millisecond)
	//	g.SendPacket(updatePacket)
	//})
	//time.Sleep(300 * time.Millisecond)
}

// func (g *BlockPlacer) PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration) {
// 	done := make(chan bool)
// 	go func() {
// 		select {
// 		case <-time.NewTimer(timeOut).C:
// 			onDone(false)
// 		case <-done:
// 		}
// 	}()
// 	go func() {
// 		if withMove {
// 			g.SendWebSocketCmdOmitResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
// 			time.Sleep(100 * time.Millisecond)
// 		}
// 		if withAirPrePlace {
// 			cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
// 			g.SendWOCmd(cmd)
// 			time.Sleep(100 * time.Millisecond)
// 		}
// 		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(signBlockName, "minecraft:", "", 1), blockDataOrStateStr)
// 		g.SendWOCmd(cmd)
// 		g.blockActorLock.Lock()
// 		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
// 			go func() {
// 				g.blockActorLock.Lock()
// 				g.SendWebSocketCmdOmitResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
// 				time.Sleep(50 * time.Millisecond)
// 				g.SendPacket(updatePacket)
// 				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
// 					onDone(true)
// 					done <- true
// 				}
// 				g.blockActorLock.Unlock()
// 			}()
// 		}
// 		g.blockActorLock.Unlock()
// 	}()
// }
