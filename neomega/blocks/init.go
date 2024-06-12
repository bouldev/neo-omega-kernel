package blocks

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

var DefaultAnyToNemcConvertor = NewToNEMCConverter()
var SchemToNemcConvertor = NewToNEMCConverter()
var schematicToNemcConvertor = NewToNEMCConverter()

const UNKNOWN_RUNTIME = uint32(0xFFFFFFFF)

var NEMC_BLOCK_VERSION = int32(0)
var NEMC_AIR_RUNTIMEID = uint32(0)
var AIR_RUNTIMEID = uint32(0)
var AIR_BLOCK = &NEMCBlock{
	Name:  "air",
	Value: 0,
	Props: make(Props, 0),
}

var nemcBlocks = []NEMCBlock{}

//go:embed "nemc.br"
var nemcBlockInfoBytes []byte

func initNemcBlocks() {
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(nemcBlockInfoBytes)))
	if err != nil {
		panic(err)
	}
	LoadNemcBlocksToGlobal(string(dataBytes))
}

//go:embed "bedrock_java_to_translate.br"
var toNemcDataLoadBedrockJavaTranslateInfo []byte

//go:embed "schem_to_translate.br"
var toNemcDataLoadSchemTranslateInfo []byte

var quickSchematicMapping [256][16]uint32

func initSchematicBlockCheck() {
	quickSchematicMapping = [256][16]uint32{}
	for i := 0; i < 256; i++ {
		blockName := schematicBlockStrings[i]
		_, found := DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), 0)
		if !found {
			panic(fmt.Errorf("schematic %v 0 not found", blockName))
		}
	}
	for blockI := 0; blockI < 256; blockI++ {
		blockName := schematicBlockStrings[blockI]
		// if blockName == "stone_slab" {
		// 	fmt.Println("slab")
		// }
		for dataI := 0; dataI < 16; dataI++ {
			rtid, found := schematicToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), int16(dataI))
			if !found || rtid == AIR_RUNTIMEID {
				rtid, _ = schematicToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), 0)
			}
			quickSchematicMapping[blockI][dataI] = rtid
		}
	}
	schematicToNemcConvertor = nil
}

func initToNemcDataLoadBedrockJava() {
	WriteNemcInfoToConvertor(DefaultAnyToNemcConvertor)
	WriteNemcInfoToConvertor(SchemToNemcConvertor)
	WriteNemcInfoToConvertor(schematicToNemcConvertor)
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadBedrockJavaTranslateInfo)))
	if err != nil {
		panic(err)
	}
	records := string(dataBytes)
	DefaultAnyToNemcConvertor.LoadConvertRecords(records, false, true)
	SchemToNemcConvertor.LoadConvertRecords(records, false, true)
	schematicToNemcConvertor.LoadConvertRecords(records, false, true)
}

func initToNemcDataLoadSchem() {
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadSchemTranslateInfo)))
	if err != nil {
		panic(err)
	}
	records := string(dataBytes)
	DefaultAnyToNemcConvertor.LoadConvertRecords(records, false, false)
	SchemToNemcConvertor.LoadConvertRecords(records, true, true)
}

func init() {
	initPreGeneratePropValInt32()
	initPreGeneratePropValInt32ForSearch()
	initNemcBlocks()
	initToNemcDataLoadBedrockJava()
	initToNemcDataLoadSchem()
	initSchematicBlockCheck()
}
