package main

import (
	"bytes"
	"fmt"
	"io"
	"neo-omega-kernel/minecraft/nbt"
	"neo-omega-kernel/neomega/blocks"
	"os"
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

func main() {
	fp, _ := os.Open("dump.brotli")
	dataBytes, _ := io.ReadAll(brotli.NewReader(fp))
	var reopened []ParsedBlock
	if err := nbt.NewDecoder(bytes.NewBuffer(dataBytes)).Decode(&reopened); err != nil {
		panic(err)
	}

	version := int32(0)
	datas := ""
	for _, block := range reopened {
		if version == 0 {
			version = (block.Version)
			datas = fmt.Sprintf("VERSION:%v\nCOUNTS:%v\n", version, len(reopened))
		} else if version != (block.Version) {
			panic(fmt.Errorf("version mismatch %v != %v", version, block.Version))
		}
		for k, v := range block.State {
			if sv, ok := v.(string); ok {
				block.State[k] = strings.TrimSpace(sv)
			}
		}
		fmt.Printf("%v %v %v\n", strings.TrimPrefix(block.Name, "minecraft:"), block.Val, blocks.PropsFromNbt(block.State).SNBTString())
		datas += fmt.Sprintf("%v %v %v\n", strings.TrimPrefix(block.Name, "minecraft:"), block.Val, blocks.PropsFromNbt(block.State).SNBTString())
	}
	outFp, _ := os.OpenFile("../nemc.br", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	brotliWriter := brotli.NewWriter(outFp)
	brotliWriter.Write([]byte(datas))
	if err := brotliWriter.Close(); err != nil {
		panic(err)
	}
	if err := outFp.Close(); err != nil {
		panic(err)
	}
}
