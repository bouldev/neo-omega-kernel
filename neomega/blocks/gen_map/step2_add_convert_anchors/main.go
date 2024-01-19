package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"neo-omega-kernel/neomega/blocks"
	"os"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

var translated = 0
var failed = 0
var redundant = 0
var translateRecord = []struct {
	Name      string
	SNBTState string
	RTID      uint32
}{}

func main() {
	fp, err := os.Open("snbt_convert.txt")
	if err != nil {
		panic(err)
	}
	dataBytes, err := io.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewBuffer(dataBytes))

	inBlockSNBT := ""

	for {
		cmd, err := reader.ReadString(':')
		if err != nil {
			break
		}
		cmd = strings.ReplaceAll(cmd, ":", " ")
		cmd = strings.TrimSpace(cmd)
		reader.ReadString(' ')
		snbt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		snbt = strings.TrimSpace(snbt)
		if cmd == "in" {
			inBlockSNBT = snbt
		}
		if cmd == "out" {
			HandleInOutSNBT(inBlockSNBT, snbt)
		}
	}
	fmt.Printf("translated %v failed %v redundant %v\n", translated, failed, redundant)
	// fmt.Println(translateRecord)
	textRecord := ""
	for _, record := range translateRecord {
		textRecord += fmt.Sprintf("%v\n%v\n%v\n", record.Name, record.SNBTState, record.RTID)
	}
	os.WriteFile("../bedrock_java_to_translate.txt", []byte(textRecord), 0755)
	outFp, _ := os.OpenFile("../bedrock_java_to_translate.br", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	brotliWriter := brotli.NewWriter(outFp)
	brotliWriter.Write([]byte(textRecord))
	if err := brotliWriter.Close(); err != nil {
		panic(err)
	}
	if err := outFp.Close(); err != nil {
		panic(err)
	}
}

func HandleInOutSNBT(inSNBT, outSNBT string) {
	outSS := strings.Split(outSNBT, "[")
	outBlockName, outBlockState := outSS[0], ""
	if len(outSS) > 1 {
		outBlockState = outSS[1]
	}
	outBlockState = strings.TrimSuffix(outBlockState, "]")

	if outBlockName == "minecraft:cherry_sign" {
		// fix
		outBlockName = "minecraft:standing_sign"
		outBlockState = strings.ReplaceAll(outBlockState, "rotation", "ground_sign_direction")
	} else if outBlockName == "minecraft:cherry_wall_sign" {
		outBlockName = "minecraft:wall_sign"
		outBlockState = strings.ReplaceAll(outBlockState, "facing", "facing_direction")
		// outBlockState = strings.ReplaceAll(outBlockState, "down", "0")
		// outBlockState = strings.ReplaceAll(outBlockState, "up", "1")
		outBlockState = strings.ReplaceAll(outBlockState, "north", "2")
		outBlockState = strings.ReplaceAll(outBlockState, "south", "3")
		outBlockState = strings.ReplaceAll(outBlockState, "west", "4")
		outBlockState = strings.ReplaceAll(outBlockState, "east", "5")
	} else if outBlockName == "minecraft:piston_extension" {
		outBlockName = "minecraft:piston_arm_collision"
		outBlockState = "block_data=0"
	} else if (outBlockName == "minecraft:piston_head") || (outBlockName == "minecraft:pistonArmCollision") {
		outBlockName = "minecraft:piston_arm_collision"
		outBlockState = "block_data=0"
	}
	outBlockNameForSearch := blocks.BlockNameForSearch(outBlockName)
	outBlockStateForSearch, err := blocks.PropsForSearchFromStr(outBlockState)
	if err != nil {
		panic(err)
	}
	rtid := blocks.UNKNOWM_RUNTIME
	found := false
	// fmt.Printf("%v %v -> %v %v \n", inBlockName, inBlockState, outBlockName, outBlockState)
	if strings.HasPrefix(outBlockState, "block_data=") {
		outBlockState = strings.TrimPrefix(outBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(outBlockState)
		rtid, found = blocks.DefaultAnyToNemcConvertor.PreciseMatchByLegacyValue(outBlockNameForSearch, int16(blockVal))
		if !found {
			rtid, found = blocks.DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(outBlockNameForSearch, int16(blockVal))
			if !found {
				panic(fmt.Errorf("not found!"))
			} else {
				fmt.Printf("fuzzy block data: %v %v\n", outBlockName, blockVal)
			}
		}
	} else {
		rtid, found = blocks.DefaultAnyToNemcConvertor.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
		if !found {
			blocks.DefaultAnyToNemcConvertor.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
			panic(fmt.Errorf("not found!"))
		}
	}
	// fmt.Println(rtid)

	inSS := strings.Split(inSNBT, "[")
	inBlockName, inBlockState := inSS[0], ""
	if len(inSS) > 1 {
		inBlockState = inSS[1]
	}

	inBlockState = strings.TrimSuffix(inBlockState, "]")
	inBlockNameForSearch := blocks.BlockNameForSearch(inBlockName)
	inBlockStateForSearch, err := blocks.PropsForSearchFromStr(inBlockState)
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(inBlockState, "block_data=") {
		inBlockState = strings.TrimPrefix(inBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(inBlockState)
		if existed, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByLegacyValue(inBlockNameForSearch, int16(blockVal), rtid); err != nil {
			fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, blockVal, rtid, outBlockStateForSearch.InPreciseSNBT())
			failed += 1
		} else if !existed {
			translated += 1
			translateRecord = append(translateRecord, struct {
				Name      string
				SNBTState string
				RTID      uint32
			}{
				inBlockNameForSearch.BaseName(),
				fmt.Sprintf("%v", blockVal),
				rtid,
			})
		} else {
			redundant += 1
		}
		// fmt.Printf("%v %v -> %v\n", inBlockNameForSearch, blockVal, rtid)
	} else {
		if existed, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, false); err != nil {
			fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, inBlockStateForSearch.InPreciseSNBT(), rtid, outBlockStateForSearch.InPreciseSNBT())
			failed += 1
		} else if !existed {
			translated += 1
			translateRecord = append(translateRecord, struct {
				Name      string
				SNBTState string
				RTID      uint32
			}{
				inBlockNameForSearch.BaseName(),
				inBlockStateForSearch.InPreciseSNBT(),
				rtid,
			})
		} else {
			redundant += 1
		}
	}
}
