package neomega

import (
	"context"
	"fmt"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega/blocks"
	"neo-omega-kernel/neomega/chunks/define"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 可以向游戏发送数据包
type GameIntractable interface {
	SendPacket(packet.Packet)
	SendPacketBytes(pktID uint32, data []byte)
}

// NoBlockAndDetachablePacketCallBack 表示没有阻塞的数据处理函数类型
// 当不需要继续读取数据时，返回 err
type NoBlockAndDetachablePacketCallback func(pk packet.Packet) error

type PacketDispatcher interface {
	SetAnyPacketCallBack(callback func(packet.Packet), newGoroutine bool)
	SetTypedPacketCallBack(packetID uint32, callback func(packet.Packet), newGoroutine bool)
	// SetOneTimeTypedPacketNoBlockCallBack(uint32, func(packet.Packet) (next bool))
	// GetMCPacketNameIDMapping 返回协议包名字与 ID 的映射关系
	GetMCPacketNameIDMapping() map[string]uint32
	// TranslateStringWantsToIDSet 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
	TranslateStringWantsToIDSet(want []string) map[uint32]bool
	AddNewNoBlockAndDetachablePacketCallBack(wants map[uint32]bool, cb NoBlockAndDetachablePacketCallback)
}

type ReactCore interface {
	Dead() chan error
	PacketDispatcher
}

type UnStartedReactCore interface {
	ReactCore
	Start()
}

type InteractCore interface {
	GameIntractable
}

type ResponseHandle interface {
	SetTimeoutResponse(timeoutResponse *packet.CommandOutput) ResponseHandle
	SetTimeout(timeOut time.Duration) ResponseHandle
	SetContext(ctx context.Context) ResponseHandle
	BlockGetResult() *packet.CommandOutput
	// callback will be call in a new goroutine, need more resource but safe
	AsyncGetResult(func(output *packet.CommandOutput))
}

type CmdSender interface {
	SendWOCmd(cmd string)
	SendWebSocketCmdOmitResponse(cmd string)
	SendPlayerCmdOmitResponse(cmd string)
	SendAICommandOmitResponse(runtimeid string, cmd string)

	SendWebSocketCmdNeedResponse(cmd string) ResponseHandle
	SendPlayerCmdNeedResponse(cmd string) ResponseHandle
	SendAICommandNeedResponse(runtimeid string, cmd string) ResponseHandle
}

// type CmdSender interface {
// 	SendWebSocketCmdOmitResponse(cmd string)
// 	SendWOCmd(cmd string)
// 	SendPlayerCmdOmitResponse(cmd string)
// 	SendWebSocketCmdNeedResponse(string, func(output *packet.CommandOutput))
// 	SendPlayerCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
// 	SendWSCmdAndInvokeOnResponseNoBlock(string, func(output *packet.CommandOutput))
// 	SendPlayerCmdAndInvokeOnResponseWithFeedbackNoBlock(string, func(output *packet.CommandOutput))
// }

type InfoSender interface {
	BotSay(msg string)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, subTitle string, title string)
}

type PlaceCommandBlockOption struct {
	X, Y, Z            int
	BlockName          string
	BlockState         string
	NeedRedStone       bool
	Conditional        bool
	Command            string
	Name               string
	TickDelay          int
	ShouldTrackOutput  bool
	ExecuteOnFirstTick bool
}

func (opt *PlaceCommandBlockOption) String() string {
	describe := ""
	describe += fmt.Sprintf("[%v,%v,%v]%v%v", opt.X, opt.Y, opt.Z, opt.BlockName, opt.BlockState)
	if opt.Name != "" {
		describe += fmt.Sprintf("\n  名字: %v", opt.Name)
	}
	if opt.Command != "" {
		describe += fmt.Sprintf("\n  指令: %v", opt.Command)
	}
	options := fmt.Sprintf("\n  红石=%v,有条件=%v,显示输出=%v,执行第一个已选项=%v,延迟=%v", opt.NeedRedStone, opt.Conditional, opt.ShouldTrackOutput, opt.ExecuteOnFirstTick, opt.TickDelay)
	describe += strings.ReplaceAll(strings.ReplaceAll(options, "true", "是"), "false", "否")
	return describe
}

func (opt *PlaceCommandBlockOption) GenCommandBlockUpdateFromOption() *packet.CommandBlockUpdate {
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

func NewPlaceCommandBlockOptionFromNBT(pos define.CubePos, blockNameAndState string, nbt map[string]interface{}) (o *PlaceCommandBlockOption, err error) {
	rtid, found := blocks.BlockStrToRuntimeID(blockNameAndState)
	if !found {
		return nil, fmt.Errorf("cannot recognize this block %v", blockNameAndState)
	}
	return NewPlaceCommandBlockOptionFromNBTAndRtid(pos, rtid, nbt)
}

func NewPlaceCommandBlockOptionFromNBTAndRtid(pos define.CubePos, rtid uint32, nbt map[string]interface{}) (o *PlaceCommandBlockOption, err error) {
	_, exist := nbt["__tag"]
	if exist {
		return nil, fmt.Errorf("flatten nemc nbt, cannot handle")
	}
	if nbt == nil {
		return nil, fmt.Errorf("nbt is empty, cannot handle")
	}
	var mode uint32
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot gen place command block option %v", r)
		}
	}()
	block, _ := blocks.RuntimeIDToBlock(rtid)
	if block.ShortName() == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if block.ShortName() == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if block.ShortName() == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		return nil, fmt.Errorf("this block %v%v is not command block", block.ShortName(), block.States().BedrockString(true))
	}
	mode = mode // just make compiler happy
	cmd, _ := nbt["Command"].(string)
	constumeName, _ := nbt["CustomName"].(string)
	exeft, _ := nbt["ExecuteOnFirstTick"].(uint8)
	tickdelay, _ := nbt["TickDelay"].(int32)     //*/
	aut, _ := nbt["auto"].(uint8)                //!needrestone
	trackoutput, _ := nbt["TrackOutput"].(uint8) //
	// lo, _ := nbt["LastOutput"].(string)
	conditionalmode, ok := nbt["conditionalMode"].(uint8)
	if !ok {
		conditionalmode = block.States().ToNBT()["conditional_bit"].(uint8)
	}
	//conditionalmode := nbt["conditionalMode"].(uint8)
	var executeOnFirstTickBit bool
	if exeft == 0 {
		executeOnFirstTickBit = false
	} else {
		executeOnFirstTickBit = true
	}
	var trackOutputBit bool
	if trackoutput == 1 {
		trackOutputBit = true
	} else {
		trackOutputBit = false
	}
	var needRedStoneBit bool
	if aut == 1 {
		needRedStoneBit = false
		//REVERSED!!
	} else {
		needRedStoneBit = true
	}
	var conditionalBit bool
	if conditionalmode == 1 {
		conditionalBit = true
	} else {
		conditionalBit = false
	}
	o = &PlaceCommandBlockOption{
		X: pos.X(), Y: pos.Y(), Z: pos.Z(),
		BlockName:          block.ShortName(),
		BlockState:         block.States().BedrockString(true),
		NeedRedStone:       needRedStoneBit,
		Conditional:        conditionalBit,
		Command:            cmd,
		Name:               constumeName,
		TickDelay:          int(tickdelay),
		ShouldTrackOutput:  trackOutputBit,
		ExecuteOnFirstTick: executeOnFirstTickBit,
	}
	return o, nil
}

// type AsyncNBTBlockPlacer interface {
// 	GenCommandBlockUpdateFromNbt(pos define.CubePos, blockName string, blockState map[string]interface{}, nbt map[string]interface{}) (cfg *packet.CommandBlockUpdate, err error)
// 	GenCommandBlockUpdateFromOption(opt *PlaceCommandBlockOption) *packet.CommandBlockUpdate
// 	AsyncPlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
// 		onDone func(done bool), timeOut time.Duration)
// 	// PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration)
// }

// type BlockPlacer interface {
// 	AsyncNBTBlockPlacer
// }

type GameChat struct {
	// 玩家名（去除前缀, e.g. <乱七八糟的前缀> 张三 -> 张三）
	Name string
	// 分割后的消息 (e.g. "前往 1 2 3" ["前往", "1", "2", "3"])
	Msg []string
	// e.g. packet.TextTypeChat
	Type byte
	// 原始消息，未分割
	RawMsg string
	// 原始玩家名，未分割
	RawName string
	// 原始参数， packet.Text 中的 Parameters
	RawParameters []string
	// 附加信息，用于传递额外的信息，一般为空，也可能是 packet.Text 数据包，不要过于依赖这个字段
	Aux any
}

type PlayerMsgListener interface {
	InterceptJustNextInput(playerName string, cb func(chat *GameChat))
	SetOnChatCallBack(func(chat *GameChat))
	SetOnSpecificCommandBlockTellCallBack(commandBlockName string, cb func(chat *GameChat))
	SetOnSpecificItemMsgCallBack(itemName string, cb func(chat *GameChat))
}

type PlayerInteract interface {
	PlayerMsgListener
	ListAllPlayers() []PlayerKit
	ListenPlayerChange(cb func(player PlayerKit, action string))
	GetPlayerKit(name string) (playerKit PlayerKit, found bool)
	GetPlayerKitByUUID(ud uuid.UUID) (playerKit PlayerKit, found bool)
	GetPlayerKitByUUIDString(ud string) (playerKit PlayerKit, found bool)
	GetPlayerKitByUniqueID(uniqueID int64) (playerKit PlayerKit, found bool)
}

type QueryResult struct {
	Dimension int `json:"dimension"`
	Position  *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`
	UUID string  `json:"uniqueId"`
	YRot float64 `json:"yRot"`
}

type PlayerKit interface {
	PlayerUQReader
	Say(msg string)
	RawSay(msg string)
	ActionBar(msg string)
	Title(msg string)
	SubTitle(subTitle string, title string)
	InterceptJustNextInput(cb func(chat *GameChat))
	// e.g. CheckCondition(func(ok),"m=c","tag=op")
	CheckCondition(onResult func(bool), conditions ...string)
	Query(onResult func([]QueryResult), conditions ...string)
	SetAbility(adventureFlagsUpdateMap, actionPermissionUpdateMap map[uint32]bool) (sent bool)
	SetAbilityString(adventureFlagsUpdateMap, actionPermissionUpdateMap map[string]bool) (sent bool)
}

type GameCtrl interface {
	InteractCore
	CmdSender
	InfoSender
	// BlockPlacer
}

type MicroOmega interface {
	Dead() chan error
	GetGameControl() GameCtrl
	GetGameListener() PacketDispatcher
	GetMicroUQHolder() MicroUQHolder
	GetPlayerInteract() PlayerInteract
	GetStructureRequester() StructureRequester
	GetBotAction() BotActionComplex
}

type UnReadyMicroOmega interface {
	MicroOmega
	// i dont wanto add this, but sometimes we need cmd to init stuffs
	NotifyChallengePassed()
	PostponeActionsAfterChallengePassed(name string, action func())
}
