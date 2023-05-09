package wasmtime

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"unicode/utf16"

	"github.com/artela-network/runtime"
)

const (
	HeaderLen = 8
)

type Memory struct {
	data []byte

	// allocate memory in wasm
	// input: size of memory, return: start address of the allocted memory
	allocate func(int32) (int32, error)
	// refresh the memory buffer
	refresh func(m *Memory)
}

var (
	LE                   = binary.LittleEndian
	_  runtime.ArgHelper = (*Memory)(nil)
)

// Write write message, pack message and write utf-16 bytes to the memory,
// return: the start address of memory
func (m *Memory) Write(msg string) (int32, error) {
	data := fmt.Sprintf("%0[1]*x%s", HeaderLen, len(msg), msg)
	u16s := UFT16Encode([]rune(data))

	ptr, err := m.allocate(int32(len(u16s) * 2))
	if err != nil {
		return 0, err
	}
	m.refresh(m)
	m.writeU16s(u16s, int(ptr))
	return ptr, nil
}

// Read read message from the memory data, read the 8-bytes header first,
// and then read the body by the length defined in header
func (m *Memory) Read(ptr int) (string, error) {
	m.refresh(m)
	if len(m.data)-ptr < HeaderLen*2 {
		return "", errors.New("message not valid")
	}

	headerU16s, err := m.readU16s(ptr, HeaderLen)
	if err != nil {
		return "", err
	}
	dataLenStr := string(UFT16Decode(headerU16s))
	dataLen, err := strconv.ParseInt(string(dataLenStr), 16, 32)
	if err != nil {
		return "", err
	}

	ptr += HeaderLen * 2
	if len(m.data)-ptr < int(dataLen)*2 {
		return "", errors.New("message not valid")
	}

	dataU16s, err := m.readU16s(ptr, int(dataLen))
	if err != nil {
		return "", nil
	}
	return string(UFT16Decode(dataU16s)), nil
}

// WriteU16s write uint16 array to a byte arrary, with byte order = little endian
// little endian matches to the wasm string
func (m *Memory) writeU16s(u16s []uint16, ptr int) {
	for i, c := range u16s {
		uint := make([]byte, 2)
		LE.PutUint16(uint, c)
		m.data[ptr+i*2] = uint[0]
		m.data[ptr+i*2+1] = uint[1]
	}
}

// ReadU16s read uint16 arrary from a byte arrary, with byte order = little endian
// little endian matches to the wasm string
func (m *Memory) readU16s(ptr int, size int) ([]uint16, error) {
	buf := bytes.NewReader(m.data[ptr:])
	u16s := make([]uint16, 0, size)
	for i := 0; i < size; i++ {
		var u16 uint16
		if err := binary.Read(buf, LE, &u16); err != nil {
			return nil, err
		}
		u16s = append(u16s, u16)
	}
	return u16s, nil
}

// UFT16Encode encode a rune slice to uint16 array
func UFT16Encode(src []rune) []uint16 {
	return utf16.Encode(src)
}

// UTF16Decode decode a uint16 array to rune slice
func UFT16Decode(src []uint16) []rune {
	return utf16.Decode(src)
}
