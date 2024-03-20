package runtime

import (
	"context"
	"github.com/artela-network/aspect-runtime/instrument"
	"github.com/artela-network/aspect-runtime/types"
	"github.com/artela-network/aspect-runtime/wasmtime"
	"github.com/pkg/errors"
)

type (
	engine func(ctx context.Context, logger types.Logger, code []byte, apis *types.HostAPIRegistry) (out types.AspectRuntime, err error)

	// nolint
	RuntimeType int
)

const (
	WASM RuntimeType = iota
)

var enginePool = map[RuntimeType]engine{
	WASM: wasmtime.NewWASMTimeRuntime,
}

// NewAspectRuntime is the factory method for creating aspect runtime
func NewAspectRuntime(ctx context.Context, logger types.Logger, runtimeType RuntimeType, code []byte, apis *types.HostAPIRegistry) (types.AspectRuntime, error) {
	engine := enginePool[runtimeType]
	if engine == nil {
		return nil, errors.New("runtime engine not support")
	}

	injectedCode, err := instrument.WasmInstrument(code)
	if err != nil {
		return nil, err
	}

	aspectRuntime, err := engine(ctx, logger, injectedCode, apis)
	if err != nil {
		return nil, err
	}

	return aspectRuntime, nil
}
