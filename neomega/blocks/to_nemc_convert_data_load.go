package blocks

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

//go:embed "bedrock_java_to_translate.br"
var toNemcDataLoadBedrockJavaTranslateInfo []byte

//go:embed "schem_to_translate.br"
var toNemcDataLoadSchemTranslateInfo []byte

func initToNemcDataLoadBedrockJava() {
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadBedrockJavaTranslateInfo)))
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewBufferString(string(dataBytes)))
	for {
		blockName, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		blockName = strings.TrimSuffix(blockName, "\n")
		if blockName == "" {
			break
		}
		blockNameToAdd := BlockNameForSearch(blockName)
		snbtState, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		snbtState = strings.TrimSuffix(snbtState, "\n")
		rtidStr, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		rtidStr = strings.TrimSuffix(rtidStr, "\n")
		rtid, err := strconv.Atoi(rtidStr)
		if err != nil {
			panic(err)
		}
		legacyBlockValue, err := strconv.Atoi(snbtState)
		if err == nil {
			if exist, err := DefaultAnyToNemcConvertor.AddAnchorByLegacyValue(blockNameToAdd, int16(legacyBlockValue), uint32(rtid)); err != nil || exist == true {
				panic(fmt.Errorf("fail to add translation: %v %v %v", blockName, legacyBlockValue, rtid))
			}
			if exist, err := SchemToNemcConvertor.AddAnchorByLegacyValue(blockNameToAdd, int16(legacyBlockValue), uint32(rtid)); err != nil || exist == true {
				panic(fmt.Errorf("fail to add translation: %v %v %v", blockName, legacyBlockValue, rtid))
			}
		} else {
			props, err := PropsForSearchFromStr(snbtState)
			if err != nil {
				// continue
				panic(err)
			}
			if exist, err := DefaultAnyToNemcConvertor.AddAnchorByState(blockNameToAdd, props, uint32(rtid), false); err != nil || exist == true {
				panic(fmt.Errorf("fail to add translation: %v %v %v", blockName, props.InPreciseSNBT(), rtid))
			}
			if exist, err := SchemToNemcConvertor.AddAnchorByState(blockNameToAdd, props, uint32(rtid), false); err != nil || exist == true {
				panic(fmt.Errorf("fail to add translation: %v %v %v", blockName, props.InPreciseSNBT(), rtid))
			}
		}

	}
}

func initToNemcDataLoadSchem() {
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadSchemTranslateInfo)))
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewBufferString(string(dataBytes)))
	for {
		blockName, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		blockName = strings.TrimSuffix(blockName, "\n")
		if blockName == "" {
			break
		}
		blockNameToAdd := BlockNameForSearch(blockName)
		snbtState, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		snbtState = strings.TrimSuffix(snbtState, "\n")
		rtidStr, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		rtidStr = strings.TrimSuffix(rtidStr, "\n")
		rtid, err := strconv.Atoi(rtidStr)
		if err != nil {
			panic(err)
		}
		props, err := PropsForSearchFromStr(snbtState)
		if err != nil {
			// continue
			panic(err)
		}
		// door direction and so on in initToNemcDataLoadBedrockJava is not precise, overwrite it
		DefaultAnyToNemcConvertor.AddAnchorByState(blockNameToAdd, props, uint32(rtid), false)
		SchemToNemcConvertor.AddAnchorByState(blockNameToAdd, props, uint32(rtid), true)
	}
}
