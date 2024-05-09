package packet

import (
	"neo-omega-kernel/minecraft/protocol"
)

type PyRpc struct {
	Value any
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *PyRpc) Marshal(io protocol.IO) {
	io.MsgPack(&pk.Value)
}
