package step1_pack_nemc_block_info

import (
	"bytes"
	"fmt"
	"neo-omega-kernel/neomega/blocks/gen_map_neo/step0_nemc_blocks_liliya"

	"github.com/andybalholm/brotli"
)

func PackInfo(blocks []step0_nemc_blocks_liliya.ParsedBlock) (raw string, compressed []byte) {
	version := int32(0)
	datas := ""
	for _, block := range blocks {
		if version == 0 {
			version = (block.Version)
			datas = fmt.Sprintf("VERSION:%v\nCOUNTS:%v\n", version, len(blocks))
		} else if version != (block.Version) {
			panic(fmt.Errorf("version mismatch %v != %v", version, block.Version))
		}
		// fmt.Printf("%v %v %v\n", block.NameWithoutMC, block.LegacyData, block.States.SNBTString())
		datas += fmt.Sprintf("%v %v %v\n", block.NameWithoutMC, block.LegacyData, block.States.SNBTString())
	}
	outBuf := bytes.NewBuffer([]byte{})
	brotliWriter := brotli.NewWriter(outBuf)
	brotliWriter.Write([]byte(datas))
	if err := brotliWriter.Close(); err != nil {
		panic(err)
	}
	return datas, outBuf.Bytes()
}
