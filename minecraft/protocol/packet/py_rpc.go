package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

const (
	PyRpcOperationTypeSend = 0x05db23ae
	PyRpcOperationTypeRecv = 0x0094d408
)

// Netease packet for Python RPC
type PyRpc struct {
	// Value from/to msgpack format
	Value any
	// OperationType is a fixed number
	OperationType uint32
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *PyRpc) Marshal(io protocol.IO) {
	io.MsgPack(&pk.Value)
	io.Uint32(&pk.OperationType)
}
