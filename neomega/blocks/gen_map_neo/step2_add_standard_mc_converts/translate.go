package step2_add_standard_mc_converts

import (
	"fmt"
	"neo-omega-kernel/neomega/blocks"
	"strconv"
	"strings"
)

type ConvertRecord struct {
	Name      string `json:"name"`
	SNBTState string `json:"states"`
	RTID      uint32 `json:"rtid"`
}

func TryAddConvert(inBlockName, inBlockState, outBlockName, outBlockState string, converter *blocks.ToNEMCConverter, mustMatch bool) (record *ConvertRecord, ok bool, notMatched bool) {
	// first find target runtime id
	outBlockNameForSearch := blocks.BlockNameForSearch(outBlockName)
	outBlockStateForSearch, err := blocks.PropsForSearchFromStr(outBlockState)
	if err != nil {
		panic(err)
	}
	rtid := blocks.UNKNOWN_RUNTIME
	found := false

	if strings.HasPrefix(outBlockState, "block_data=") {
		outBlockState = strings.TrimPrefix(outBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(outBlockState)
		rtid, found = converter.PreciseMatchByLegacyValue(outBlockNameForSearch, int16(blockVal))
		if !found {
			if !mustMatch {
				return nil, false, true
			}
			rtid, found = converter.TryBestSearchByLegacyValue(outBlockNameForSearch, int16(blockVal))
			if !found {
				panic(fmt.Sprintf("not found! %v %v", outBlockNameForSearch, blockVal))
			} else {
				targetBlock, _ := blocks.RuntimeIDToBlockNameWithStateStr(rtid)
				fmt.Printf("fuzzy block data: %v %v -> %v\n", outBlockName, blockVal, targetBlock)
			}
		}
		if rtid == uint32(blocks.AIR_RUNTIMEID) {
			return nil, true, false
		}
	} else {
		rtid, found = converter.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
		if !found {
			if !mustMatch {
				return nil, false, true
			}
			converter.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
			panic(fmt.Sprintf("not found! %v %v", outBlockNameForSearch, outBlockStateForSearch))
		}
		if rtid == uint32(blocks.AIR_RUNTIMEID) {
			return nil, true, false
		}
	}

	inBlockNameForSearch := blocks.BlockNameForSearch(inBlockName)
	inBlockStateForSearch, err := blocks.PropsForSearchFromStr(inBlockState)
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(inBlockState, "block_data=") {
		inBlockState = strings.TrimPrefix(inBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(inBlockState)
		if existed, err := converter.AddAnchorByLegacyValue(inBlockNameForSearch, int16(blockVal), rtid); err != nil {
			// fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, blockVal, rtid, outBlockStateForSearch.InPreciseSNBT())
			return nil, false, false
		} else if !existed {
			return &ConvertRecord{
				inBlockNameForSearch.BaseName(),
				fmt.Sprintf("%v", blockVal),
				rtid,
			}, true, false
		} else {
			return nil, true, false
		}
		// fmt.Printf("%v %v -> %v\n", inBlockNameForSearch, blockVal, rtid)
	} else {
		if existed, err := converter.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, false); err != nil {
			// fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, inBlockStateForSearch.InPreciseSNBT(), rtid, outBlockStateForSearch.InPreciseSNBT())
			return nil, false, false
		} else if !existed {
			return &ConvertRecord{
				inBlockNameForSearch.BaseName(),
				inBlockStateForSearch.InPreciseSNBT(),
				rtid,
			}, true, false
		} else {
			return nil, true, false
		}
	}

}
