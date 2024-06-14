package byte_frame_conn

import (
	"fmt"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/cascade_conn/can_close"
	"neo-omega-kernel/minecraft_neo/cascade_conn/defines"
	"net"
	"sync"
	"time"

	"github.com/sandertv/go-raknet"
)

type ByteFrameConnection struct {
	can_close.CanCloseWithError
	netConn net.Conn
	enc     *packet.Encoder
	dec     *packet.Decoder

	// bufferedSend is a slice of byte slices containing packets that are 'written'. They are buffered until
	// they are sent each 20th of a second.
	bufferedSend [][]byte
	sendMu       sync.Mutex
}

func NewConnectionFromNet(netConn net.Conn) defines.ByteFrameConn {
	conn := &ByteFrameConnection{
		netConn: netConn,
		enc:     packet.NewEncoder(netConn),
		dec:     packet.NewDecoder(netConn),
	}
	conn.CanCloseWithError = can_close.NewClose(func() { conn.netConn.Close() })
	conn.dec.DisableBatchPacketLimit()
	go conn.writeRoutine(time.Second / 20)
	return conn
}

func (c *ByteFrameConnection) writeRoutine(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for range ticker.C {
		if err := c.Flush(); err != nil {
			c.CloseWithError(err)
			return
		}
	}
}

func (conn *ByteFrameConnection) Flush() error {
	select {
	case <-conn.WaitClosed():
		return fmt.Errorf("flush on close connection")
	default:
	}
	conn.sendMu.Lock()
	defer conn.sendMu.Unlock()

	if len(conn.bufferedSend) > 0 {
		if err := conn.enc.Encode(conn.bufferedSend); err != nil && !raknet.ErrConnectionClosed(err) {
			// Should never happen.
			return fmt.Errorf("error encoding packet batch: %v", err)
		}
		// First manually clear out conn.bufferedSend so that re-using the slice after resetting its length to
		// 0 doesn't result in an 'invisible' memory leak.
		for i := range conn.bufferedSend {
			conn.bufferedSend[i] = nil
		}
		// Slice the conn.bufferedSend to a length of 0 so we don't have to re-allocate space in this slice
		// every time.
		conn.bufferedSend = conn.bufferedSend[:0]
	}
	return nil
}

func (conn *ByteFrameConnection) EnableEncryption(key [32]byte) {
	conn.enc.EnableEncryption(key)
	conn.dec.EnableEncryption(key)
}

func (conn *ByteFrameConnection) EnableCompression(algorithm packet.Compression) {
	conn.enc.EnableCompression(algorithm)
	conn.dec.EnableCompression(algorithm)
}

func (conn *ByteFrameConnection) WriteBytePacket(packet []byte) {
	if conn.Closed() {
		return
	}
	conn.sendMu.Lock()
	defer conn.sendMu.Unlock()
	conn.bufferedSend = append(conn.bufferedSend, packet)
}

func (conn *ByteFrameConnection) Lock() {
	conn.sendMu.Lock()
}
func (conn *ByteFrameConnection) UnLock() {
	conn.sendMu.Unlock()
}

func (conn *ByteFrameConnection) ReadRoutine(onPacket func([]byte)) {
	for {
		packets, err := conn.dec.Decode()
		if err != nil {
			if !raknet.ErrConnectionClosed(err) {
				fmt.Printf("error reading from dialer connection: %v\n", err)
			}
			conn.CloseWithError(err)
		}
		for _, data := range packets {
			onPacket(data)
		}
	}
}
