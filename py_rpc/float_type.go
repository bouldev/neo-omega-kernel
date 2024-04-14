package py_rpc

import (
	"encoding/binary"
	"math"
)

type PyRpcFloatObject struct {
	Value float64
}

func (o *PyRpcFloatObject) Marshal() []byte {
	if float64(float32(o.Value)) == o.Value {
		buf := make([]byte, 1+4)
		buf[0] = 0xca
		binary.BigEndian.PutUint32(buf[1:], math.Float32bits(float32(o.Value)))
		return buf
	} else {
		buf := make([]byte, 1+8)
		buf[0] = 0xcb
		binary.BigEndian.PutUint64(buf[1:], math.Float64bits(o.Value))
		return buf
	}
}

func (o *PyRpcFloatObject) Parse(v []byte) uint {
	if v[0] == 0xca {
		bits := binary.BigEndian.Uint32(v[1:])
		o.Value = float64(math.Float32frombits(bits))
		return 1 + 4
	} else if v[0] == 0xcb {
		bits := binary.BigEndian.Uint64(v[1:])
		o.Value = math.Float64frombits(bits)
		return 1 + 8
	}
	panic("PyRpcFloatObject/Parse: Invalid value")
}

func (*PyRpcFloatObject) Type() uint {
	return FloatType
}

func (o *PyRpcFloatObject) MakeGo() any {
	return o.Value
}

func (o *PyRpcFloatObject) FromGo(v any) {
	pv := v.(float64)
	o.Value = pv
}
