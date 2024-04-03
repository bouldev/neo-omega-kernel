package blocks

import "fmt"

func RuntimeIDToBlock(runtimeID uint32) (block *NEMCBlock, found bool) {
	if int(runtimeID) >= len(nemcBlocks) {
		return AIR_BLOCK, false
	}
	return &nemcBlocks[runtimeID], true
}

func LegacyBlockToRuntimeID(name string, data uint16) (runtimeID uint32, found bool) {
	return DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(name), int16(data))
}

func RuntimeIDToState(runtimeID uint32) (baseName string, properties map[string]any, found bool) {
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air", nil, false
	}
	return block.Name, block.Props.ToNBT(), true
}

// coral_block ["coral_color":"yellow", "dead_bit":false] true
func RuntimeIDToBlockNameWithStateStr(runtimeID uint32) (blockNameWithState string, found bool) {
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air []", false
	}
	return block.Name + " " + block.Props.BedrockString(true), true
}

func RuntimeIDToBlockNameAndStateStr(runtimeID uint32) (blockName, blockState string, found bool) {
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air", "[]", false
	}
	return block.Name, block.Props.BedrockString(true), true
}

func BlockNameAndStateToRuntimeID(name string, properties map[string]any) (runtimeID uint32, found bool) {
	props, err := PropsForSearchFromNbt(properties)
	if err != nil {
		// legacy capability
		fmt.Println(err)
		return uint32(AIR_RUNTIMEID), false
	}
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(BlockNameForSearch(name), props)
	return rtid, found
}

func BlockNameAndStateStrToRuntimeID(name string, stateStr string) (runtimeID uint32, found bool) {
	props, err := PropsForSearchFromStr(stateStr)
	if err != nil {
		// legacy capability
		fmt.Println(err)
		return uint32(AIR_RUNTIMEID), false
	}
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(BlockNameForSearch(name), props)
	return rtid, found
}

func BlockStrToRuntimeID(blockNameWithOrWithoutState string) (runtimeID uint32, found bool) {
	blockName, blockProps := ConvertStringToBlockNameAndPropsForSearch(blockNameWithOrWithoutState)
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(blockName, blockProps)
	return rtid, found
}

func SchemBlockStrToRuntimeID(blockNameWithOrWithoutState string) (runtimeID uint32, found bool) {
	blockName, blockProps := ConvertStringToBlockNameAndPropsForSearch(blockNameWithOrWithoutState)
	rtid, _, found := SchemToNemcConvertor.TryBestSearchByState(blockName, blockProps)
	return rtid, found
}

// func SchematicToRuntimeID(blockIdx uint8, data uint8) (runtimeID uint32, found bool) {
// 	if int(blockIdx) >= len(SchematicBlockStrings) {
// 		return AIR_RUNTIMEID, false
// 	}
// 	blockName := SchematicBlockStrings[blockIdx]
// 	return LegacyBlockToRuntimeID(blockName, uint16(data))
// }
