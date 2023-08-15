package runtimetypes

import (
	"errors"
	"github.com/tetratelabs/wazero/api"
)

type Memory interface {
	Write(ptr int32, data []byte)
	Read(ptr int32, size int32) []byte
	Allocate(size int32) (int32, error)
}

type wazeroMemory struct {
	mem   api.Memory
	alloc func(int32) (int32, error)
}

func NewWAZeroMemory(mem api.Memory, alloc func(int32) (int32, error)) Memory {
	return &wazeroMemory{mem: mem, alloc: alloc}
}

func (w *wazeroMemory) Write(ptr int32, data []byte) {
	w.mem.Write(uint32(ptr), data)
}

func (w *wazeroMemory) Read(ptr int32, size int32) []byte {
	data, ok := w.mem.Read(uint32(ptr), uint32(size))
	if !ok {
		return nil
	}

	return data
}

func (w *wazeroMemory) Allocate(size int32) (int32, error) {
	ptr, err := w.alloc(size)
	if err != nil {
		return ptr, err
	}

	if ptr == 0 {
		return ptr, errors.New("memory pointer is 0")
	}
	return ptr, nil
}

type wasmTimeMemory struct {
	data  func() []byte
	alloc func(int32) (int32, error)
}

func NewWASMTimeMemory(data func() []byte, alloc func(int32) (int32, error)) Memory {
	return &wasmTimeMemory{
		data:  data,
		alloc: alloc,
	}
}

// Data return the whole linear memory.
func (m *wasmTimeMemory) Write(ptr int32, data []byte) {
	buf := m.data()
	for i := 0; i < len(data); i++ {
		buf[ptr] = data[i]
		ptr++
	}
}

func (m *wasmTimeMemory) Read(ptr int32, size int32) []byte {
	return m.data()[ptr : ptr+size]
}

// Allocate allocate a contiguous space on linear memory
func (m *wasmTimeMemory) Allocate(size int32) (int32, error) {
	ptr, err := m.alloc(size)
	if err != nil {
		return ptr, err
	}

	if ptr == 0 {
		return ptr, errors.New("memory pointer is 0")
	}
	return ptr, nil
}
