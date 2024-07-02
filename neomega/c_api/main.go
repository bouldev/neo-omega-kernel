package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"
	"unsafe"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/bundle"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/access_helper"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/info_collect_utils"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
)

var GOmegaCore neomega.MicroOmega
var GPacketNameIDMapping map[string]uint32
var GPacketIDNameMapping map[uint32]string
var GPool packet.Pool

//export OmegaAvailable
func OmegaAvailable() bool {
	return true
}

const (
	EventTypeOmegaConnErr           = "OmegaConnErr"
	EventTypeCommandResponseCB      = "CommandResponseCB"
	EventTypeNewPacket              = "MCPacket"
	EventTypePlayerInterceptedInput = "PlayerInterceptInput"
	EventTypePlayerChange           = "PlayerChange"
	EventTypeChat                   = "Chat"
	EventTypeNamedCommandBlockMsg   = "NamedCommandBlockMsg"
)

type GEvent struct {
	EventType string
	// 调用语言可能希望将某个回调绑定到事件回调上，由于我们无法将事件传入 go 内部，
	// 因此此字段用以帮助调用语言找到目标回调
	RetrieverID string
	// Golang 有自己的GC，当一个数据从 GO 的视野中消失的时候，得假设它可能已经被回收了
	// 然而，实际上event的数据可以是任何类型的
	// 我们要求外部语言在通过 epoll 知道一个 event 发生（知道 event 的 emitter 和 retrieve id 后）
	// 立刻调用 omit event 忽略这个事件（这样 go 也可以顺利回收资源）
	// 或者 consume xxx 将这个事件立刻转为特定的类型
	Data any
}

var GEventsChan chan *GEvent
var GCurrentEvent *GEvent

//export EventPoll
func EventPoll() (EventType *C.char, RetrieverID *C.char) {
	e := <-GEventsChan
	GCurrentEvent = e
	return C.CString(e.EventType), C.CString(e.RetrieverID)
}

//export OmitEvent
func OmitEvent() {
	GCurrentEvent = nil
}

// Async Actions

//export ConsumeCommandResponseCB
func ConsumeCommandResponseCB() *C.char {
	p := (GCurrentEvent.Data).(*packet.CommandOutput)
	bs, _ := json.Marshal(p)
	return C.CString(string(bs))
}

//export SendWebSocketCommandNeedResponse
func SendWebSocketCommandNeedResponse(cmd *C.char, retrieverID *C.char) {
	GoRetrieverID := C.GoString(retrieverID)
	GOmegaCore.GetGameControl().SendWebSocketCmdNeedResponse(C.GoString(cmd)).AsyncGetResult(func(p *packet.CommandOutput) {
		GEventsChan <- &GEvent{EventTypeCommandResponseCB, GoRetrieverID, p}
	})
}

//export SendPlayerCommandNeedResponse
func SendPlayerCommandNeedResponse(cmd *C.char, retrieverID *C.char) {
	GoRetrieverID := C.GoString(retrieverID)
	GOmegaCore.GetGameControl().SendPlayerCmdNeedResponse(C.GoString(cmd)).AsyncGetResult(func(p *packet.CommandOutput) {
		GEventsChan <- &GEvent{EventTypeCommandResponseCB, GoRetrieverID, p}
	})
}

// One-Way Action

//export SendWOCommand
func SendWOCommand(cmd *C.char) {
	GOmegaCore.GetGameControl().SendWOCmd(C.GoString(cmd))
}

//export SendWebSocketCommandOmitResponse
func SendWebSocketCommandOmitResponse(cmd *C.char) {
	GOmegaCore.GetGameControl().SendWebSocketCmdOmitResponse(C.GoString(cmd))
}

//export SendPlayerCommandOmitResponse
func SendPlayerCommandOmitResponse(cmd *C.char) {
	GOmegaCore.GetGameControl().SendPlayerCmdOmitResponse(C.GoString(cmd))
}

//export SendGamePacket
func SendGamePacket(packetID int, jsonStr *C.char) (err *C.char) {
	pk := GPool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return C.CString(_err.Error())
	}
	GOmegaCore.GetGameControl().SendPacket(pk)
	return C.CString("")
}

type NoEOFByteReader struct {
	s []byte
	i int
}

func (nbr *NoEOFByteReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	n = copy(b, nbr.s[nbr.i:])
	nbr.i += n
	return
}

func (nbr *NoEOFByteReader) ReadByte() (b byte, err error) {
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	b = nbr.s[nbr.i]
	nbr.i++
	return b, nil
}

func bytesToCharArr(goByteSlice []byte) *C.char {
	ptr := C.malloc(C.size_t(len(goByteSlice)))
	C.memmove(ptr, (unsafe.Pointer)(&goByteSlice[0]), C.size_t(len(goByteSlice)))
	return (*C.char)(ptr)
}

//export JsonStrAsIsGamePacketBytes
func JsonStrAsIsGamePacketBytes(packetID int, jsonStr *C.char) (pktBytes *C.char, l int, err *C.char) {
	pk := GPool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	b := &bytes.Buffer{}
	w := protocol.NewWriter(b, 0)
	// hdr := pk.ID()
	// w.Varuint32(&hdr)
	pk.Marshal(w)
	bs := b.Bytes()
	l = len(bs)
	return bytesToCharArr(bs), l, nil
}

//export PlaceCommandBlock
func PlaceCommandBlock(option *C.char) {
	opt := neomega.PlaceCommandBlockOption{}
	json.Unmarshal([]byte(C.GoString(option)), &opt)
	GOmegaCore.GetBotAction().HighLevelPlaceCommandBlock(&opt, 3)
	// ba := GOmegaCore.GetGameControl().GenCommandBlockUpdateFromOption(&opt)
	// GOmegaCore.GetGameControl().AsyncPlaceCommandBlock(define.CubePos{
	// 	opt.X, opt.Y, opt.Z,
	// }, opt.BlockName, opt.BlockState, false, false, ba, func(done bool) {
	// 	if !done {
	// 		fmt.Printf("place command block @ [%v,%v,%v] fail\n", opt.X, opt.Y, opt.Z)
	// 	} else {
	// 		fmt.Printf("place command block @ [%v,%v,%v] ok\n", opt.X, opt.Y, opt.Z)
	// 	}
	// }, time.Second*10)
}

// listeners

// disconnect event

//export ConsumeOmegaConnError
func ConsumeOmegaConnError() *C.char {
	err := (GCurrentEvent.Data).(error)
	return C.CString(err.Error())
}

// packet event

var GAllPacketsListenerEnabled = false

//export ListenAllPackets
func ListenAllPackets() {
	if GAllPacketsListenerEnabled {
		panic("should only call ListenAllPackets once")
	}
	GAllPacketsListenerEnabled = true
	GOmegaCore.GetGameListener().SetAnyPacketCallBack(func(p packet.Packet) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeNewPacket,
			RetrieverID: GPacketIDNameMapping[p.ID()],
			Data:        p,
		}
	}, true)
}

//export GetPacketNameIDMapping
func GetPacketNameIDMapping() *C.char {
	marshal, err := json.Marshal(GPacketNameIDMapping)
	if err != nil {
		panic(err)
	}
	return C.CString(string(marshal))
}

//export ConsumeMCPacket
func ConsumeMCPacket() (packetDataAsJsonStr *C.char, convertError *C.char) {
	p := (GCurrentEvent.Data).(packet.Packet)
	marshal, err := json.Marshal(p)
	packetDataAsJsonStr = C.CString(string(marshal))
	convertError = nil
	if err != nil {
		convertError = C.CString(string(err.Error()))
	}
	return
}

// Bot
//
//export GetClientMaintainedBotBasicInfo
func GetClientMaintainedBotBasicInfo() *C.char {
	basicInfo := GOmegaCore.GetMicroUQHolder().GetBotBasicInfo()
	basicInfoMap := map[string]any{
		"BotName":      basicInfo.GetBotName(),
		"BotRuntimeID": basicInfo.GetBotRuntimeID(),
		"BotUniqueID":  basicInfo.GetBotUniqueID(),
		"BotIdentity":  basicInfo.GetBotIdentity(),
		"BotUUIDStr":   basicInfo.GetBotUUIDStr(),
	}
	data, _ := json.Marshal(basicInfoMap)
	return C.CString(string(data))
}

//export GetClientMaintainedExtendInfo
func GetClientMaintainedExtendInfo() *C.char {
	extendInfo := GOmegaCore.GetMicroUQHolder().GetExtendInfo()
	extendInfoMap := map[string]any{}
	if thres, found := extendInfo.GetCompressThreshold(); found {
		extendInfoMap["CompressThreshold"] = thres
	}
	if worldGameMode, found := extendInfo.GetWorldGameMode(); found {
		extendInfoMap["WorldGameMode"] = worldGameMode
	}
	if worldDifficulty, found := extendInfo.GetWorldDifficulty(); found {
		extendInfoMap["WorldDifficulty"] = worldDifficulty
	}
	if time, found := extendInfo.GetTime(); found {
		extendInfoMap["Time"] = time
	}
	if dayTime, found := extendInfo.GetDayTime(); found {
		extendInfoMap["DayTime"] = dayTime
	}
	if timePercent, found := extendInfo.GetDayTimePercent(); found {
		extendInfoMap["TimePercent"] = timePercent
	}
	if gameRules, found := extendInfo.GetGameRules(); found {
		extendInfoMap["GameRules"] = gameRules
	}
	data, _ := json.Marshal(extendInfoMap)
	return C.CString(string(data))
}

// players
var GPlayers *sync_wrapper.SyncKVMap[string, neomega.PlayerKit]

//export GetAllOnlinePlayers
func GetAllOnlinePlayers() *C.char {
	players := GOmegaCore.GetPlayerInteract().ListAllPlayers()
	retPlayers := []string{}
	for _, player := range players {
		uuidStr, _ := player.GetUUIDString()
		GPlayers.Set(uuidStr, player)
		retPlayers = append(retPlayers, uuidStr)
	}
	data, _ := json.Marshal(retPlayers)
	return C.CString(string(data))
}

//export GetPlayerByName
func GetPlayerByName(name *C.char) *C.char {
	player, found := GOmegaCore.GetPlayerInteract().GetPlayerKit(C.GoString(name))
	if found {
		uuidStr, _ := player.GetUUIDString()
		GPlayers.Set(uuidStr, player)
		return C.CString(uuidStr)
	}
	return C.CString("")
}

//export GetPlayerByUUID
func GetPlayerByUUID(uuid *C.char) *C.char {
	player, found := GOmegaCore.GetPlayerInteract().GetPlayerKitByUUIDString(C.GoString(uuid))
	if found {
		uuidStr, _ := player.GetUUIDString()
		GPlayers.Set(uuidStr, player)
		return C.CString(uuidStr)
	}
	return C.CString("")
}

//export ReleaseBindPlayer
func ReleaseBindPlayer(uuidStr *C.char) {
	GPlayers.Delete(C.GoString(uuidStr))
}

//export PlayerName
func PlayerName(uuidStr *C.char) *C.char {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	name, _ := p.GetUsername()
	return C.CString(name)
}

//export PlayerEntityUniqueID
func PlayerEntityUniqueID(uuidStr *C.char) int64 {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	entityUniqueID, _ := p.GetEntityUniqueID()
	return entityUniqueID
}

//export PlayerLoginTime
func PlayerLoginTime(uuidStr *C.char) int64 {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	loginTime, _ := p.GetLoginTime()
	return loginTime.Unix()
}

//export PlayerPlatformChatID
func PlayerPlatformChatID(uuidStr *C.char) *C.char {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	name, _ := p.GetPlatformChatID()
	return C.CString(name)
}

//export PlayerBuildPlatform
func PlayerBuildPlatform(uuidStr *C.char) int32 {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	buildPlatform, _ := p.GetBuildPlatform()
	return buildPlatform
}

//export PlayerSkinID
func PlayerSkinID(uuidStr *C.char) *C.char {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	SkinID, _ := p.GetSkinID()
	return C.CString(SkinID)
}

// //export PlayerPropertiesFlag
// func PlayerPropertiesFlag(uuidStr *C.char) uint32 {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	PropertiesFlag, _ := p.GetPropertiesFlag()
// 	return PropertiesFlag
// }

// //export PlayerCommandPermissionLevel
// func PlayerCommandPermissionLevel(uuidStr *C.char) uint32 {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	CommandPermissionLevel, _ := p.GetCommandPermissionLevel()
// 	return CommandPermissionLevel
// }

// //export PlayerActionPermissions
// func PlayerActionPermissions(uuidStr *C.char) uint32 {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	ActionPermissions, _ := p.GetActionPermissions()
// 	return ActionPermissions
// }

// //export PlayerGetAbilityString
// func PlayerGetAbilityString(uuidStr *C.char) *C.char {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	adventureFlagsMap, actionPermissionMap, _ := p.GetAbilityString()
// 	abilityMap := map[string]map[string]bool{
// 		"AdventureFlagsMap":   adventureFlagsMap,
// 		"ActionPermissionMap": actionPermissionMap,
// 	}
// 	data, _ := json.Marshal(abilityMap)
// 	return C.CString(string(data))
// }

// //export PlayerOPPermissionLevel
// func PlayerOPPermissionLevel(uuidStr *C.char) uint32 {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	OPPermissionLevel, _ := p.GetOPPermissionLevel()
// 	return OPPermissionLevel
// }

// //export PlayerCustomStoredPermissions
// func PlayerCustomStoredPermissions(uuidStr *C.char) uint32 {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	CustomStoredPermissions, _ := p.GetCustomStoredPermissions()
// 	return CustomStoredPermissions
// }

//export PlayerCanBuild
func PlayerCanBuild(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanBuild()
	return hasAbility
}

//export PlayerSetBuild
func PlayerSetBuild(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetBuildAbility(allow)
}

//export PlayerCanMine
func PlayerCanMine(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanMine()
	return hasAbility
}

//export PlayerSetMine
func PlayerSetMine(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetMineAbility(allow)
}

//export PlayerCanDoorsAndSwitches
func PlayerCanDoorsAndSwitches(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanDoorsAndSwitches()
	return hasAbility
}

//export PlayerSetDoorsAndSwitches
func PlayerSetDoorsAndSwitches(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetDoorsAndSwitchesAbility(allow)
}

//export PlayerCanOpenContainers
func PlayerCanOpenContainers(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanOpenContainers()
	return hasAbility
}

//export PlayerSetOpenContainers
func PlayerSetOpenContainers(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetOpenContainersAbility(allow)
}

//export PlayerCanAttackPlayers
func PlayerCanAttackPlayers(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanAttackPlayers()
	return hasAbility
}

//export PlayerSetAttackPlayers
func PlayerSetAttackPlayers(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetAttackPlayersAbility(allow)
}

//export PlayerCanAttackMobs
func PlayerCanAttackMobs(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanAttackMobs()
	return hasAbility
}

//export PlayerSetAttackMobs
func PlayerSetAttackMobs(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetAttackMobsAbility(allow)
}

//export PlayerCanOperatorCommands
func PlayerCanOperatorCommands(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanOperatorCommands()
	return hasAbility
}

//export PlayerSetOperatorCommands
func PlayerSetOperatorCommands(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetOperatorCommandsAbility(allow)
}

//export PlayerCanTeleport
func PlayerCanTeleport(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.CanTeleport()
	return hasAbility
}

//export PlayerSetTeleport
func PlayerSetTeleport(uuidStr *C.char, allow bool) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SetTeleportAbility(allow)
}

//export PlayerStatusInvulnerable
func PlayerStatusInvulnerable(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.StatusInvulnerable()
	return hasAbility
}

//export PlayerStatusFlying
func PlayerStatusFlying(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.StatusFlying()
	return hasAbility
}

//export PlayerStatusMayFly
func PlayerStatusMayFly(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	hasAbility, _ := p.StatusMayFly()
	return hasAbility
}

//export PlayerDeviceID
func PlayerDeviceID(uuidStr *C.char) *C.char {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	name, _ := p.GetDeviceID()
	return C.CString(name)
}

//export PlayerEntityRuntimeID
func PlayerEntityRuntimeID(uuidStr *C.char) uint64 {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	if p == nil {
		return 0
	}
	EntityRuntimeID, found := p.GetEntityRuntimeID()
	if !found {
		return 0
	}
	return EntityRuntimeID
}

//export PlayerEntityMetadata
func PlayerEntityMetadata(uuidStr *C.char) *C.char {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	entityMetadata, _ := p.GetEntityMetadata()
	data, _ := json.Marshal(entityMetadata)
	return C.CString(string(data))
}

//export PlayerIsOP
func PlayerIsOP(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	isOP, _ := p.IsOP()
	return isOP
}

//export PlayerOnline
func PlayerOnline(uuidStr *C.char) bool {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	return p.StillOnline()
}

//export PlayerChat
func PlayerChat(uuidStr *C.char, msg *C.char) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.Say(C.GoString(msg))
}

//export PlayerTitle
func PlayerTitle(uuidStr *C.char, title, subTitle *C.char) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.SubTitle(C.GoString(subTitle), C.GoString(title))
}

//export PlayerActionBar
func PlayerActionBar(uuidStr *C.char, actionBar *C.char) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	p.ActionBar(C.GoString(actionBar))
}

// //export SetPlayerAbility
// func SetPlayerAbility(uuidStr *C.char, jsonFlags *C.char) {
// 	p, _ := GPlayers.Get(C.GoString(uuidStr))
// 	// abilityMap := map[string]map[string]bool{
// 	// 	"AdventureFlagsMap":   adventureFlagsMap,
// 	// 	"ActionPermissionMap": actionPermissionMap,
// 	// }
// 	abilityMap := map[string]map[string]bool{}
// 	json.Unmarshal([]byte(C.GoString(jsonFlags)), &abilityMap)
// 	adventureFlagsMap := abilityMap["AdventureFlagsMap"]
// 	actionPermissionMap := abilityMap["ActionPermissionMap"]
// 	fmt.Println(adventureFlagsMap)
// 	fmt.Println(actionPermissionMap)
// 	p.SetAbilityString(adventureFlagsMap, actionPermissionMap)
// }

//export InterceptPlayerJustNextInput
func InterceptPlayerJustNextInput(uuidStr *C.char, retrieverID *C.char) {
	p, _ := GPlayers.Get(C.GoString(uuidStr))
	retrieverIDStr := C.GoString(retrieverID)
	p.InterceptJustNextInput(func(chat *neomega.GameChat) {
		GEventsChan <- &GEvent{
			EventType:   EventTypePlayerInterceptedInput,
			RetrieverID: retrieverIDStr,
			Data:        chat,
		}
	})
}

//export ConsumeChat
func ConsumeChat() *C.char {
	chat := GCurrentEvent.Data.(*neomega.GameChat)
	bs, _ := json.Marshal(chat)
	return C.CString(string(bs))
}

var GListenPlayerChangeListened = false

//export ListenPlayerChange
func ListenPlayerChange() {
	if GListenPlayerChangeListened {
		panic("ListenPlayerChange should only called once")
	}
	GListenPlayerChangeListened = true
	GOmegaCore.GetPlayerInteract().ListenPlayerChange(func(player neomega.PlayerKit, action string) {
		uuidStr, _ := player.GetUUIDString()
		GPlayers.Set(uuidStr, player)
		GEventsChan <- &GEvent{
			EventType:   EventTypePlayerChange,
			RetrieverID: uuidStr,
			Data:        action,
		}
	})
}

//export ConsumePlayerChange
func ConsumePlayerChange() (change *C.char) {
	return C.CString(GCurrentEvent.Data.(string))
}

var GListenChatListened = false

//export ListenChat
func ListenChat() {
	if GListenChatListened {
		panic("ListenPlayerChat should only called once")
	}
	GListenChatListened = true
	GOmegaCore.GetPlayerInteract().SetOnChatCallBack(func(chat *neomega.GameChat) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeChat,
			RetrieverID: "",
			Data:        chat,
		}
	})
}

//export ListenCommandBlock
func ListenCommandBlock(name *C.char) {
	gName := C.GoString(name)
	GOmegaCore.GetPlayerInteract().SetOnSpecificCommandBlockTellCallBack(gName, func(chat *neomega.GameChat) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeNamedCommandBlockMsg,
			RetrieverID: gName,
			Data:        chat,
		}
	})
}

// utils

//export FreeMem
func FreeMem(address unsafe.Pointer) {
	C.free(address)
}

func prepareOmegaAPIs(omegaCore neomega.MicroOmega) {
	GEventsChan = make(chan *GEvent, 1024)
	GOmegaCore = omegaCore
	GPacketNameIDMapping = GOmegaCore.GetGameListener().GetMCPacketNameIDMapping()
	{
		GPacketIDNameMapping = map[uint32]string{}
		for name, id := range GPacketNameIDMapping {
			GPacketIDNameMapping[id] = name
		}
	}
	GPool = packet.NewPool()
	go func() {
		err := <-omegaCore.Dead()
		GOmegaCore = nil
		GEventsChan <- &GEvent{
			EventTypeOmegaConnErr,
			"",
			err,
		}
	}()
	GPlayers = sync_wrapper.NewSyncKVMap[string, neomega.PlayerKit]()
}

//export ConnectOmega
func ConnectOmega(address *C.char) (Cerr *C.char) {
	if GOmegaCore != nil {
		return C.CString("connect has been established")
	}
	var node defines.Node
	// ctx := context.Background()
	{
		client, err := underlay_conn.NewClientFromBasicNet(C.GoString(address), time.Second)
		if err != nil {
			return C.CString(err.Error())
		}
		slave, err := nodes.NewZMQSlaveNode(client)
		if err != nil {
			return C.CString(err.Error())
		}
		node = nodes.NewGroup("github.com/OmineDev/neomega-core/neomega", slave, false)
		if !node.CheckNetTag("access-point") {
			return C.CString(i18n.T(i18n.S_no_access_point_in_network))
		}
	}
	omegaCore, err := bundle.NewEndPointMicroOmega(node)
	if err != nil {
		return C.CString(err.Error())
	}
	prepareOmegaAPIs(omegaCore)
	return nil
}

//export StartOmega
func StartOmega(address *C.char, impactOptionsJson *C.char) (Cerr *C.char) {
	if GOmegaCore != nil {
		return C.CString("connect has been established")
	}
	var node defines.Node
	accessOption := access_helper.DefaultOptions()
	// ctx := context.Background()
	{
		impactOption := &access_helper.ImpactOption{}
		json.Unmarshal([]byte(C.GoString(impactOptionsJson)), &impactOption)
		if err := info_collect_utils.ReadUserInfoAndUpdateImpactOptions(impactOption); err != nil {
			return C.CString(err.Error())
		}

		accessOption.ImpactOption = impactOption
		accessOption.MakeBotCreative = true
		accessOption.DisableCommandBlock = false
		accessOption.ReasonWithPrivilegeStuff = true

		{
			server, err := underlay_conn.NewServerFromBasicNet(C.GoString(address))
			if err != nil {
				panic(err)
			}
			// server := nodes.NewSimpleZMQServer(socket)
			master := nodes.NewZMQMasterNode(server)
			node = nodes.NewGroup("github.com/OmineDev/neomega-core/neomega", master, false)
		}
	}
	ctx := context.Background()
	omegaCore, err := access_helper.ImpactServer(ctx, node, accessOption)
	if err != nil {
		return C.CString(err.Error())
	}
	prepareOmegaAPIs(omegaCore)
	return nil
}

func main() {
	//Windows: go build  -tags fbconn -o fbconn.dll -buildmode=c-shared main.go
	//Linux: go build -tags fbconn -o libfbconn.so -buildmode=c-shared main.go
	//Macos: go build -o omega_conn.dylib -buildmode=c-shared main.go
	//将生成的文件 (fbconn.dll 或 libfbconn.so 或 fbconn.dylib) 放在 conn.py 同一个目录下
}
