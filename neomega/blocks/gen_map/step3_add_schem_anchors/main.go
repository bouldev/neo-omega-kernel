package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"neo-omega-kernel/neomega/blocks"
	"os"
	"strings"

	"github.com/andybalholm/brotli"
)

//go:embed block_1_18_java_to_bedrock.json
var schemJsonData []byte

type JavaToBedrockMappingIn struct {
	Name       string         `json:"bedrock_identifier"`
	Properties map[string]any `json:"bedrock_states"`
}

var redundant = 0
var overwrite = 0
var translated = 0

var translateRecord = []struct {
	Name      string
	SNBTState string
	RTID      uint32
}{}

func main() {
	// fmt.Println(blocks.DefaultAnyToNemcConvertor)
	javaBlocks := map[string]JavaToBedrockMappingIn{}
	err := json.Unmarshal(schemJsonData, &javaBlocks)
	if err != nil {
		panic(err)
	}
	for blockIn, bedrockBlockDescribe := range javaBlocks {
		outBlockNameForSearch := blocks.BlockNameForSearch(bedrockBlockDescribe.Name)
		// TODO CHECK IF THIS EXIST IN 1.19
		if strings.Contains(outBlockNameForSearch.BaseName(), "mangrove_roots") {
			continue
		}
		outBlockStateForSearch, err := blocks.PropsForSearchFromNbt(bedrockBlockDescribe.Properties)
		if err != nil {
			panic(err)
		}

		rtid, found := blocks.DefaultAnyToNemcConvertor.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
		if !found {
			blocks.DefaultAnyToNemcConvertor.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
			panic("not found!")
		}

		fmt.Println(outBlockNameForSearch, outBlockStateForSearch.InPreciseSNBT(), rtid)
		inSS := strings.Split(blockIn, "[")
		inBlockName, inBlockState := inSS[0], ""
		if len(inSS) > 1 {
			inBlockState = inSS[1]
		}
		hasWaterLoggedInfo := strings.Contains(inBlockState, "waterlogged")
		var inBlockStateFroSearchWaterLogged *blocks.PropsForSearch
		inBlockState = strings.TrimSuffix(inBlockState, "]")
		if hasWaterLoggedInfo {
			inBlockStateFroSearchWaterLogged, err = blocks.PropsForSearchFromStr(inBlockState)
			if err != nil {
				panic(err)
			}
			inBlockState = strings.ReplaceAll(inBlockState, ",waterlogged=true", "")
			inBlockState = strings.ReplaceAll(inBlockState, ",waterlogged=false", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=true,", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=false,", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=true", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=false", "")

		}

		inBlockNameForSearch := blocks.BlockNameForSearch(inBlockName)
		inBlockStateForSearch, err := blocks.PropsForSearchFromStr(inBlockState)
		if err != nil {
			panic(err)
		}
		if strings.HasPrefix(inBlockState, "block_data=") {
			panic("not implement")
		} else {
			if exist, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, false); err != nil {
				overwrite++
				translateRecord = append(translateRecord, struct {
					Name      string
					SNBTState string
					RTID      uint32
				}{
					inBlockNameForSearch.BaseName(),
					inBlockStateForSearch.InPreciseSNBT(),
					rtid,
				})
				if _, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, true); err != nil {
					panic(err)
				}
			} else if exist {
				redundant++
			} else {
				translated++
				translateRecord = append(translateRecord, struct {
					Name      string
					SNBTState string
					RTID      uint32
				}{
					inBlockNameForSearch.BaseName(),
					inBlockStateForSearch.InPreciseSNBT(),
					rtid,
				})
			}
			// fmt.Println(inBlockNameForSearch, inBlockStateForSearch.InPreciseSNBT(), hasWaterLoggedTrueInfo)
			if inBlockStateFroSearchWaterLogged != nil {
				if exist, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByState(inBlockNameForSearch, inBlockStateFroSearchWaterLogged, rtid, false); err != nil {
					overwrite++
					translateRecord = append(translateRecord, struct {
						Name      string
						SNBTState string
						RTID      uint32
					}{
						inBlockNameForSearch.BaseName(),
						inBlockStateFroSearchWaterLogged.InPreciseSNBT(),
						rtid,
					})
					if _, err := blocks.DefaultAnyToNemcConvertor.AddAnchorByState(inBlockNameForSearch, inBlockStateFroSearchWaterLogged, rtid, true); err != nil {
						panic(err)
					}
				} else if exist {
					redundant++
				} else {
					translated++
					translateRecord = append(translateRecord, struct {
						Name      string
						SNBTState string
						RTID      uint32
					}{
						inBlockNameForSearch.BaseName(),
						inBlockStateFroSearchWaterLogged.InPreciseSNBT(),
						rtid,
					})
				}
			}
		}
	}
	fmt.Printf("ok %v overwrite %v redundant %v\n", translated, overwrite, redundant)
	textRecord := ""
	for _, record := range translateRecord {
		textRecord += fmt.Sprintf("%v\n%v\n%v\n", record.Name, record.SNBTState, record.RTID)
	}
	os.WriteFile("../schem_to_translate.txt", []byte(textRecord), 0755)
	outFp, _ := os.OpenFile("../schem_to_translate.br", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	brotliWriter := brotli.NewWriter(outFp)
	brotliWriter.Write([]byte(textRecord))
	if err := brotliWriter.Close(); err != nil {
		panic(err)
	}
	if err := outFp.Close(); err != nil {
		panic(err)
	}
}
