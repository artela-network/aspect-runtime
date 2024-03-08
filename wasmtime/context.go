package wasmtime

import (
	"context"
	"errors"
	"github.com/artela-network/aspect-runtime/types"
	"github.com/bytecodealliance/wasmtime-go/v14"
)

type Context struct {
	context.Context

	// memory *Memory
	Instance *wasmtime.Instance
	Store    *wasmtime.Store

	gasCounterGlobal *wasmtime.Global
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
	}
}

func (c *Context) WriteMemory(ptr int32, data []byte) error {
	mem, err := c.memory()
	if err != nil {
		return err
	}

	if (ptr + int32(len(data))) > int32(len(mem)) {
		return errors.New("memory out of bound")
	}

	copy(mem[ptr:], data)
	return nil
}

func (c *Context) ReadMemory(ptr int32, size int32) ([]byte, error) {
	mem, err := c.memory()
	if err != nil {
		return nil, err
	}

	if (ptr + size) > int32(len(mem)) {
		return nil, errors.New("memory out of bound")
	}

	dataCopy := make([]byte, size)
	copy(dataCopy, mem[ptr:ptr+size])

	return dataCopy, nil
}

func (c *Context) Reset() {
	c.Instance = nil
	c.Store = nil
	c.gasCounterGlobal = nil
}

func (c *Context) memory() ([]byte, error) {
	memExport := c.Instance.GetExport(c.Store, "memory")
	if memExport == nil {
		return nil, errors.New("memory export not found")
	}

	mem := memExport.Memory()
	if mem == nil {
		return nil, errors.New("memory export is not a memory")
	}

	return mem.UnsafeData(c.Store), nil
}

func (c *Context) AllocMemory(size int32) (int32, error) {
	memoryAllocator := c.Instance.GetFunc(c.Store, "allocate")
	if memoryAllocator == nil {
		return 0, errors.New("function 'allocate' does not exist")
	}

	res, err := memoryAllocator.Call(c.Store, size)
	if err != nil {
		return 0, err
	}

	return res.(int32), nil
}

// gasCounter get the gas counter from wasm,
// "__gas_counter__" global variable is an i64 injected by wasm instrument lib
func (c *Context) gasCounter() (*wasmtime.Global, error) {
	if c.gasCounterGlobal != nil {
		return c.gasCounterGlobal, nil
	}

	export := c.Instance.GetExport(c.Store, "__gas_counter__")
	if export == nil {
		return nil, errors.New("gas counter not exported")
	}

	c.gasCounterGlobal = export.Global()
	return c.gasCounterGlobal, nil
}

func (c *Context) RemainingGas() (int64, error) {
	gasCounter, err := c.gasCounter()
	if err != nil {
		return 0, err
	}

	leftover := gasCounter.Get(c.Store).I64()
	if leftover < 0 {
		return 0, types.OutOfGasError
	}

	return leftover / types.EVMGasToWASMGasMultiplier, nil
}

func (c *Context) ConsumeGas(gas int64) error {
	gasCounter, err := c.gasCounter()
	if err != nil {
		return err
	}

	leftover := gasCounter.Get(c.Store).I64()
	if leftover < gas {
		return errors.New("out of gas")
	}

	if err := gasCounter.Set(c.Store, wasmtime.ValI64(leftover-gas)); err != nil {
		return err
	}

	return nil
}

func (c *Context) AddGas(gas int64) error {
	// check overflow
	if gas > types.MaxGas {
		return errors.New("gas overflow")
	}

	gasCounter, err := c.gasCounter()
	if err != nil {
		return err
	}

	if err := gasCounter.Set(c.Store, wasmtime.ValI64(gas*types.EVMGasToWASMGasMultiplier)); err != nil {
		return err
	}

	return nil
}
