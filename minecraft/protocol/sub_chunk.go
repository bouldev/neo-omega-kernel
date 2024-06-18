package protocol

import "math"

const (
	HeightMapDataNone = iota
	HeightMapDataHasData
	HeightMapDataTooHigh
	HeightMapDataTooLow
)

const (
	SubChunkRequestModeLimitless = math.MaxUint32 - iota
	SubChunkRequestModeLimited
)

const (
	SubChunkResultSuccess = iota + 1
	SubChunkResultChunkNotFound
	SubChunkResultInvalidDimension
	SubChunkResultPlayerNotFound
	SubChunkResultIndexOutOfBounds
	SubChunkResultSuccessAllAir
)

// SubChunkEntry contains the data of a sub-chunk entry relative to a center sub chunk position, used for the sub-chunk
// requesting system introduced in v1.18.10.
type SubChunkEntry struct {
	// Offset contains the offset between the sub-chunk position and the center position.
	Offset SubChunkOffset
	// Result is always one of the constants defined in the SubChunkResult constants.
	Result byte
	// RawPayload contains the serialized sub-chunk data.
	RawPayload []byte
	// HeightMapType is always one of the constants defined in the HeightMapData constants.
	HeightMapType byte
	// HeightMapData is the data for the height map.
	// Netease
	HeightMapData []uint8
	// BlobHash is the hash of the blob.
	BlobHash uint64
}

// Marshal encodes/decodes a SubChunkEntry assuming the blob cache is enabled.
func (x *SubChunkEntry) Marshal(r IO) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	if x.Result != SubChunkResultSuccessAllAir {
		r.ByteSlice(&x.RawPayload)
	}
	r.Uint8(&x.HeightMapType)
	if x.HeightMapType == HeightMapDataHasData {
		FuncSliceOfLen(r, 256, &x.HeightMapData, r.Uint8) // For Netease
	}
	r.Uint64(&x.BlobHash)
}

// SubChunkEntryNoCache encodes/decodes a SubChunkEntry assuming the blob cache is not enabled.
func SubChunkEntryNoCache(r IO, x *SubChunkEntry) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	r.ByteSlice(&x.RawPayload)
	r.Uint8(&x.HeightMapType)
	if x.HeightMapType == HeightMapDataHasData {
		FuncSliceOfLen(r, 256, &x.HeightMapData, r.Uint8) // For Netease
	}
}

// SubChunkOffset represents an offset from the base position of another sub chunk.
// Netease
type SubChunkOffset [3]uint8

// Marshal encodes/decodes a SubChunkOffset.
// Netease: Int8 -> Uint8
func (x *SubChunkOffset) Marshal(r IO) {
	r.Uint8(&x[0])
	r.Uint8(&x[1])
	r.Uint8(&x[2])
}
