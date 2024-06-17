package access_helper

import (
	"neo-omega-kernel/minecraft/protocol/login"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/cascade_conn/defines"
	"neo-omega-kernel/minecraft_neo/login_and_spawn_core"
	"sync"
)

type InfinityQueue struct {
	mu             sync.Mutex
	queuedPacket   []*defines.RawPacketAndByte
	upcomingPacket chan *defines.RawPacketAndByte
}

func (q *InfinityQueue) takeDeferredPacket() (*defines.RawPacketAndByte, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queuedPacket) == 0 {
		return nil, false
	}
	data := q.queuedPacket[0]
	// Explicitly clear out the packet at offset 0. When we slice it to remove the first element, that element
	// will not be garbage collectable, because the array it's in is still referenced by the slice. Doing this
	// makes sure garbage collecting the packet is possible.
	q.queuedPacket[0] = nil
	q.queuedPacket = q.queuedPacket[1:]
	return data, true
}

func (q *InfinityQueue) deferPacket(pk *defines.RawPacketAndByte) {
	q.mu.Lock()
	q.queuedPacket = append(q.queuedPacket, pk)
	q.mu.Unlock()
}

func (q *InfinityQueue) PutPacket(pk packet.Packet, data []byte) {
	select {
	case previous := <-q.upcomingPacket:
		// There was already a packet in this channel, so take it out and defer it so that it is read
		// next.
		q.deferPacket(previous)
	default:
	}
	q.upcomingPacket <- &defines.RawPacketAndByte{
		Pk:  pk,
		Raw: data,
	}
}

func (q *InfinityQueue) ReadPacketAndBytes() (pk packet.Packet, data []byte) {
	if pkAndBytes, ok := q.takeDeferredPacket(); ok {
		return pkAndBytes.Pk, pkAndBytes.Raw
	}
	pkAndBytes := <-q.upcomingPacket
	return pkAndBytes.Pk, pkAndBytes.Raw
}

func NewInfinityQueue() *InfinityQueue {
	return &InfinityQueue{
		mu:             sync.Mutex{},
		queuedPacket:   make([]*defines.RawPacketAndByte, 0),
		upcomingPacket: make(chan *defines.RawPacketAndByte, 8),
	}
}

type shallowWrap struct {
	defines.ByteFrameConnBase
	defines.PacketConnBase
	*login_and_spawn_core.Core
	*InfinityQueue
	identityData login.IdentityData
}

func (w *shallowWrap) IdentityData() login.IdentityData {
	return w.identityData
}

func (w *shallowWrap) WritePacket(pk packet.Packet) {
	w.PacketConnBase.WritePacket(pk)
}
