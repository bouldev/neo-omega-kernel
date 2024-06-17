package packet

import (
	"neo-omega-kernel/minecraft/protocol"
)

type NemcWhatever struct {
	Value any
}

// ID ...
func (*NemcWhatever) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *NemcWhatever) Marshal(io protocol.IO) {

}
