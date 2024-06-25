package defines

import (
	"errors"
	"fmt"
	"neo-omega-kernel/neomega/encoding/little_endian"

	"github.com/google/uuid"
)

var Empty = Values{}
var ErrInsufficientData = errors.New("insufficient data")

func (r Values) IsEmpty() bool {
	return len(r) == 0
}

func (r Values) EqualString(val string) bool {
	if empty := r.IsEmpty(); empty {
		return val == ""
	}
	return string(r[0]) == val
}

func (r Values) ToString() (string, error) {
	if empty := r.IsEmpty(); empty {
		return "", ErrInsufficientData
	}
	return string(r[0]), nil
}

func (r Values) ToBool() (bool, error) {
	if empty := r.IsEmpty(); empty {
		return false, ErrInsufficientData
	}
	if len(r[0]) == 0 {
		return false, ErrInsufficientData
	}
	return r[0][0] != 0, nil
}

func (r Values) ToByte() (byte, error) {
	if empty := r.IsEmpty(); empty {
		return 0, ErrInsufficientData
	}
	if len(r[0]) == 0 {
		return 0, ErrInsufficientData
	}
	return r[0][0], nil
}

func (r Values) ConsumeHead() Values {
	if len(r) == 0 {
		return r
	}
	return r[1:]
}

func (r Values) Extend(vals ...Values) Values {
	newR := r
	for _, val := range vals {
		newR = append(newR, val...)
	}
	return newR
}

func (r Values) ExtendFrags(frags ...[]byte) Values {
	return append(r, frags...)
}

func (r Values) ToStrings() []string {
	if empty := r.IsEmpty(); empty {
		return []string{}
	}
	ss := make([]string, len(r))
	for i, v := range r {
		ss[i] = string(v)
	}
	return ss
}

func (r Values) ToInt64() (int64, error) {
	if empty := r.IsEmpty(); empty || len(r[0]) != 8 {
		return 0, ErrInsufficientData
	}
	return little_endian.GetInt64(r[0])
}

func (r Values) ToInt32() (int32, error) {
	if empty := r.IsEmpty(); empty || len(r[0]) != 4 {
		return 0, ErrInsufficientData
	}
	return little_endian.GetInt32(r[0])
}

func (r Values) ToUint32() (uint32, error) {
	if empty := r.IsEmpty(); empty || len(r[0]) != 4 {
		return 0, ErrInsufficientData
	}
	return little_endian.GetUint32(r[0])
}

func (r Values) ToBytes() ([]byte, error) {
	if empty := r.IsEmpty(); empty {
		return nil, ErrInsufficientData
	}
	return r[0], nil
}

func (r Values) ToUUID() (ud uuid.UUID, err error) {
	if empty := r.IsEmpty(); empty {
		return ud, ErrInsufficientData
	}
	return uuid.FromBytes(r[0])
}

func FromUUID(ud uuid.UUID) Values {
	bs := ud[:]
	return FromFrags(bs)
}

func FromFrags(vals ...[]byte) Values {
	if len(vals) == 0 {
		return Empty
	}
	return vals
}

func FromString(val string) Values {
	return Values{[]byte(val)}
}

func FromBool(b bool) Values {
	if b {
		return Values{[]byte{byte(1)}}
	}
	return Values{[]byte{byte(0)}}
}

func FromByte(b byte) Values {
	return Values{[]byte{b}}
}

func FromInt64(val int64) Values {
	return Values{little_endian.MakeInt64(int64(val))}
}

func FromUint64(val uint64) Values {
	return Values{little_endian.MakeUint64(val)}
}

func FromUint32(val uint32) Values {
	return Values{little_endian.MakeUint32(val)}
}

func FromInt32(val int32) Values {
	return Values{little_endian.MakeInt32(val)}
}

func FromStrings(vals ...string) Values {
	rets := make([][]byte, len(vals))
	for i, v := range vals {
		rets[i] = []byte(v)
	}
	return rets
}

var errorNoResult = errors.New("no result")

func UnwrapOutput(rets Values) (Values, error) {
	if rets.IsEmpty() {
		return Empty, errorNoResult
	} else {
		if rets.EqualString("ok") {
			return rets.ConsumeHead(), nil
		} else {
			return Empty, fmt.Errorf(rets.ConsumeHead().ToString())
		}
	}
}

func WrapOutput(rets Values, err error) Values {
	if err != nil {
		return FromStrings("err", err.Error())
	} else {
		return FromString("ok").Extend(rets)
	}
}
