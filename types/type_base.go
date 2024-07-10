package types

import (
	"context"
)

// IType is the interface of all runtime types
type IType interface {
	// Marshal serialize the type to byte array
	Marshal(value interface{}) []byte

	// Unmarshal deserialize the data to the type
	Unmarshal(data []byte) (interface{}, error)
}

type HostContext interface {
	RemainingGas() uint64
	SetGas(gas uint64)
	SetVMContext(vmContext VMContext)
}

type VMContext interface {
	context.Context

	WriteMemory(ptr int32, data []byte) error
	ReadMemory(ptr int32, size int32) ([]byte, error)
	AllocMemory(size int32) (int32, error)
	RemainingEVMGas() (int64, error)
	RemainingWASMGas() (int64, error)
	ConsumeWASMGas(gas int64) error
	AddEVMGas(gas int64) error
	SetWASMGas(gas int64) error
	Logger() Logger
}

type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) Logger
}

type AspectRuntime interface {
	Call(method string, gas int64, args ...interface{}) (interface{}, int64, error)
	Destroy()
	ResetStore(ctx context.Context, apis *HostAPIRegistry) error
	Context() context.Context
	Logger() Logger
}

type Validator interface {
	Validate([]byte) error
}
