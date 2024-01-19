package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"neo-omega-kernel/neomega/alter/nbt"
	"os"
)

func main() {
	data, err := os.ReadFile("r.0.0.mca")
	if err != nil {
		panic(err)
	}
	var r io.Reader = bytes.NewReader(data[1:])

	switch data[0] {
	default:
		err = errors.New("unknown compression")
	case 1:
		r, err = gzip.NewReader(r)
	case 2:
		r, err = zlib.NewReader(r)
	case 3:
		// none compression
	}
	if err != nil {
		panic(err)
	}

	d := nbt.NewDecoder(r)
	// d.DisallowUnknownFields()
	var c any
	_, err = d.Decode(c)
	fmt.Println(c)
	if err != nil {
		panic(err)
	}
}
