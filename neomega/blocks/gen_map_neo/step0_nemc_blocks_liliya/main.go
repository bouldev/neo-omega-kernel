package step0_nemc_blocks_liliya

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"neo-omega-kernel/neomega/blocks"
	"os"
	"sort"
	"strings"
)

type RawState struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (s RawState) ToValue() blocks.PropVal {
	if s.Type == "string" {
		return blocks.PropValFromString(s.Value.(string))
	} else if s.Type == "int" {
		return blocks.PropValFromInt32(int32(s.Value.(float64)))
	} else if s.Type == "byte" {
		if s.Value.(float64) == 0 {
			return blocks.PropVal0
		} else if s.Value.(float64) == 1 {
			return blocks.PropVal1
		} else {
			panic(s.Value)
		}
	} else {
		panic(s.Type)
	}
}

type RawBlockPalette struct {
	LegacyData    uint16     `json:"data"` // up to 5469 @ cobblestone_wall
	BlockName     string     `json:"name"`
	States        []RawState `json:"states"`
	BlockNameHash uint64     `json:"name_hash"`  // maybe some hash of block name
	NetworkID     uint32     `json:"network_id"` // maybe some hash of whole block?
}

func (p RawBlockPalette) DumpStates() (StateOrder []string, State map[string]blocks.PropVal, States blocks.Props) {
	StateOrder = []string{}
	State = map[string]blocks.PropVal{}
	States = blocks.Props{}
	for _, rawState := range p.States {
		p := rawState.ToValue()
		State[rawState.Name] = p
		StateOrder = append(StateOrder, rawState.Name)

		States = append(States, struct {
			Name  string
			Value blocks.PropVal
		}{
			Name:  rawState.Name,
			Value: p,
		})
	}
	if !sort.StringsAreSorted(StateOrder) {
		fmt.Println(StateOrder)
	}
	return StateOrder, State, States
}

type RawData struct {
	Blocks []RawBlockPalette `json:"blocks"`
}

type ParsedBlock struct {
	NameWithoutMC string
	States        blocks.Props
	// State         map[string]blocks.PropVal
	// StateOrder    []string
	LegacyData    uint16
	NemcRuntimeID int32
	Version       int32
}

func ConvertRawData(rawData *RawData) []ParsedBlock {
	b := []byte{1, 20, 10, 0}
	Version := int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
	parsedBlocks := []ParsedBlock{}
	for rtid, block := range rawData.Blocks {
		_, _, States := block.DumpStates()
		parsedBlock := ParsedBlock{
			NameWithoutMC: strings.TrimPrefix(block.BlockName, "minecraft:"),
			States:        States,
			LegacyData:    block.LegacyData,
			NemcRuntimeID: int32(rtid),
			// TODO: version
			Version: Version,
		}
		parsedBlocks = append(parsedBlocks, parsedBlock)
	}
	return parsedBlocks
}

func GetParsedBlock(filePath string) []ParsedBlock {
	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	var rawData RawData
	if err = json.Unmarshal(rawBytes, &rawData); err != nil {
		panic(err)
	}
	return ConvertRawData(&rawData)
}
