package little_endian

import (
	"math"
)

type CanReadOutBytes interface {
	ReadOut(len int) (b []byte, err error)
}

func Bool(r CanReadOutBytes) (bool, error) {
	b, err := r.ReadOut(1)
	if err != nil {
		return false, err
	}
	return uint8(b[0]) == 1, nil
}

// Int8 ...
func Int8(r CanReadOutBytes) (int8, error) {
	b, err := r.ReadOut(1)
	if err != nil {
		return 0, err
	}
	return int8(b[0]), nil
}

func Uint8(r CanReadOutBytes) (uint8, error) {
	b, err := r.ReadOut(1)
	if err != nil {
		return 0, err
	}
	return uint8(b[0]), nil
}

// Int16 ...
func Int16(r CanReadOutBytes) (int16, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return 0, err
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), nil
}

func Uint16(r CanReadOutBytes) (uint16, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return 0, err
	}
	return uint16(b[0]) | uint16(b[1])<<8, nil
}

// Int32 ...
func Int32(r CanReadOutBytes) (int32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return GetInt32(b)
}

func GetInt32(b []byte) (int32, error) {
	if len(b) != 4 {
		return 0, ErrDataLengthMismatch
	}
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

func Uint32(r CanReadOutBytes) (uint32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return GetUint32(b)
}

func GetUint32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, ErrDataLengthMismatch
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24, nil
}

// Int64 ...
func Int64(r CanReadOutBytes) (int64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return GetInt64(b)
}

func GetInt64(b []byte) (int64, error) {
	if len(b) != 8 {
		return 0, ErrDataLengthMismatch
	}
	return int64(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

func Uint64(r CanReadOutBytes) (uint64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56, nil
}

// Float32 ...
func Float32(r CanReadOutBytes) (float32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Float64 ...
func Float64(r CanReadOutBytes) (float64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

// String ...
func String(r CanReadOutBytes) (string, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return "", err
	}
	stringLength := int(uint16(b[0]) | uint16(b[1])<<8)
	data, err := r.ReadOut(stringLength)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
