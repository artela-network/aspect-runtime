package runtime

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/artela-network/aspect-runtime/types"
	"github.com/artela-network/aspect-runtime/wasmtime"
	wasm "github.com/bytecodealliance/wasmtime-go/v20"
)

type (
	engine func(ctx context.Context, logger types.Logger, code []byte, apis *types.HostAPIRegistry) (out types.AspectRuntime, err error)

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

	startTime := time.Now()
	injectedCode, err := wasm.Instrument(code)
	if err != nil {
		return nil, err
	}
	logger.Debug("instrumentation done", "duration", time.Since(startTime).String(),
		"beforeSize", len(code),
		"afterSize", len(injectedCode))

	startTime = time.Now()
	aspectRuntime, err := engine(ctx, logger, injectedCode, apis)
	if err != nil {
		return nil, err
	}
	logger.Debug("runtime created", "duration", time.Since(startTime).String())

	return aspectRuntime, nil
}
