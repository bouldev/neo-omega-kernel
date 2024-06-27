package underlay_conn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"neo-omega-kernel/minecraft_neo/can_close"
	"neo-omega-kernel/neomega/encoding/little_endian"

	// "neo-omega-kernel/minecraft_neo/defines"
	"net"
	"sync"
	"time"
)

type Compression interface {
	Compress(decompressed []byte) ([]byte, error)
	Decompress(compressed []byte) ([]byte, error)
}

type Encryption interface {
	Encrypt(message []byte) ([]byte, error)
	Decrypt(cipher []byte) ([]byte, error)
}

type ByteFrameConnection struct {
	can_close.CanCloseWithError
	netConn net.Conn

	compression  Compression
	encryption   Encryption
	bufferedSend [][]byte
	sendMu       sync.Mutex
}

func NewConnectionFromNet(netConn net.Conn) *ByteFrameConnection {
	conn := &ByteFrameConnection{
		// close underlay conn on err
		CanCloseWithError: can_close.NewClose(func() { netConn.Close() }),
		netConn:           netConn,
	}
	go conn.writeRoutine(time.Second / 20)
	return conn
}

func NewConnectionFromNetWithCtx(netConn net.Conn, ctx context.Context) *ByteFrameConnection {
	conn := NewConnectionFromNet(netConn)
	go func() {
		select {
		case <-conn.WaitClosed():
		case <-ctx.Done():
			conn.CloseWithError(ctx.Err())
		}
	}()
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
	if conn.Closed() {
		return conn.CloseError()
	}
	conn.sendMu.Lock()
	defer conn.sendMu.Unlock()

	if len(conn.bufferedSend) > 0 {
		if err := conn.writePackets(conn.bufferedSend); err != nil {
			// Should never happen.
			return err
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

func (conn *ByteFrameConnection) EnableEncryption(encryption Encryption) {
	conn.encryption = encryption
}

func (conn *ByteFrameConnection) EnableCompression(algorithm Compression) {
	conn.compression = algorithm
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
		pks, err := conn.readPackets()
		if err != nil {
			conn.CloseWithError(err)
			return
		}
		for _, data := range pks {
			onPacket(data)
		}
	}
}

func writeVaruint32(dst io.Writer, x uint32, b []byte) error {
	b[4] = 0
	b[3] = 0
	b[2] = 0
	b[1] = 0
	b[0] = 0

	i := 0
	for x >= 0x80 {
		b[i] = byte(x) | 0x80
		i++
		x >>= 7
	}
	b[i] = byte(x)
	_, err := dst.Write(b[:i+1])
	return err
}

func readVaruint32(src io.ByteReader, x *uint32) error {
	var v uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := src.ReadByte()
		if err != nil {
			return err
		}
		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return nil
		}
	}
	return errors.New("varuint32 did not terminate after 5 bytes")
}

func (conn *ByteFrameConnection) writePackets(packets [][]byte) (err error) {
	buf := bytes.NewBuffer([]byte{})
	l := make([]byte, 5)
	for _, packet := range packets {
		// Each packet is prefixed with a varuint32 specifying the length of the packet.
		if err := writeVaruint32(buf, uint32(len(packet)), l); err != nil {
			return fmt.Errorf("error writing varuint32 length: %v", err)
		}
		if _, err := buf.Write(packet); err != nil {
			return fmt.Errorf("error writing packet payload: %v", err)
		}
	}
	data := buf.Bytes()
	if conn.compression != nil {
		var err error
		data, err = conn.compression.Compress(data)
		if err != nil {
			return fmt.Errorf("error compressing packet: %v", err)
		}
	}
	if conn.encryption != nil {
		// If the encryption session is not nil, encryption is enabled, meaning we should encrypt the
		// compressed data of this packet.
		data, err = conn.encryption.Encrypt(data)
		if err != nil {
			return fmt.Errorf("error encrypt packet: %v", err)
		}
	}
	dataLen := len(data)
	lenBytes := little_endian.MakeUint32(uint32(dataLen))
	if n, err := conn.netConn.Write(lenBytes); err != nil {
		return err
	} else if n != 4 {
		for n != 4 {
			if consumed, err := conn.netConn.Write(lenBytes[n:]); err != nil {
				return err
			} else {
				n += consumed
			}
		}
	}
	if n, err := conn.netConn.Write(data); err != nil {
		return err
	} else if n != dataLen {
		for n != dataLen {
			if consumed, err := conn.netConn.Write(data[n:]); err != nil {
				return err
			} else {
				n += consumed
			}
		}
	}
	return nil
}

func (conn *ByteFrameConnection) readPackets() (packets [][]byte, err error) {
	lenBytes := make([]byte, 4)
	_, err = io.ReadFull(conn.netConn, lenBytes)
	if err != nil {
		return nil, err
	}
	dataLen, err := little_endian.GetUint32(lenBytes)
	if err != nil {
		return nil, err
	}
	data := make([]byte, dataLen)
	_, err = io.ReadFull(conn.netConn, data)
	if err != nil {
		return nil, err
	}
	if conn.encryption != nil {
		data, err = conn.encryption.Decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("error decrypt packet: %v", err)
		}
	}
	if conn.compression != nil {
		data, err = conn.compression.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("error decompress packet: %v", err)
		}
	}
	b := bytes.NewBuffer(data)
	for b.Len() != 0 {
		var length uint32
		if err := readVaruint32(b, &length); err != nil {
			return nil, err
		}
		packets = append(packets, b.Next(int(length)))
	}
	return packets, nil
}
