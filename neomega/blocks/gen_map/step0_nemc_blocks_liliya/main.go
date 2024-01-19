package main

import (
	"bytes"
	"fmt"
	"io"
	"neo-omega-kernel/minecraft/nbt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

type ParsedBlock struct {
	Name          string
	State         map[string]any
	StateOrder    []string
	Val           int16
	NemcRuntimeID int32
	Version       int32
}

func ParseBlockState(rawLines []string) (State map[string]any, StateOrder []string) {
	State = map[string]any{}
	StateOrder = []string{}
	entriesStr := strings.TrimSuffix(strings.TrimPrefix(rawLines[0], "TAG_Compound: "), " entries")
	entries, _ := strconv.Atoi(entriesStr)
	if rawLines[0] != fmt.Sprintf("TAG_Compound: %v entries", entries) {
		panic(rawLines[0])
	}
	if rawLines[1] != "{" {
		panic(rawLines[1])
	}
	if rawLines[len(rawLines)-1] != "}" {
		panic(rawLines[len(rawLines)-1])
	}
	lines := rawLines[2 : len(rawLines)-1]
	if len(lines) != entries*2 {
		panic(lines)
	}
	for i := 0; i < entries; i++ {
		nameLine := lines[i*2]
		name := strings.TrimPrefix(nameLine, "// ")
		if nameLine != fmt.Sprintf("// %v", name) {
			panic(nameLine)
		}
		StateOrder = append(StateOrder, name)
		valLine := lines[i*2+1]
		if strings.HasPrefix(valLine, "TAG_Int: ") {
			valStr := strings.TrimPrefix(valLine, "TAG_Int: ")
			val, _ := strconv.Atoi(valStr)
			State[name] = int32(val)
			if valLine != fmt.Sprintf("TAG_Int: %v", val) {
				panic(valLine)
			}
		} else if strings.HasPrefix(valLine, "TAG_Byte:") {
			valStr := strings.TrimPrefix(valLine, "TAG_Byte:")
			valStr = strings.TrimSpace(valStr)
			val := uint8(1)
			if valStr == "" {
				val = uint8(0)
			}
			State[name] = val
			sval := "\x01"
			if val == 0 {
				sval = ""
			}
			if valLine != fmt.Sprintf("TAG_Byte:%v", string(sval)) && valLine != fmt.Sprintf("TAG_Byte: %v", string(sval)) {
				panic(valLine)
			}
		} else if strings.HasPrefix(valLine, "TAG_String:") {
			val := strings.TrimPrefix(valLine, "TAG_String:")
			State[name] = val
			if valLine != fmt.Sprintf("TAG_String:%v", val) {
				panic(valLine)
			}
		} else {
			panic(valLine)
		}
	}
	if len(StateOrder) != entries {
		panic(entries)
	}
	if len(State) != entries {
		panic(State)
	}
	if !sort.StringsAreSorted(StateOrder) {
		fmt.Println(StateOrder)
	}
	return
}

func ParseBlock(rawBlock string, rtid int) ParsedBlock {
	b := ParsedBlock{
		NemcRuntimeID: int32(rtid),
	}
	lines := strings.Split(rawBlock, "\n")
	if lines[0] != "TAG_Compound: 4 entries" {
		panic(lines[0])
	}
	if lines[1] != "{" {
		panic(lines[1])
	}
	if lines[len(lines)-1] != "}" {
		panic(lines[len(lines)-1])
	}
	lines = lines[2 : len(lines)-1]
	if lines[0] != "// name" {
		panic(lines[0])
	}
	name := strings.TrimPrefix(lines[1], "TAG_String: ")
	if lines[1] != "TAG_String: "+name {
		panic(lines[1])
	}
	b.Name = name
	if lines[2] != "// states" {
		panic(lines[2])
	}
	versionStr := lines[len(lines)-1]
	versionStr = strings.TrimSpace(strings.TrimPrefix(versionStr, "TAG_Int: "))
	version, _ := strconv.Atoi(versionStr)
	if lines[len(lines)-1] != fmt.Sprintf("TAG_Int: %v", version) {
		panic(lines[len(lines)-1])
	}
	if lines[len(lines)-2] != "// version" {
		panic(lines[len(lines)-2])
	}
	b.Version = int32(version)
	valStr := lines[len(lines)-3]
	valStr = strings.TrimSpace(strings.TrimPrefix(valStr, "TAG_Short: "))
	val, _ := strconv.Atoi(valStr)
	if lines[len(lines)-3] != fmt.Sprintf("TAG_Short: %v", val) {
		panic(lines[len(lines)-3])
	}
	if lines[len(lines)-4] != "// val" {
		panic(lines[len(lines)-4])
	}
	b.Val = int16(val)
	lines = lines[3 : len(lines)-4]
	b.State, b.StateOrder = ParseBlockState(lines)
	return b
}

func main() {
	var rawData string
	{
		fp, _ := os.Open("block.txt")
		byteData, _ := io.ReadAll(fp)
		rawData = string(byteData)
	}
	var splitedRawData []string
	currentBlock := ""
	{
		splitedRawData = make([]string, 0)

		for _, line := range strings.Split(rawData, "\n") {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.HasPrefix(line, "JS count:") {
				if currentBlock != "" {
					splitedRawData = append(splitedRawData, strings.TrimSpace(currentBlock))
					currentBlock = ""
				}
				if line != fmt.Sprintf("JS count: %v", len(splitedRawData)) {
					panic(len(splitedRawData))
				}

				continue
			}
			currentBlock += "\n" + line
		}
	}
	splitedRawData = append(splitedRawData, strings.TrimSpace(currentBlock))
	parsedBlocks := []ParsedBlock{}
	for nemcRuntimeID, rawBlock := range splitedRawData {
		parsedBlocks = append(parsedBlocks, ParseBlock(rawBlock, nemcRuntimeID))
	}

	{

		// bufWriter := bufio.NewWriter(fp)
		//

		buf := bytes.NewBuffer([]byte{})
		nbtEncoder := nbt.NewEncoder(buf)
		if err := nbtEncoder.Encode(parsedBlocks); err != nil {
			panic(err)
		}

		fp, _ := os.OpenFile("dump.brotli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		brotliWriter := brotli.NewWriter(fp)
		brotliWriter.Write(buf.Bytes())
		if err := brotliWriter.Close(); err != nil {
			panic(err)
		}
		if err := fp.Close(); err != nil {
			panic(err)
		}
	}

	{
		fp, _ := os.Open("dump.brotli")
		dataBytes, _ := io.ReadAll(brotli.NewReader(fp))
		var reopened []ParsedBlock
		if err := nbt.NewDecoder(bytes.NewBuffer(dataBytes)).Decode(&reopened); err != nil {
			panic(err)
		}

		// double check
		for i, orig := range parsedBlocks {
			reopen := reopened[i]
			if orig.Name != reopen.Name || orig.NemcRuntimeID != reopen.NemcRuntimeID || orig.Val != reopen.Val || orig.Version != reopen.Version {
				panic(orig)
			}
			if fmt.Sprintf("%v", orig.StateOrder) != fmt.Sprintf("%v", reopen.StateOrder) {
				panic(orig)
			}
			for k, v := range orig.State {
				if v != reopen.State[k] {
					panic(orig)
				}
			}
		}
	}
	fmt.Println("ok")
}
