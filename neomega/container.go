package neomega

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OmineDev/neomega-core/neomega/blocks"

	"github.com/mitchellh/mapstructure"
)

type ItemTypeDescription string

const (
	SimpleNonBlockItem     = ItemTypeDescription("simple non-block item")
	SimpleBlockItem        = ItemTypeDescription("simple block item")
	NeedHotBarNonBlockItem = ItemTypeDescription("need hotbar non-block item")
	NeedHotBarBlockItem    = ItemTypeDescription("need hotbar block item")

	NeedAuxBlockNonBlockItem  = ItemTypeDescription("need aux block non-block item")
	NeedAuxBlockBlockItem     = ItemTypeDescription("need aux block block item")
	ComplexBlockItemContainer = ItemTypeDescription("complex block item: container")
	ComplexBlockItemUnknown   = ItemTypeDescription("complex block item: unknown")

	ComplexBlockContainer = "container"
	ComplexBlockUnknown   = "unknown"

	KnownItemWrittenBook  = "written_book"
	KnownItemWritableBook = "writable_book"
)

// IsBlock: if item can be put as a block, it could have RelatedBlockStateString
func (d ItemTypeDescription) IsBlock() bool { return strings.Contains(string(d), " block item") }

// IsSimple: if can be fully get by replace/give, etc., return true
func (d ItemTypeDescription) IsSimple() bool { return strings.HasPrefix(string(d), "simple") }

// NeedHotbar: item that requires putted in hot bar when generating, usually enchant
func (d ItemTypeDescription) NeedHotbar() bool {
	return strings.Contains(string(d), "need hotbar") || d.NeedAuxBlock()
}

// KnownItem: item that is known to have specific operations in generating, e.g. book
func (d ItemTypeDescription) KnownItem() string {
	if d.IsBlock() {
		return ""
	} else if strings.ContainsAny(string(d), ":") {
		return strings.TrimSpace(strings.Split(string(d), ":")[1])
	}
	return ""
}

// NeedAuxItem: item need aux block when generating, e.g. rename
func (d ItemTypeDescription) NeedAuxBlock() bool {
	return strings.Contains(string(d), "need aux block") || d.IsComplexBlock()
}

// IsComplexBlock: when item put as a block, the block contains certain info
func (d ItemTypeDescription) IsComplexBlock() bool {
	return strings.HasPrefix(string(d), "complex block item")
}

func (d ItemTypeDescription) ComplexBlockSubType() string {
	ss := strings.Split(string(d), ":")
	if len(ss) == 0 {
		return "unknown"
	} else {
		subType := strings.TrimSpace(ss[1])
		if subType == ComplexBlockContainer {
			return ComplexBlockContainer
		}
		return ComplexBlockUnknown
	}
}

// to avoid using interface or "any" we enum all complex block supported here
type ComplexBlockData struct {
	// is a container
	Container map[uint8]*ContainerSlotItemStack `json:"container,omitempty"`
	// unknown (describe as a nbt)
	Unknown map[string]any `json:"unknown_nbt,omitempty"`
}

func (d *ComplexBlockData) String(indent string) string {
	if d.Container != nil {
		out := "容器内容:"
		for _, slot := range d.Container {
			out += "\n" + indent + "    " + slot.String(indent+"    ")
		}
		return out
	}
	return ""
}

// to avoid using interface or "any" we enum all known item supported here
type KnownItemData struct {
	// is a book
	Pages []string `json:"pages,omitempty"`
	// author if is a written book
	BookAuthor string `json:"book_author,omitempty"`
	// book name
	BookName string `json:"book_name,omitempty"`
	// unknown (describe as a nbt)
	Unknown map[string]any `json:"known_nbt,omitempty"`
}

func (d *KnownItemData) String(indent string) string {
	out := fmt.Sprintf("书名: %v 作者: %v 页数: %v", d.BookName, d.BookAuthor, len(d.Pages))
	return out
}

type Item struct {
	// Basic Item: item can be given by replace/give, etc.
	Name  string `json:"name"`
	Value int16  `json:"val"`
	// can place on, can break, lock, etc. safe to be nil
	Components *ItemComponentsInGiveOrReplace `json:"components,omitempty"`
	// if item can be put as a block, it could have RelatedBlockStateString, safe to be empty
	IsBlock                        bool   `json:"is_block"`
	RelatedBlockBedrockStateString string `json:"block_bedrock_state_string,omitempty"`
	// END Basic Item

	// KnownItem
	RelatedKnownItemData *KnownItemData `json:"known_item_data,omitempty"`
	// End Known Item

	// NeedHotBarItem: item that requires putted in hot bar when generating, usually enchant, safe to be empty
	Enchants map[string]int32 `json:"enchants,omitempty"`
	// End NeedHotBarItem

	// NeedAuxItem: item need aux block when generating, e.g. rename, safe to be empty
	DisplayName string `json:"display_name,omitempty"`
	// End NeedAuxItem

	// Complex Block: when item put as a block, the block contains certain info, safe to be empty
	RelateComplexBlockData *ComplexBlockData `json:"complex_block_data,omitempty"`
	// End Complex Block
}

func (i *Item) String(indent string) string {
	out := fmt.Sprintf("%v[特殊值=%v]", i.Name, i.Value)
	if i.Components != nil {
		if len(i.Components.CanPlaceOn) > 1 {
			out += "\n" + indent + "   |可被放置于: "
			for _, name := range i.Components.CanPlaceOn {
				out += " " + name
			}
		}
		if len(i.Components.CanDestroy) > 1 {
			out += "\n" + indent + "   |可破坏: "
			for _, name := range i.Components.CanDestroy {
				out += " " + name
			}
		}
		if i.Components.ItemLock != "" || i.Components.KeepOnDeath {
			out += "\n" + indent + "   |"
			if i.Components.ItemLock != "" {
				out += "锁定在物品栏"
			}
			if i.Components.KeepOnDeath {
				if i.Components.ItemLock != "" {
					out += "&"
				}
				out += "死亡时保留"
			}
		}
	}
	if i.RelatedKnownItemData != nil {
		out += "\n" + indent + "   |信息: " + i.RelatedKnownItemData.String(indent+"    ")
	}
	if len(i.Enchants) > 0 {
		out += "\n" + indent + "   |附魔: "
		for enchant, level := range i.Enchants {
			out += fmt.Sprintf("%v级魔咒:%v; ", level, enchant)
		}
	}
	if i.DisplayName != "" {
		out += "\n" + indent + "   |被命名为: " + i.DisplayName
	}
	if i.RelateComplexBlockData != nil && i.RelateComplexBlockData.Container != nil {
		out += "\n" + indent + "   |" + i.RelateComplexBlockData.String(indent+"    ")
	}
	return out
}

func (i *Item) GetJsonString() string {
	bs, _ := json.Marshal(i)
	return string(bs)
}

func (i *Item) GetTypeDescription() ItemTypeDescription {
	if i.IsBlock {
		if i.RelateComplexBlockData != nil {
			if i.RelateComplexBlockData.Container != nil {
				return ComplexBlockItemContainer
			}
			return ComplexBlockItemUnknown
		}
		if i.DisplayName != "" {
			return NeedAuxBlockBlockItem
		} else {
			if len(i.Enchants) > 0 {
				return NeedHotBarBlockItem
			} else {
				return SimpleBlockItem
			}
		}
	} else {
		knownItem := ""
		shortName := strings.TrimPrefix(i.Name, "minecraft:")
		knownItem = map[string]string{
			"written_book":  ": " + KnownItemWrittenBook,
			"writable_book": ": " + KnownItemWritableBook,
		}[shortName]
		if i.DisplayName != "" {
			return NeedAuxBlockNonBlockItem + ItemTypeDescription(knownItem)
		} else {
			if len(i.Enchants) > 0 || knownItem != "" {
				return NeedHotBarNonBlockItem + ItemTypeDescription(knownItem)
			} else {
				return SimpleNonBlockItem + ItemTypeDescription(knownItem)
			}
		}
	}
}

type ContainerSlotItemStack struct {
	Item  *Item `json:"item"`
	Count uint8 `json:"count"`
}

func (s *ContainerSlotItemStack) String(indent string) string {
	out := fmt.Sprintf("%v个 %v", s.Count, s.Item.String(indent))
	return out
}

func GenContainerItemsInfoFromItemsNbt(itemsNbt []any) (container map[uint8]*ContainerSlotItemStack, err error) {
	container = map[uint8]*ContainerSlotItemStack{}
	for _, itemNbt := range itemsNbt {
		item, ok := itemNbt.(map[string]any)
		if !ok {
			err = fmt.Errorf("fail to decode item nbt: %v", itemNbt)
			continue
		}
		itemNbt := &containerSlotItemNBT{}
		if err = mapstructure.Decode(item, &itemNbt); err != nil {
			err = fmt.Errorf("fail to decode item nbt: %v, %v", itemNbt, err)
			continue
		}
		slot, itemStack := itemNbt.toContainerSlotItemStack()
		container[slot] = itemStack
	}
	return container, err
}

// GlowItemFrame, ItemFrame ["item"]
func GenItemInfoFromItemFrameNBT(itemNbt any) (*Item, error) {
	item, ok := itemNbt.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fail to decode item nbt: %v", itemNbt)
	}
	ir := &containerSlotItemNBT{}
	if err := mapstructure.Decode(item, &ir); err != nil {
		return nil, fmt.Errorf("fail to decode item nbt: %v, %v", itemNbt, err)
	}
	_, itemStack := ir.toContainerSlotItemStack()

	return itemStack.Item, nil
}

type containerSlotItemNBT struct {
	Slot       uint8
	Count      uint8
	Name       string
	Damage     int16 // if is a non-block item, use this as value
	CanDestroy []string
	CanPlaceOn []string
	Tag        struct {
		Display struct {
			Name string
		} `mapstructure:"display"`
		Enchant []struct {
			ID    int32 `mapstructure:"id"`
			Level int32 `mapstructure:"lvl"`
		} `mapstructure:"ench"`
		KeepOnDeath uint8 `mapstructure:"minecraft:keep_on_death"`
		ItemLock    uint8 `mapstructure:"minecraft:item_lock"`
		// if is chest/container
		Items []any
		// if is a book
		Pages []struct {
			Text string `mapstructure:"text"`
		} `mapstructure:"pages"`
		// if is written book
		Title  string `mapstructure:"title"`
		Author string `mapstructure:"author"`
	} `mapstructure:"tag"`
	Block struct {
		Name  string         `mapstructure:"name"`
		State map[string]any `mapstructure:"states"`
		Value int16          `mapstructure:"val"` // if is a block item, use this as value
	} `mapstructure:"Block"`
}

func (i *containerSlotItemNBT) toContainerSlotItemStack() (slot uint8, stack *ContainerSlotItemStack) {
	slot = i.Slot
	stack = &ContainerSlotItemStack{
		Count: i.Count,
		Item: &Item{
			Name:                           i.Name,
			Value:                          i.Damage,
			IsBlock:                        false,
			RelatedBlockBedrockStateString: "",
			Components: &ItemComponentsInGiveOrReplace{
				CanPlaceOn: i.CanPlaceOn,
				CanDestroy: i.CanDestroy,
				ItemLock: map[uint8]ItemLockPlace{
					1: ItemLockPlaceSlot,
					2: ItemLockPlaceInventory,
				}[i.Tag.ItemLock],
				KeepOnDeath: i.Tag.KeepOnDeath == 1,
			},
			Enchants:    make(map[string]int32),
			DisplayName: i.Tag.Display.Name,
		},
	}
	if stack.Item.Components.IsEmpty() {
		stack.Item.Components = nil
	}
	// if strings.Contains(stack.Item.Name, "box") {
	// 	fmt.Println("box")
	// }
	if i.Block.Name != "" {
		stack.Item.IsBlock = true
		stack.Item.Value = i.Block.Value
		states := map[string]any{}
		if len(i.Block.State) > 0 {
			states = i.Block.State
		}
		rtid, found := blocks.BlockNameAndStateToRuntimeID(i.Block.Name, states)
		if !found {
			fmt.Printf("unknown nested block: %v %v", i.Block.Name, states)
			stack.Item.IsBlock = false
		} else {
			blockNameNoStates, statesStr, _ := blocks.RuntimeIDToBlockNameAndStateStr(rtid)
			i.Block.Name = blockNameNoStates
			stack.Item.RelatedBlockBedrockStateString = statesStr
			stack.Item.Name = blockNameNoStates
		}
		// if len(i.Block.State) > 0 {
		// 	stack.Item.RelatedBlockBedrockStateString = "["
		// 	props := make([]string, 0)
		// 	for k, v := range i.Block.State {
		// 		if bv, ok := v.(uint8); ok {
		// 			if bv == 0 {
		// 				v = "false"
		// 			} else {
		// 				v = "true"
		// 			}
		// 		}
		// 		if sv, ok := v.(string); ok {
		// 			v = fmt.Sprintf("\"%v\"", sv)
		// 		}
		// 		props = append(props, fmt.Sprintf("\"%v\": %v", k, v))
		// 	}
		// 	stateStr := strings.Join(props, ", ")
		// 	stack.Item.RelatedBlockBedrockStateString = fmt.Sprintf("[%v]", stateStr)
		// }
	}
	for _, enchant := range i.Tag.Enchant {
		stack.Item.Enchants[fmt.Sprintf("%v", enchant.ID)] = enchant.Level
	}
	// a chest
	if len(i.Tag.Items) > 0 && stack.Item.IsBlock {
		if stack.Item.RelateComplexBlockData == nil {
			stack.Item.RelateComplexBlockData = &ComplexBlockData{}
		}
		stack.Item.RelateComplexBlockData.Container, _ = GenContainerItemsInfoFromItemsNbt(i.Tag.Items)
	}
	shortName := strings.TrimPrefix(i.Name, "minecraft:")
	// a book/written_book
	if shortName == "writable_book" || shortName == "written_book" {
		if stack.Item.RelatedKnownItemData == nil {
			stack.Item.RelatedKnownItemData = &KnownItemData{}
		}
		stack.Item.RelatedKnownItemData.Pages = make([]string, len(i.Tag.Pages))
		for i, data := range i.Tag.Pages {
			stack.Item.RelatedKnownItemData.Pages[i] = data.Text
		}
	}

	if shortName == "written_book" {
		if stack.Item.RelatedKnownItemData == nil {
			stack.Item.RelatedKnownItemData = &KnownItemData{}
		}
		stack.Item.RelatedKnownItemData.BookAuthor = i.Tag.Author
		stack.Item.RelatedKnownItemData.BookName = i.Tag.Title
	}
	return
}
