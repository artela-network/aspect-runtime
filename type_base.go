package runtime

import (
	"reflect"
	"unsafe"
)

// IType is the interface of all runtime types
type IType interface {
	// Store write value to linear memory,
	// and return the start address
	Store() (int32, error)

	// Load read value from linear memory with the given ptr
	Load(ptr int32)

	Set(value interface{}) error
	Get() interface{}

	DataType() reflect.Type
}

type TypeHeader struct {
	dataType int16

	// the length of basic data, or 4 for reference type
	dataLen int32
}

// HStore write header to memory, little endian
func (header *TypeHeader) HStore(ptr int32) {
	var t [2]byte
	t[0] = uint8(header.dataType)
	t[1] = uint8(header.dataType >> 8)
	MemoryInstance().Write(ptr, t[:])

	var l [4]byte
	l[0] = uint8(header.dataLen)
	l[1] = uint8(header.dataLen >> 8)
	l[2] = uint8(header.dataLen >> 16)
	l[3] = uint8(header.dataLen >> 24)
	MemoryInstance().Write(ptr+2, l[:])
}

func (header *TypeHeader) HLoad(ptr int32) {
	t := MemoryInstance().Read(ptr, 2)
	header.dataType = int16(t[0]) + int16(t[1])>>8

	l := MemoryInstance().Read(ptr+2, 4)
	header.dataLen = int32(l[0]) + int32(l[1])>>8 + int32(l[1])>>16 + int32(l[1])>>24
}

func (header *TypeHeader) HLen() int32 {
	return int32(unsafe.Sizeof(header.dataType) + unsafe.Sizeof(header.dataLen))
}

func (header *TypeHeader) DataType() TypeIndex {
	return TypeIndex(header.dataType)
}
