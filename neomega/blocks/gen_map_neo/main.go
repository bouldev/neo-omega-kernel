package main

import (
	"fmt"
	"neo-omega-kernel/neomega/blocks"
	"neo-omega-kernel/neomega/blocks/gen_map_neo/step0_nemc_blocks_liliya"
	"neo-omega-kernel/neomega/blocks/gen_map_neo/step1_pack_nemc_block_info"
	"neo-omega-kernel/neomega/blocks/gen_map_neo/step2_add_standard_mc_converts"
	"neo-omega-kernel/neomega/blocks/gen_map_neo/step3_add_schem_mapping"
	"os"
)

func main() {
	parsedBlocks := step0_nemc_blocks_liliya.GetParsedBlock("./data/block_palette_2.12.json")
	rawString, compressed := step1_pack_nemc_block_info.PackInfo(parsedBlocks)
	if err := os.WriteFile("nemc.br", compressed, 0755); err != nil {
		panic(err)
	}
	blocks.LoadNemcBlocksToGlobal(rawString)
	snbtInOut := step2_add_standard_mc_converts.ReadSnbtFile("./data/snbt_convert.txt")
	records := step2_add_standard_mc_converts.GenMCToNemcTranslateRecords(snbtInOut)
	rawRecords, compressed := step2_add_standard_mc_converts.PackConvertRecord(records)
	if err := os.WriteFile("bedrock_java_to_translate.br", compressed, 0755); err != nil {
		panic(err)
	}
	convertor := blocks.NewToNEMCConverter()
	blocks.WriteNemcInfoToConvertor(convertor)
	convertor.LoadConvertRecords(rawRecords, false, true)
	testBlock := `grass`
	blockName, blockProps := blocks.ConvertStringToBlockNameAndPropsForSearch(testBlock)
	rtid, _, ok := convertor.TryBestSearchByState(blockName, blockProps)
	if !ok {
		panic("test fail")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rawSchemData, err := os.ReadFile("./data/block_1_18_java_to_bedrock.json")
	if err != nil {
		panic(err)
	}
	rawRecords, compressed = step3_add_schem_mapping.GenSchemConvertRecord(rawSchemData, convertor)
	if err := os.WriteFile("schem_to_translate.br", compressed, 0755); err != nil {
		panic(err)
	}
	rtid, _, ok = convertor.TryBestSearchByState(blockName, blockProps)
	if !ok {
		panic("test fail")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
}
