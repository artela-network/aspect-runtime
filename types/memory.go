package runtimetypes

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
	for i := 0; i < len(data); i++ {
		buf[ptr] = data[i]
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
