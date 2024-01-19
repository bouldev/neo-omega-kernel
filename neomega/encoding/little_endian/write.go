package little_endian

import (
	"math"
	"unsafe"
)

type CanWriteBytes interface {
	Write([]byte) error
}

// WriteInt8 ...
func WriteBool(w CanWriteBytes, x bool) error {
	b := uint8(0)
	if x {
		b = 1
	}
	return w.Write([]byte{b})
}

// WriteInt8 ...
func WriteInt8(w CanWriteBytes, x int8) error {
	return w.Write([]byte{byte(x)})
}

func WriteUint8(w CanWriteBytes, x uint8) error {
	return w.Write([]byte{byte(x)})
}

// WriteInt16 ...
func WriteInt16(w CanWriteBytes, x int16) error {
	return w.Write([]byte{byte(x), byte(x >> 8)})
}

func WriteUint16(w CanWriteBytes, x uint16) error {
	return w.Write([]byte{byte(x), byte(x >> 8)})
}

// WriteInt32 ...
func WriteInt32(w CanWriteBytes, x int32) error {
	return w.Write(MakeInt32(x))
}

func MakeInt32(x int32) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}
}

func WriteUint32(w CanWriteBytes, x uint32) error {
	return w.Write(MakeUint32(x))
}

func MakeUint32(x uint32) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}
}

// WriteInt64 ...
func WriteInt64(w CanWriteBytes, x int64) error {
	return w.Write(MakeInt64(x))
}

func MakeInt64(x int64) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24),
		byte(x >> 32), byte(x >> 40), byte(x >> 48), byte(x >> 56)}
}

func WriteUint64(w CanWriteBytes, x uint64) error {
	return w.Write(MakeUint64(x))
}

func MakeUint64(x uint64) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24),
		byte(x >> 32), byte(x >> 40), byte(x >> 48), byte(x >> 56)}
}

// WriteFloat32 ...
func WriteFloat32(w CanWriteBytes, x float32) error {
	bits := math.Float32bits(x)
	return w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24)})
}

// WriteFloat64 ...
func WriteFloat64(w CanWriteBytes, x float64) error {
	bits := math.Float64bits(x)
	return w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
		byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56)})
}

// WriteString ...
func WriteString(w CanWriteBytes, x string) error {
	if len(x) > math.MaxInt16 {
		return ErrStringLengthExceeds
	}
	length := int16(len(x))
	if err := w.Write([]byte{byte(length), byte(length >> 8)}); err != nil {
		return err
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return err
	}
	return nil
}
