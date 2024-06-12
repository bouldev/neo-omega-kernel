package blocks

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

type NEMCBlock struct {
	Name           string
	Props          Props
	Value          uint16
	StateSNBT      string
	PropsForSearch *PropsForSearch
}

func LoadNemcBlocksToGlobal(dataBytes string) {
	reader := bufio.NewReader(bytes.NewBufferString(dataBytes))
	{
		version, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if !strings.HasPrefix(version, "VERSION:") {
			panic(fmt.Errorf("expect VERSION:, get %v", version))
		} else {
			versionInt, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(version, "VERSION:")))
			NEMC_BLOCK_VERSION = int32(versionInt)
			if NEMC_BLOCK_VERSION == 0 {
				panic(fmt.Errorf("cannot get nemc block version: %v", version))
			} else {
				// fmt.Printf("nemc block version: %v\n", NEMC_BLOCK_VERSION)
			}
		}
	}
	numBlocks := 0
	{
		count, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if !strings.HasPrefix(count, "COUNTS:") {
			panic(fmt.Errorf("expect COUNTS:, get %v", count))
		} else {
			numBlocks, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(count, "COUNTS:")))
			if numBlocks == 0 {
				panic(fmt.Errorf("cannot get nemc block count: %v", count))
			} else {
				// fmt.Printf("nemc block count: %v\n", numBlocks)
			}
		}
	}
	nemcBlocks = make([]NEMCBlock, numBlocks)
	for runtimeID := int32(0); runtimeID < int32(numBlocks); runtimeID++ {
		blockName, err := reader.ReadString(' ')
		blockName = strings.TrimSpace(blockName)
		blockName = strings.TrimPrefix(blockName, "minecraft:")
		if err != nil {
			panic(err)
		}
		blockValStr, err := reader.ReadString(' ')
		blockValStr = strings.TrimSpace(blockValStr)
		if err != nil {
			panic(err)
		}
		stateSNBT, err := reader.ReadString('\n')
		stateSNBT = strings.TrimSpace(stateSNBT)
		// TODO: check
		stateSNBT = strings.ReplaceAll(stateSNBT, "minecraft:", "")
		if err != nil {
			panic(err)
		}
		// fmt.Printf("%v,%v,%v\n", blockName, blockValStr, stateSNBT)
		blockVal, err := strconv.Atoi(blockValStr)
		if err != nil {
			panic(err)
		}
		propsForSearch, err := PropsForSearchFromStr(stateSNBT)
		if err != nil {
			panic(err)
		}
		nemcBlocks[runtimeID] = NEMCBlock{
			Name:           blockName,
			Props:          PropsFromSNBT(stateSNBT),
			Value:          uint16(blockVal),
			StateSNBT:      stateSNBT,
			PropsForSearch: propsForSearch,
		}
		if nemcBlocks[runtimeID].Props.SNBTString() != stateSNBT {
			panic(fmt.Errorf("snbt error: %v!=%v", stateSNBT, nemcBlocks[runtimeID].Props.SNBTString()))
		}
		if blockName == "air" {
			NEMC_AIR_RUNTIMEID = uint32(runtimeID)
			AIR_RUNTIMEID = NEMC_AIR_RUNTIMEID
		}
	}
	if NEMC_AIR_RUNTIMEID == 0 || AIR_RUNTIMEID == 0 {
		panic("cannot found air runtime id")
	}
}

func WriteNemcInfoToConvertor(convertor *ToNEMCConverter) {
	for runtimeID, nemcBlock := range nemcBlocks {
		if exist, err := convertor.AddAnchorByLegacyValue(BlockNameForSearch(nemcBlock.Name), int16(nemcBlock.Value), uint32(runtimeID)); err != nil {
			panic(err)
		} else if exist {
			panic("should not happen")
		}
		if exist, err := convertor.AddAnchorByState(BlockNameForSearch(nemcBlock.Name), nemcBlock.PropsForSearch, uint32(runtimeID), false); err != nil {
			panic(err)
		} else if exist {
			panic("should not happen")
		}
	}
}
