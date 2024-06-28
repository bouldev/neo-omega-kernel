package neomega

import (
	"encoding/json"
	"math"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega/chunks"
	"neo-omega-kernel/neomega/chunks/define"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mitchellh/mapstructure"
)

type PosAndDimensionInfo struct {
	Dimension      int
	InOverWorld    bool
	InNether       bool
	InEnd          bool
	HeadPosPrecise mgl32.Vec3
	FeetPosPrecise mgl32.Vec3
	HeadBlockPos   define.CubePos
	FeetBlockPos   define.CubePos
	YRot           float64
}

// information safe to be set from outside & when setting, no bot action module is required
// only information which is shared and *read only* by different modules should be placed here
type BotActionInfo struct {
	*PosAndDimensionInfo
	TempStructureName string
}

type StructureBlockSettings struct {
	DataField         string  `mapstructure:"dataField"`
	IgnoreEntities    uint8   `mapstructure:"ignoreEntities"`
	IncludePlayers    uint8   `mapstructure:"includePlayers"`
	Integrity         float32 `mapstructure:"integrity"`
	Mirror            uint8   `mapstructure:"mirror"`
	RedstoneSaveMode  int32   `mapstructure:"redstoneSaveMode"`
	IgnoreBlocks      int32   `mapstructure:"removeBlocks"`
	Rotation          uint8   `mapstructure:"rotation"`
	ShowBoundingBox   uint8   `mapstructure:"showBoundingBox"`
	AnimationMode     uint8   `mapstructure:"animationMode"`
	AnimationDuration float32 `mapstructure:"animationSeconds"`
	Seed              int64   `mapstructure:"seed"`

	XStructureSize int32 `mapstructure:"xStructureSize"`
	YStructureSize int32 `mapstructure:"yStructureSize"`
	ZStructureSize int32 `mapstructure:"zStructureSize"`

	XStructureOffset int32 `mapstructure:"xStructureOffset"`
	YStructureOffset int32 `mapstructure:"yStructureOffset"`
	ZStructureOffset int32 `mapstructure:"zStructureOffset"`

	StructureName      string `mapstructure:"structureName"`
	StructureBlockType int32
}

func (s *StructureBlockSettings) FromNBT(nbt map[string]any) {
	mapstructure.Decode(nbt, s)
}

func (s *StructureBlockSettings) LoadTypeFromString(state string) {
	if strings.Contains(state, "save") {
		s.StructureBlockType = packet.StructureBlockSave
	} else if strings.Contains(state, "load") {
		s.StructureBlockType = packet.StructureBlockLoad
	} else if strings.Contains(state, "corner") {
		s.StructureBlockType = packet.StructureBlockCorner
	} else if strings.Contains(state, "data") {
		s.StructureBlockType = packet.StructureBlockData
	}
}

type CmdCannotGetResponse interface {
	Send()
}

type CmdCanGetResponse interface {
	CmdCannotGetResponse
	SendAndGetResponse() ResponseHandle
}

type GeneralCommand interface {
	Send()
	AsPlayer() CmdCanGetResponse
	AsWebSocket() CmdCanGetResponse
}

// HighLevel means it contains complex steps
// and most importantly, not all details are disclosed!
type BotActionHighLevel interface {
	HighLevelPlaceSign(targetPos define.CubePos, signBlock string, opt *SignBlockPlaceOption) (err error)
	HighLevelPlaceCommandBlock(option *PlaceCommandBlockOption, maxRetry int) error
	HighLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error
	HighLevelEnsureBotNearby(pos define.CubePos, threshold float32) error
	HighLevelRemoveSpecificBlockSideEffect(pos define.CubePos, wantAir bool, backupName string) (deferFunc func(), err error)
	HighLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error)
	HighLevelEnchantItem(slot uint8, enchants map[string]int32) (err error)
	HighLevelListenItemPicked(timeout time.Duration) (actionChan chan protocol.InventoryAction, cancel func(), err error)
	HighLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlots map[uint8]bool, maxRetriesTotal int) (targetSlotsGetInfo map[uint8]bool, err error)
	HighLevelSetContainerContent(pos define.CubePos, containerInfo map[uint8]*ContainerSlotItemStack) (err error)
	HighLevelGenContainer(pos define.CubePos, containerInfo map[uint8]*ContainerSlotItemStack, block string) (err error)
	HighLevelWriteBook(slotID uint8, pages []string) (err error)
	HighLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error)
	HighLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8) error
	HighLevelMakeItem(item *Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error
	HighLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error
}

type BotAction interface {
	SelectHotBar(slotID uint8) error
	SleepTick(ticks int)
	UseHotBarItem(slot uint8) (err error)
	UseHotBarItemOnBlock(blockPos define.CubePos, blockNEMCRuntimeID uint32, face int32, slot uint8) (err error)
	UseHotBarItemOnBlockWithBotOffset(blockPos define.CubePos, botOffset define.CubePos, blockNEMCRuntimeID uint32, face int32, slot uint8) (err error)
	// TapBlockUsingHotBarItem(blockPos define.CubePos, blockNEMCRuntimeID uint32, slotID uint8) (err error)
	MoveItemFromInventoryToEmptyContainerSlots(pos define.CubePos, blockNemcRtid uint32, blockName string, moveOperations map[uint8]uint8) error
	UseAnvil(pos define.CubePos, blockNemcRtid uint32, slot uint8, newName string) error
	DropItemFromHotBar(slot uint8) error
	MoveItemInsideHotBarOrInventory(sourceSlot, targetSlot, count uint8) error
	SetStructureBlockData(pos define.CubePos, settings *StructureBlockSettings)
}

type ItemLockPlace string

const (
	ItemLockPlaceInventory = ItemLockPlace("lock_in_inventory") //2
	ItemLockPlaceSlot      = ItemLockPlace("lock_in_slot")      //1
)

type ItemComponentsInGiveOrReplace struct {
	CanPlaceOn  []string      `json:"can_place_on,omitempty"`
	CanDestroy  []string      `json:"can_destroy,omitempty"`
	ItemLock    ItemLockPlace `json:"item_lock,omitempty"`
	KeepOnDeath bool          `json:"keep_on_death,omitempty"`
}

func (c *ItemComponentsInGiveOrReplace) IsEmpty() bool {
	return (len(c.CanPlaceOn) == 0) && (len(c.CanDestroy) == 0) && (c.ItemLock == "") && (!c.KeepOnDeath)
}

func (c *ItemComponentsInGiveOrReplace) ToString() string {
	if c == nil {
		return ""
	}
	components := map[string]any{}

	if len(c.CanPlaceOn) > 0 {
		components["minecraft:can_place_on"] = map[string][]string{"blocks": c.CanPlaceOn}
	}
	if len(c.CanDestroy) > 0 {
		components["minecraft:can_destroy"] = map[string][]string{"blocks": c.CanDestroy}
	}
	if c.ItemLock != "" {
		components["item_lock"] = map[string]string{"mode": string(c.ItemLock)}
	}
	if c.KeepOnDeath {
		components["keep_on_death"] = map[string]string{}
	}
	if len(components) == 0 {
		return ""
	}
	// will not fail
	bs, _ := json.Marshal(components)
	return string(bs)
}

type CommandHelper interface {
	// if in overworld, send in WO manner, otherwise send in websocket manner
	ConstructDimensionLimitedWOCommand(cmd string) CmdCannotGetResponse
	// if in overworld, Send() in WO manner, otherwise Send() in websocket manner
	ConstructDimensionLimitedGeneralCommand(cmd string) GeneralCommand
	ConstructGeneralCommand(cmd string) GeneralCommand
	// an uuid string but replaced by special chars
	GenAutoUnfilteredUUID() string
	ReplaceHotBarItemCmd(slotID int32, item string) CmdCanGetResponse
	ReplaceBotHotBarItemFullCmd(slotID int32, itemName string, count uint8, value int32, components *ItemComponentsInGiveOrReplace) CmdCanGetResponse
	ReplaceContainerBlockItemCmd(pos define.CubePos, slotID int32, item string) CmdCanGetResponse
	ReplaceContainerItemFullCmd(pos define.CubePos, slotID int32, itemName string, count uint8, value int32, components *ItemComponentsInGiveOrReplace) CmdCanGetResponse

	BackupStructureWithGivenNameCmd(start define.CubePos, size define.CubePos, name string) CmdCanGetResponse
	BackupStructureWithAutoNameCmd(start define.CubePos, size define.CubePos) (name string, cmd CmdCanGetResponse)
	RevertStructureWithGivenNameCmd(start define.CubePos, name string) CmdCanGetResponse

	SetBlockCmd(pos define.CubePos, blockString string) GeneralCommand
	SetBlockRelativeCmd(pos define.CubePos, blockString string) GeneralCommand
	FillBlocksWithRangeCmd(startPos define.CubePos, endPos define.CubePos, blockString string) GeneralCommand
	FillBlocksWithSizeCmd(startPos define.CubePos, size define.CubePos, blockString string) GeneralCommand
}

type BotActionComplex interface {
	CommandHelper
	BotAction
	BotActionHighLevel
}

type SignBlockText struct {
	HideGlowOutline   uint8
	IgnoreLighting    uint8
	PersistFormatting uint8
	SignTextColor     int32
	Text              string
	// TextOwner         string
}

func (t SignBlockText) Color() colorful.Color {
	R, G, B := uint8(t.SignTextColor>>16), uint8(t.SignTextColor>>8), uint8(t.SignTextColor)
	return colorful.Color{R: float64(R) / 256.0, G: float64(G) / 256.0, B: float64(B) / 256.0}
}

// data from PhoenixBuilder/fastbuilder/bdump/nbt_assigner/shared_pool.go
var colorToDyeColorName map[colorful.Color]string = map[colorful.Color]string{
	{R: 240 / 256.0, G: 240 / 256.0, B: 240 / 256.0}: "white_dye",      // 白色染料
	{R: 157 / 256.0, G: 151 / 256.0, B: 151 / 256.0}: "light_gray_dye", // 淡灰色染料
	{R: 71 / 256.0, G: 79 / 256.0, B: 82 / 256.0}:    "gray_dye",       // 灰色染料
	{R: 0 / 256.0, G: 0 / 256.0, B: 0 / 256.0}:       "",               // 黑色染料
	{R: 131 / 256.0, G: 84 / 256.0, B: 50 / 256.0}:   "brown_dye",      // 棕色染料
	{R: 176 / 256.0, G: 46 / 256.0, B: 38 / 256.0}:   "red_dye",        // 红色染料
	{R: 249 / 256.0, G: 128 / 256.0, B: 29 / 256.0}:  "orange_dye",     // 橙色染料
	{R: 254 / 256.0, G: 216 / 256.0, B: 61 / 256.0}:  "yellow_dye",     // 黄色染料
	{R: 128 / 256.0, G: 199 / 256.0, B: 31 / 256.0}:  "lime_dye",       // 黄绿色染料
	{R: 94 / 256.0, G: 124 / 256.0, B: 22 / 256.0}:   "green_dye",      // 绿色染料
	{R: 22 / 256.0, G: 156 / 256.0, B: 156 / 256.0}:  "cyan_dye",       // 青色染料
	{R: 58 / 256.0, G: 179 / 256.0, B: 218 / 256.0}:  "light_blue_dye", // 淡蓝色染料
	{R: 60 / 256.0, G: 68 / 256.0, B: 170 / 256.0}:   "blue_dye",       // 蓝色染料
	{R: 137 / 256.0, G: 50 / 256.0, B: 184 / 256.0}:  "purple_dye",     // 紫色染料
	{R: 199 / 256.0, G: 78 / 256.0, B: 189 / 256.0}:  "magenta_dye",    // 品红色染料
	{R: 243 / 256.0, G: 139 / 256.0, B: 170 / 256.0}: "pink_dye",       // 粉红色染料
}

func (t SignBlockText) GetDyeName() string {
	targetC := t.Color()
	bestColor := ""
	distance := math.Inf(1)
	for colorRGB, dyeName := range colorToDyeColorName {
		dis := colorRGB.DistanceLab(targetC)
		if dis == 0 {
			return dyeName
		}
		if dis < distance {
			distance = dis
			bestColor = dyeName
		}
	}
	return bestColor
}

// id HangingSign
type SignBlockPlaceOption struct {
	FrontText, BackText SignBlockText
	IsWaxed             uint8
}

func (opt *SignBlockPlaceOption) ToNBT() map[string]any {
	out := map[string]any{}
	mapstructure.Decode(opt, &out)
	return out
}

func NbtToSignBlock(nbt map[string]any) *SignBlockPlaceOption {
	if nbt["id"] != "Sign" && nbt["id"] != "HangingSign" {
		return nil
	}
	if _, found := nbt["IsWaxed"]; !found {
		text, ok := nbt["Text"].(string)
		if !ok {
			text = ""
		}
		lighting, _ := nbt["IgnoreLighting"].(uint8)
		return &SignBlockPlaceOption{
			IsWaxed: 0,
			FrontText: SignBlockText{
				HideGlowOutline:   0,
				Text:              text,
				IgnoreLighting:    lighting,
				PersistFormatting: 1,
				SignTextColor:     -16777216,
			},
			BackText: SignBlockText{
				HideGlowOutline:   0,
				Text:              text,
				IgnoreLighting:    lighting,
				PersistFormatting: 1,
				SignTextColor:     -16777216,
			},
		}
	}
	opt := &SignBlockPlaceOption{}
	if err := mapstructure.Decode(nbt, &opt); err != nil {
		return nil
	}
	return opt
}
