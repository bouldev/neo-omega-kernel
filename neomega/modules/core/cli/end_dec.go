package main

import (
	"fmt"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
)

type oneFrameIO struct {
	data []byte
}

func (io *oneFrameIO) Write(p []byte) (n int, err error) {
	io.data = p
	return len(p), nil
}
func (io *oneFrameIO) Read(p []byte) (n int, err error) {
	panic("should not use this")
}

func (io *oneFrameIO) ReadPacket() ([]byte, error) {
	return io.data, nil
}

func byteSlicesToBytes(ss [][]byte) []byte {
	io := &oneFrameIO{}
	enc := packet.NewEncoder(io)
	enc.EnableCompression(packet.SnappyCompression)
	enc.Encode(ss)
	return io.data
}

func bytesToBytesSlices(ss []byte) [][]byte {
	io := &oneFrameIO{data: ss}
	enc := packet.NewDecoder(io)
	enc.EnableCompression(packet.SnappyCompression)
	pks, _ := enc.Decode()
	return pks
}

func main() {
	out := byteSlicesToBytes([][]byte{[]byte("Hello"), []byte("World")})
	outs := bytesToBytesSlices(out)
	for _, o := range outs {
		fmt.Println(string(o))
	}
}
