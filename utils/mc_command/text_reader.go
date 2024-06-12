package mc_command

import (
	"fmt"
	"strings"
)

// TODO: move this part as a common module
type TextReader interface {
	// if "" means EOF
	Current() string
	CurrentThenNext() string
	Next()
	Snapshot() func()
}

type SimpleTextReader struct {
	underlayData []rune
	ptr          int
	currentS     string
}

func (r *SimpleTextReader) String() string {
	if r.currentS == "" {
		return fmt.Sprintf("%v[EOF]", string(r.underlayData))
	}
	before := r.underlayData[0:r.ptr]
	after := r.underlayData[r.ptr+1 : len(r.underlayData)]
	return fmt.Sprintf("%v[%v]%v", string(before), string(r.currentS), string(after))
}
func (r *SimpleTextReader) Current() string {
	return r.currentS
}

func (r *SimpleTextReader) CurrentThenNext() string {
	s := r.currentS
	r.Next()
	return s
}

func (r *SimpleTextReader) Next() {
	r.ptr += 1
	if r.ptr == len(r.underlayData) {
		r.currentS = ""
	} else {
		r.currentS = string(r.underlayData[r.ptr])
	}
}

func (r *SimpleTextReader) Snapshot() func() {
	ptr := r.ptr
	currentS := r.currentS
	return func() {
		r.ptr = ptr
		r.currentS = currentS
	}
}

func NewSimpleTextReader(data []rune) TextReader {
	r := &SimpleTextReader{
		underlayData: data,
		ptr:          0,
	}
	if len(r.underlayData) > 0 {
		c := r.underlayData[r.ptr]
		r.currentS = string(c)
	}
	return r
}

func CleanStringAndNewSimpleTextReader(data string) TextReader {
	return NewSimpleTextReader([]rune(strings.TrimSpace(data)))
}
