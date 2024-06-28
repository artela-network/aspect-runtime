package runtime

import (
	"context"
	"github.com/artela-network/aspect-runtime/types"
	"github.com/artela-network/aspect-runtime/wasmtime"
	"github.com/pkg/errors"
	"sync"
)

type validatorConstructor func(ctx context.Context, logger types.Logger) (types.Validator, error)

var (
	validatorRegistry = map[RuntimeType]validatorConstructor{
		WASM: wasmtime.NewWASMTimeValidator,
	}
	validatorCache sync.Map
)

// NewValidator creates a new wasm code validator
func NewValidator(ctx context.Context, logger types.Logger, runtimeType RuntimeType) (types.Validator, error) {
	if validator, ok := validatorCache.Load(runtimeType); ok {
		return validator.(types.Validator), nil
	}

	constructor := validatorRegistry[runtimeType]
	if constructor == nil {
		return nil, errors.New("runtime engine not support")
	}

	validator, err := constructor(ctx, logger)
	if err != nil {
		return nil, err
	}
	validatorCache.Store(runtimeType, validator)
	logger.Debug("validator created")

	return validator, nil
}
