package structure

import (
	"context"
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/string_wrapper"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

type StructureRequester struct {
	ctrl               neomega.GameIntractable
	react              neomega.PacketDispatcher
	EntityUniqueID     int64
	structureListeners *sync_wrapper.SyncKVMap[string, []func(neomega.StructureResponse)]
}

func NewStructureRequester(ctrl neomega.GameIntractable, react neomega.PacketDispatcher, uq neomega.MicroUQHolder) *StructureRequester {
	d := &StructureRequester{
		ctrl:               ctrl,
		react:              react,
		EntityUniqueID:     uq.GetBotBasicInfo().GetBotUniqueID(),
		structureListeners: sync_wrapper.NewSyncKVMap[string, []func(neomega.StructureResponse)](),
	}
	d.react.SetTypedPacketCallBack(packet.IDStructureTemplateDataResponse, func(p packet.Packet) {
		pk := p.(*packet.StructureTemplateDataResponse)
		if cbs, found := d.structureListeners.Get(pk.StructureName); found {
			r := newStructureResponse(pk)
			for _, cb := range cbs {
				cb(r)
			}
			d.structureListeners.Delete(pk.StructureName)
		}
	}, true)
	return d
}

func (o *StructureRequester) ListenStructureResponse(structureName string, callback func(response neomega.StructureResponse)) {
	ls, found := o.structureListeners.Get(structureName)
	if !found {
		ls = make([]func(neomega.StructureResponse), 0)
	}
	ls = append(ls, callback)
	o.structureListeners.Set(structureName, ls)
}

func (o *StructureRequester) requestStructure(pos define.CubePos, size define.CubePos, structureName string, requestType byte) {
	o.ctrl.SendPacket(&packet.StructureTemplateDataRequest{
		StructureName: structureName,
		Position:      protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
		Settings: protocol.StructureSettings{
			PaletteName:               "default",
			IgnoreEntities:            true,
			IgnoreBlocks:              false,
			Size:                      protocol.BlockPos{int32(size.X()), int32(size.Y()), int32(size.Z())},
			Offset:                    protocol.BlockPos{0, 0, 0},
			LastEditingPlayerUniqueID: o.EntityUniqueID,
			Rotation:                  0,
			Mirror:                    0,
			Integrity:                 100,
			Seed:                      0,
			AllowNonTickingChunks:     false,
		},
		RequestType: requestType,
	})
}

type StructureRequestResultHandler struct {
	d             *StructureRequester
	pos           define.CubePos
	size          define.CubePos
	structureName string
	ctx           context.Context
	requestType   byte
}

func (sr *StructureRequestResultHandler) SetContext(ctx context.Context) neomega.StructureRequestResultHandler {
	sr.ctx = ctx
	return sr
}

func (sr *StructureRequestResultHandler) SetTimeout(timeout time.Duration) neomega.StructureRequestResultHandler {
	ctx, _ := context.WithTimeout(sr.ctx, timeout)
	sr.ctx = ctx
	return sr
}

type StructureResponse struct {
	raw              *packet.StructureTemplateDataResponse
	decodedStructure *neomega.DecodedStructure
}

func newStructureResponse(r *packet.StructureTemplateDataResponse) neomega.StructureResponse {
	return &StructureResponse{
		raw: r,
	}
}

func (sr *StructureResponse) Raw() *packet.StructureTemplateDataResponse {
	return sr.raw
}

type StructureContent struct {
	Version   int32    `mapstructure:"format_version" nbt:"format_version"`
	Size      [3]int32 `mapstructure:"size" nbt:"size"`
	Origin    [3]int32 `mapstructure:"structure_world_origin" nbt:"structure_world_origin"`
	Structure struct {
		BlockIndices [2]interface{} `mapstructure:"block_indices" nbt:"block_indices"`
		//Entities     []map[string]interface{} `mapstructure:"entities"`
		Palette struct {
			Default struct {
				BlockPositionData map[string]struct {
					Nbt map[string]interface{} `mapstructure:"block_entity_data" nbt:"block_entity_data"`
				} `mapstructure:"block_position_data" nbt:"block_position_data"`
				BlockPalette []struct {
					Name   string                 `mapstructure:"name" nbt:"name"`
					States map[string]interface{} `mapstructure:"states" nbt:"states"`
					Value  int16                  `mapstructure:"val" nbt:"val"`
				} `mapstructure:"block_palette" nbt:"block_palette"`
			} `mapstructure:"default" nbt:"default"`
		} `mapstructure:"palette" nbt:"palette"`
	} `mapstructure:"structure" nbt:"structure"`
	decoded *neomega.DecodedStructure
}

func (structure *StructureContent) FromNBT(nbt map[string]any) error {
	err := mapstructure.Decode(nbt, &structure)
	if err != nil {
		return err
	}
	return nil
}
func (structure *StructureContent) Decode() *neomega.DecodedStructure {
	nbts := map[define.CubePos]map[string]interface{}{}
	for _, blockNbt := range structure.Structure.Palette.Default.BlockPositionData {
		x, y, z, ok := define.GetPosFromNBT(blockNbt.Nbt)
		if ok {
			nbts[define.CubePos{x, y, z}] = blockNbt.Nbt
		}
	}
	// BlockPalettes := make(map[string]*neomega.BlockPalettes)
	paletteLookUp := make([]uint32, len(structure.Structure.Palette.Default.BlockPalette))
	for paletteIdx, palette := range structure.Structure.Palette.Default.BlockPalette {
		rtid, _ := blocks.BlockNameAndStateToRuntimeID(palette.Name, palette.States)
		paletteLookUp[paletteIdx] = rtid
		// hashName := fmt.Sprintf("%v[%v]", palette.Name, chunk.PropsToStateString(palette.States, false))
		// BlockPalettes[hashName] = &neomega.BlockPalettes{
		// 	Name:   palette.Name,
		// 	States: palette.States,
		// 	Value:  palette.Value,
		// 	RTID:   rtid,
		// 	// NemcRtid:    chunk.AirRID,
		// }
	}
	var foreground, background []uint32
	{
		BlockIndices0, BlockIndices1 := (structure.Structure.BlockIndices[0]).([]int32), (structure.Structure.BlockIndices[1]).([]int32)
		foreground = make([]uint32, len(BlockIndices0))
		background = make([]uint32, len(BlockIndices1))
		_v := int32(0)
		for i, v := range BlockIndices0 {
			_v = v
			if _v != -1 {
				foreground[i] = paletteLookUp[_v]
			} else {
				foreground[i] = blocks.AIR_RUNTIMEID
			}
		}
		for i, v := range BlockIndices1 {
			_v = v
			if _v != -1 {
				background[i] = paletteLookUp[_v]
			} else {
				background[i] = blocks.AIR_RUNTIMEID
			}
		}
	}
	decodeStructure := &neomega.DecodedStructure{
		Version: structure.Version,
		Size: define.CubePos{
			int(structure.Size[0]), int(structure.Size[1]), int(structure.Size[2]),
		},
		Origin: define.CubePos{
			int(structure.Origin[0]), int(structure.Origin[1]), int(structure.Origin[2]),
		},
		ForeGround: foreground,
		BackGround: background,
		Nbts:       nbts,
		// BlockPalettes: BlockPalettes,
	}
	structure.decoded = decodeStructure
	return decodeStructure
}

func (sr *StructureResponse) Decode() (s *neomega.DecodedStructure, err error) {
	if !sr.raw.Success {
		return nil, fmt.Errorf("response get fail result")
	}
	if sr.decodedStructure != nil {
		return sr.decodedStructure, nil
	}
	structureData := sr.raw.StructureTemplate
	structure := &StructureContent{}
	err = structure.FromNBT(structureData)
	if err != nil {
		return nil, err
	}
	decodeStructure := structure.Decode()
	sr.decodedStructure = decodeStructure
	return decodeStructure, nil

}

func (sr *StructureRequestResultHandler) AsyncGetResult(callback func(r neomega.StructureResponse, err error)) {
	w := make(chan neomega.StructureResponse, 1)
	sr.d.ListenStructureResponse(sr.structureName, func(response neomega.StructureResponse) {
		w <- response
	})
	sr.d.requestStructure(sr.pos, sr.size, sr.structureName, sr.requestType)
	go func() {
		select {
		case response := <-w:
			if response.Raw().Success {
				callback(response, nil)
			} else {
				callback(response, fmt.Errorf("response get fail result"))
			}
		case <-sr.ctx.Done():
			callback(nil, sr.ctx.Err())
		}
	}()
}

func (sr *StructureRequestResultHandler) BlockGetResult() (r neomega.StructureResponse, err error) {
	w := make(chan neomega.StructureResponse, 1)
	sr.d.ListenStructureResponse(sr.structureName, func(response neomega.StructureResponse) {
		w <- response
	})
	sr.d.requestStructure(sr.pos, sr.size, sr.structureName, sr.requestType)
	select {
	case response := <-w:
		if response.Raw().Success {
			return response, nil
		} else {
			return response, fmt.Errorf("response get fail result")
		}
	case <-sr.ctx.Done():
		return nil, sr.ctx.Err()
	}
}

func (o *StructureRequester) RequestStructure(pos define.CubePos, size define.CubePos, structureName string) neomega.StructureRequestResultHandler {
	return &StructureRequestResultHandler{
		o, pos, size, structureName, context.Background(), packet.StructureTemplateRequestExportFromSave,
	}
}

func (o *StructureRequester) RequestStructureWithAutoName(pos define.CubePos, size define.CubePos) neomega.StructureRequestResultHandler {
	name := string_wrapper.ReplaceWithUnfilteredLetter(uuid.New().String())
	return o.RequestStructure(pos, size, name)
}
