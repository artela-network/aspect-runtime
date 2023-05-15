package runtimetypes

type Memory struct {
	data  func() []byte
	alloc func(int32) (int32, error)
}

<<<<<<< HEAD:types/memory.go
func NewMemory(data func() []byte, alloc func(int32) (int32, error)) *Memory {
	return &Memory{
=======
// // TODO (Eric) remove global memory, this is unuseable for multi-thread to run wasm.
// var globalMemory *Memory

// func MemoryInstance() *Memory {
// 	return globalMemory
// }

func NewMemory(data func() []byte, alloc func(int32) (int32, error)) Memory {
	return Memory{
>>>>>>> b954a1a (add context):memory.go
		data:  data,
		alloc: alloc,
	}
}

// Data return the whole linear memory.
func (m Memory) Write(ptr int32, data []byte) {
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
