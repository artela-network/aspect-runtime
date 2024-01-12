package types

import (
	"context"

	"github.com/bytecodealliance/wasmtime-go/v14"
)

// IType is the interface of all runtime types
type IType interface {
	// Marshal serialize the type to byte array
	Marshal(value interface{}) []byte

	// Unmarshal desialize the data to the type
	Unmarshal(data []byte) (interface{}, error)
}

type Context struct {
	context.Context

	// memory *Memory
	Instance *wasmtime.Instance
	Store    *wasmtime.Store
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
	}
}

func (c *Context) SetInstance(
	instance *wasmtime.Instance,
	store *wasmtime.Store,
) {
	c.Instance = instance
	c.Store = store
}
