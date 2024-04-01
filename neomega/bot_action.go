package neomega

import (
	"encoding/json"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega/chunks"
	"neo-omega-kernel/neomega/chunks/define"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
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
	HighLevelPlaceSign(targetPos define.CubePos, text string, lighting bool, signBlock string) (err error)
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
	// TapBlockUsingHotBarItem(blockPos define.CubePos, blockNEMCRuntimeID uint32, slotID uint8) (err error)
	MoveItemFromInventoryToEmptyContainerSlots(pos define.CubePos, blockNemcRtid uint32, blockName string, moveOperations map[uint8]uint8) error
	UseAnvil(pos define.CubePos, slot uint8, newName string) error
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
