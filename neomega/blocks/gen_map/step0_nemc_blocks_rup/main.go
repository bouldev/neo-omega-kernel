package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

//go:embed block_runtime_ids_1_19_10_22_with_nbt.json
var blockStateData []byte

func main() {
	var v [][]any
	json.Unmarshal(blockStateData, &v)
	for _, b := range v {
		b64Str := b[2]
		nbtBytes, _ := base64.StdEncoding.DecodeString(string(b64Str.(string)))
		nbtBytes = nbtBytes[6:]
		strLen := nbtBytes[0]
		nbtBytes = nbtBytes[1:]
		blockName := string(nbtBytes[:strLen]) // name
		// fmt.Println(name)
		nbtBytes = nbtBytes[strLen:]
		// var data map[string]any
		// decoder := nbt.NewDecoderWithEncoding(bytes.NewBuffer(nbtBytes), nbt.NetworkLittleEndian)
		// decoder.Decode(&data)

		shouldBe10 := nbtBytes[0]
		nbtBytes = nbtBytes[1:]

		if shouldBe10 != 10 {
			panic(fmt.Errorf("should be 10"))
		}
		strLen = nbtBytes[0]
		nbtBytes = nbtBytes[1:]
		shouldBeStates := string(nbtBytes[:strLen])
		nbtBytes = nbtBytes[strLen:]
		if string(shouldBeStates) != "states" {
			panic(fmt.Errorf("should be states"))
		}
		typeByte := byte(0)
		tagLen := byte((0))
		states := map[string]any{}
		for {
			typeByte = nbtBytes[0]
			if typeByte == 0 {
				nbtBytes = nbtBytes[1:]
				break
			}
			tagLen = nbtBytes[1]
			nbtBytes = nbtBytes[2:]
			propName := string(nbtBytes[:tagLen])
			nbtBytes = nbtBytes[tagLen:]
			if typeByte == 8 { // string?
				stringLen := nbtBytes[0]
				nbtBytes = nbtBytes[1:]
				val := string(nbtBytes[:stringLen])
				states[propName] = val
				nbtBytes = nbtBytes[stringLen:]
			} else if typeByte == 3 { // int32
				intBytes := nbtBytes[:4]
				nbtBytes = nbtBytes[4:]
				val := int32(uint32(intBytes[0]) | uint32(intBytes[1])<<8 | uint32(intBytes[2])<<16 | uint32(intBytes[3])<<24)
				states[propName] = val
			} else if typeByte == 1 { // byte
				val := (nbtBytes[0] == 1)
				if !val && nbtBytes[0] != 0 {
					panic(fmt.Errorf("get byte %v", nbtBytes[0]))
				}
				states[propName] = val
				nbtBytes = nbtBytes[1:]
			} else {
				panic(typeByte)
			}
		}
		// nbtBytes[0]==3
		strLen = nbtBytes[1]
		nbtBytes = nbtBytes[2:]
		shouldBeVersion := string(nbtBytes[:strLen])
		nbtBytes = nbtBytes[strLen:]
		if string(shouldBeVersion) != "version" {
			panic(fmt.Errorf("should be version"))
		}
		intBytes := nbtBytes[:4]
		nbtBytes = nbtBytes[4:]
		version := int32(uint32(intBytes[0]) | uint32(intBytes[1])<<8 | uint32(intBytes[2])<<16 | uint32(intBytes[3])<<24)
		if nbtBytes[0] != 0 {
			panic("should be end")
		}
		fmt.Println(blockName, states, version)
	}
}
