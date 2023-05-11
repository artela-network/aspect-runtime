package runtime

import (
	"reflect"
	"unsafe"
)

type Memory struct {
	data  func() []byte
	alloc func(int32) (int32, error)
}

func NewMemory(data func() []byte, alloc func(int32) (int32, error)) {
	GlobalMemory = &Memory{
		data:  data,
		alloc: alloc,
	}
}

// Data return the whole linear memory.
func (m *Memory) Write(ptr int32, data []byte) {
	for i := 0; i < len(data); i++ {
		m.data()[ptr] = data[i]
		ptr++
	}
}

func (m *Memory) Read(ptr int32, size int32) []byte {
	return m.data()[ptr : ptr+size]
}

// Allocate allocate a contiguous space on linear memory
func (m *Memory) Allocate(size int32) (int32, error) {
	return m.alloc(size)
}

var GlobalMemory *Memory

func MemoryInstance() *Memory {
	return GlobalMemory
}

// TypeIndex defines the index of runtime type
type TypeIndex int16

const (
	// TypeEmpty
	TypeInt8 TypeIndex = iota
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeBool
	TypeString // string with utf-8 encoder
	TypeByteArray
)

// func (header *TypeHeader) Write

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

type TypeValue interface {
	ByteArray | String
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
	MemoryInstance().Write(ptr+4, t[:])
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

var typeObjectMapping = map[TypeIndex]IType{
	TypeByteArray: NewByteArrary(),
	TypeString:    NewString(),
}

var typeMapping = map[string]TypeIndex{
	"[]byte":  TypeByteArray,
	"string":  TypeString,
	"*string": TypeString,
}
