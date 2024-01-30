package neomega

import (
	"context"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega/blocks"
	"neo-omega-kernel/neomega/mirror"
	"neo-omega-kernel/neomega/mirror/chunk"
	"neo-omega-kernel/neomega/mirror/define"
	"time"
)

type BlockPalettes struct {
	Name   string
	States map[string]interface{}
	Value  int16
	RTID   uint32
	// NemcRtid    uint32
}
type DecodedStructure struct {
	Version    int32
	Size       define.CubePos
	Origin     define.CubePos
	ForeGround []uint32
	BackGround []uint32
	Nbts       map[define.CubePos]map[string]interface{}
	// BlockPalettes map[string]*BlockPalettes
}

func (d *DecodedStructure) IndexOf(pos define.CubePos) int {
	YZ := d.Size.Y() * d.Size.Z()
	return YZ*pos.X() + d.Size.Z()*pos.Y() + pos.Z()
}

func (d *DecodedStructure) BlockOf(pos define.CubePos) (foreGround, backGround uint32) {
	idx := d.IndexOf(pos)
	return d.ForeGround[idx], d.BackGround[idx]
}

func RangeSplits(start int, len int, algin int) (ranges [][2]int) {
	ranges = make([][2]int, 0, (len+algin-1)/algin)
	currentStart := start
	for currentStart < start+len {
		currentEnd := ((currentStart + algin) / algin) * algin
		if currentEnd > start+len {
			currentEnd = start + len
		}
		ranges = append(ranges, [2]int{currentStart, currentEnd - currentStart})
		currentStart = currentEnd
	}
	return ranges
}

func (structure *DecodedStructure) DumpToChunkProvider(chunkProvider mirror.ChunkProvider) (err error) {
	background, foreground := structure.BackGround, structure.ForeGround
	rtid := uint32(0)
	chunkRangesX := RangeSplits(structure.Origin.X(), structure.Size[0], 16)
	chunkRangesZ := RangeSplits(structure.Origin.Z(), structure.Size[2], 16)
	var chunkData *mirror.ChunkData
	for _, xRange := range chunkRangesX {
		for _, zRange := range chunkRangesZ {
			blockPos00 := define.CubePos{int(xRange[0]), int(structure.Origin.Y()), int(zRange[0])}
			chunkPos := define.ChunkPos{int32(blockPos00.X() >> 4), int32(blockPos00.Z() >> 4)}
			if chunkData != nil {
				if err = chunkProvider.Write(chunkData); err != nil {
					return err
				}
			}
			chunkData = chunkProvider.Get(chunkPos)
			if chunkData == nil {
				chunkData = &mirror.ChunkData{
					Chunk:     chunk.New(blocks.AIR_RUNTIMEID, define.WorldRange),
					BlockNbts: make(map[define.CubePos]map[string]interface{}),
					SyncTime:  time.Now().Unix(),
					ChunkPos:  chunkPos,
				}
			}
			for x := xRange[0]; x < xRange[0]+xRange[1]; x++ {
				for z := zRange[0]; z < zRange[0]+zRange[1]; z++ {
					for y := structure.Origin.Y(); y < structure.Origin.Y()+structure.Size[1]; y++ {
						blockPos := define.CubePos{int(x), int(y), int(z)}
						iPos := blockPos.Sub(structure.Origin)
						i := iPos.Z() + int(iPos.Y())*structure.Size[2] + iPos.X()*(structure.Size[1]*structure.Size[2])
						rtid = background[i]
						if rtid != blocks.AIR_RUNTIMEID {
							chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, int16(y), uint8(blockPos.Z())&0xF, 0, rtid)
						}
						rtid = foreground[i]
						if rtid != blocks.AIR_RUNTIMEID {
							chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, int16(y), uint8(blockPos.Z())&0xF, 0, rtid)
						}
						// TODO: Check Block Offset or Block Pos
						nbt, found := structure.Nbts[blockPos]
						if found {
							chunkData.BlockNbts[blockPos] = nbt
						}
					}
				}
			}

		}
	}

	if chunkData != nil {
		if err = chunkProvider.Write(chunkData); err != nil {
			return err
		}
	}
	return nil
}

// func (structure *DecodedStructure) DumpToChunkProvider(chunkProvider mirror.ChunkProvider) (err error) {
// 	background, foreground := structure.BackGround, structure.ForeGround
// 	ybase := int16(structure.Origin[1])
// 	rtid := uint32(0)

// 	var chunkData *mirror.ChunkData
// 	for x := 0; x < structure.Size[0]; x++ {
// 		for y := int16(0); y < int16(structure.Size[1]); y++ {
// 			for z := 0; z < structure.Size[2]; z++ {
// 				blockPos := define.CubePos{int(x), int(y), int(z)}.Add(structure.Origin)
// 				chunkPos := define.ChunkPos{int32(blockPos.X() >> 4), int32(blockPos.Z() >> 4)}
// 				if chunkData == nil || chunkData.ChunkPos != chunkPos {
// 					if chunkData != nil {
// 						if err = chunkProvider.Write(chunkData); err != nil {
// 							return err
// 						}
// 					}
// 					chunkData = chunkProvider.Get(chunkPos)
// 					if chunkData == nil {
// 						chunkData = &mirror.ChunkData{
// 							Chunk:     chunk.New(blocks.AIR_RUNTIMEID, define.WorldRange),
// 							BlockNbts: make(map[define.CubePos]map[string]interface{}),
// 							SyncTime:  time.Now().Unix(),
// 							ChunkPos:  chunkPos,
// 						}
// 					}
// 				}
// 				i := z + int(y)*structure.Size[2] + x*(structure.Size[1]*structure.Size[2])
// 				rtid = background[i]
// 				if rtid != blocks.AIR_RUNTIMEID {
// 					chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, y+ybase, uint8(blockPos.Z())&0xF, 0, rtid)
// 				}
// 				rtid = foreground[i]
// 				if rtid != blocks.AIR_RUNTIMEID {
// 					chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, y+ybase, uint8(blockPos.Z())&0xF, 0, rtid)
// 				}
// 				// TODO: Check Block Offset or Block Pos
// 				nbt, found := structure.Nbts[blockPos]
// 				if found {
// 					chunkData.BlockNbts[blockPos] = nbt
// 				}
// 			}
// 		}
// 	}
// 	if chunkData != nil {
// 		if err = chunkProvider.Write(chunkData); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

type StructureResponse interface {
	Raw() *packet.StructureTemplateDataResponse
	Decode() (s *DecodedStructure, err error)
}
type StructureRequestResultHandler interface {
	BlockGetResult() (r StructureResponse, err error)
	AsyncGetResult(callback func(r StructureResponse, err error))
	SetContext(ctx context.Context) StructureRequestResultHandler
	SetTimeout(timeout time.Duration) StructureRequestResultHandler
}
type StructureRequester interface {
	RequestStructure(pos define.CubePos, size define.CubePos, structureName string) StructureRequestResultHandler
	RequestStructureWithAutoName(pos define.CubePos, size define.CubePos) StructureRequestResultHandler
}
