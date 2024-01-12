package types

import "errors"

type Memory struct {
	data  func() []byte
	alloc func(int32) (int32, error)
}

func NewMemory(data func() []byte, alloc func(int32) (int32, error)) *Memory {
	return &Memory{
		data:  data,
		alloc: alloc,
	}
}

// Data return the whole linear memory.
func (m *Memory) Write(ptr int32, data []byte) {
	buf := m.data()
	copy(buf[ptr:], data[:])
}

func (m *Memory) Read(ptr int32, size int32) []byte {
	data := m.data()[ptr : ptr+size]
	copied := make([]byte, len(data))
	copy(copied, data)
	return copied
}

// Allocate allocate a contiguous space on linear memory
func (m *Memory) Allocate(size int32) (int32, error) {
	ptr, err := m.alloc(size)
	if err != nil {
		return ptr, err
	}

	if ptr == 0 {
		return ptr, errors.New("memory pointer is 0")
	}
	return ptr, nil
}
