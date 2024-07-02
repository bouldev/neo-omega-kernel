package bot_action

import (
	"fmt"
	"runtime/debug"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
)

type MaintainedGameInfo struct {
	currentSlot          uint8
	currentItemRequestID int32
	inventoryContents    *sync_wrapper.SyncKVMap[uint32, *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]]
}

func NewMaintainedGameInfo(listener neomega.ReactCore) *MaintainedGameInfo {
	info := &MaintainedGameInfo{
		currentSlot:          0,
		currentItemRequestID: 1,
		// inventoryContents:    sync_wrapper.NewSyncKVMap[uint32, *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]](),
	}
	listener.SetAnyPacketCallBack(info.UpdateFromPacket, false)
	return info
}

func (o *MaintainedGameInfo) NextItemRequestID() int32 {
	o.currentItemRequestID -= 2
	return o.currentItemRequestID
}

func (e *MaintainedGameInfo) ensureInventoryWindow(windowID uint32) *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance] {
	if e.inventoryContents == nil {
		e.inventoryContents = sync_wrapper.NewSyncKVMap[uint32, *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]]()
	}
	slots, found := e.inventoryContents.Get(windowID)
	if found {
		return slots
	}
	empty := sync_wrapper.NewSyncKVMap[uint8, *protocol.ItemInstance]()
	e.inventoryContents.Set(windowID, empty)
	return slots
}

func (e *MaintainedGameInfo) writeInventoryWindow(windowID uint32, slots *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]) {
	if e.inventoryContents == nil {
		e.inventoryContents = sync_wrapper.NewSyncKVMap[uint32, *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance]]()
	}
	e.inventoryContents.Set(windowID, slots)
}

func (e *MaintainedGameInfo) writeInventorySlot(windowID uint32, slotID uint8, slot *protocol.ItemInstance) {
	e.ensureInventoryWindow(windowID)
	slots, _ := e.inventoryContents.Get(windowID)
	slots.Set(slotID, slot)
}

func (e *MaintainedGameInfo) GetInventoryWindow(windowID uint32) (slots *sync_wrapper.SyncKVMap[uint8, *protocol.ItemInstance], found bool) {
	if e.inventoryContents == nil {
		return nil, false
	}
	return e.inventoryContents.Get(windowID)
}

func (e *MaintainedGameInfo) GetInventorySlot(windowID uint32, slotID uint8) (slot *protocol.ItemInstance, found bool) {
	if slots, found := e.GetInventoryWindow(windowID); found {
		return slots.Get(slotID)
	}
	return nil, false
}

func (uq *MaintainedGameInfo) UpdateFromPacket(pk packet.Packet) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("Maintained Game Info Update Error: ", r)
			debug.PrintStack()
		}
	}()
	switch p := pk.(type) {
	case *packet.InventoryContent:
		for key, _value := range p.Content {
			value := _value
			if value.Stack.ItemType.NetworkID != -1 {
				// fmt.Println("InventoryContent", p.WindowID, uint8(key), value)
				uq.writeInventorySlot(p.WindowID, uint8(key), &value)
			}
		}
	case *packet.InventorySlot:
		// fmt.Println("InventorySlot", p)
		uq.writeInventorySlot(p.WindowID, uint8(p.Slot), &p.NewItem)
	case *packet.InventoryTransaction:
		// fmt.Println("InventoryTransaction", p)
		for _, _value := range p.Actions {
			value := _value
			if value.SourceType != protocol.InventoryActionSourceContainer {
				continue
			}
			// fmt.Println("InventoryTransactionI", uint32(value.WindowID), uint8(value.InventorySlot), value.NewItem)
			uq.writeInventorySlot(uint32(value.WindowID), uint8(value.InventorySlot), &value.NewItem)
		}
	}
}
