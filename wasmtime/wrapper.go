package wasmtime

import (
	"fmt"
	"github.com/bytecodealliance/wasmtime-go/v14"
	"reflect"
	"time"

	"github.com/pkg/errors"

	types "github.com/artela-network/aspect-runtime/types"
)

func Wrap(apiRegistry *types.HostAPIRegistry, module types.Module, ns types.NameSpace, method types.MethodName,
	hostFunc *types.HostFuncWithGasRule) (interface{}, error) {
	errNotSupport := errors.New("host function not supported")

	fn := hostFunc.Func
	gasRule := hostFunc.GasRule
	hostCtx := hostFunc.HostContext

	t := reflect.TypeOf(fn)
	if t.NumOut() > 2 || t.NumOut() == 0 {
		return nil, errNotSupport
	}

	if t.NumOut() == 1 {
		switch t.NumIn() {
		case 0:
			return func() *wasmtime.Trap {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				_, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn)
				return trap
			}, nil

		case 1:
			return func(arg int32) *wasmtime.Trap {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				_, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg)
				return trap
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) *wasmtime.Trap {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				_, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg1, arg2)
				return trap
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) *wasmtime.Trap {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				_, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg1, arg2, arg3)
				return trap
			}, nil
		}
	} else if t.NumOut() == 2 {
		switch t.NumIn() {
		case 0:
			return func() (int32, *wasmtime.Trap) {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				out, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 1:
			return func(arg int32) (int32, *wasmtime.Trap) {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				out, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) (int32, *wasmtime.Trap) {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				out, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg1, arg2)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) (int32, *wasmtime.Trap) {
				startTime := time.Now()

				defer func() {
					apiRegistry.Context().Logger().Info("host func done",
						"duration", time.Since(startTime).String(),
						"module", module,
						"namespace", ns,
						"method", method)
				}()

				out, trap := executeWrapper(apiRegistry.Context(), hostCtx, gasRule, fn, arg1, arg2, arg3)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil
		}
	}

	return nil, errNotSupport
}

func executeWrapper(vmCtx types.VMContext, hostCtx types.HostContext, gasRule types.HostFuncGasRule, fn interface{}, ptrs ...int32) ([]int32, *wasmtime.Trap) {
	gasRule.SetContext(vmCtx)

	args, paramSize, err := paramsRead(vmCtx, ptrs...)
	if paramSize > 0 {
		if err := gasRule.ConsumeGas(paramSize); err != nil {
			return nil, wasmtime.NewTrap(err.Error())
		}
	}
	if err != nil {
		return nil, wasmtime.NewTrap(fmt.Sprintf("read params failed"))
	}
	v := reflect.ValueOf(fn)

	// need to sync the gas in vm to host
	remaining, err := vmCtx.RemainingWASMGas()
	if err != nil {
		return nil, wasmtime.NewTrap(fmt.Sprintf("read gas failed, %v", err))
	}

	// NOTE: during the host call, the gas consumed is in EVM metric, so it will be 1000x than WASM metric,
	//       e.g. if the gas in WASM remaining is 11111, when put to EVM is will be 10,
	//       after the host call, the gas put back to WASM should be (10 - EVM Gas Cost) * 1000 + 1111
	remainderGas := remaining % types.EVMGasToWASMGasMultiplier
	hostCtx.SetGas(uint64(remaining) / types.EVMGasToWASMGasMultiplier)

	res := v.Call(args)

	// after the host call we need to sync the gas in host back to vm
	remaining = int64(hostCtx.RemainingGas())
	if err := vmCtx.SetWASMGas(remaining*types.EVMGasToWASMGasMultiplier + remainderGas); err != nil {
		return nil, nil
	}

	outPtrs, err := paramListWrite(vmCtx, res)
	if err != nil && err.Error() == types.OutOfGasError.Error() {
		return nil, wasmtime.NewTrap(types.OutOfGasError.Error())
	}

	if err != nil {
		return nil, wasmtime.NewTrap(fmt.Sprintf("write params failed, %v", err))
	}
	return outPtrs, nil
}

func paramsRead(ctx types.VMContext, ptrs ...int32) ([]reflect.Value, int64, error) {
	args := make([]reflect.Value, len(ptrs))
	paramSize := int64(0)

	if len(args) == 0 {
		return args, paramSize, nil
	}

	for i, ptr := range ptrs {
		h := &types.TypeHeader{}
		header, err := ctx.ReadMemory(ptr, types.HeaderLen)
		if err != nil {
			return nil, paramSize, err
		}

		dataType, dataLen, err := h.Unmarshal(header)
		if err != nil {
			return nil, paramSize, err
		}

		paramSize += int64(dataLen)

		reqType, err := types.TypeObjectMapping(dataType)
		if err != nil {
			return nil, paramSize, err
		}

		reqData, err := ctx.ReadMemory(ptr, types.HeaderLen+dataLen)
		if err != nil {
			return nil, paramSize, err
		}

		value, err := reqType.Unmarshal(reqData)
		if err != nil {
			return nil, paramSize, err
		}
		args[i] = reflect.ValueOf(value)
	}

	return args, paramSize, nil
}

func storeValue(ctx types.VMContext, value reflect.Value) (int32, error) {
	retIndex := types.AssertType(value.Interface())

	resType, err := types.TypeObjectMapping(retIndex)
	if err != nil {
		return 0, err
	}

	data := resType.Marshal(value.Interface())
	ptr, err := ctx.AllocMemory(int32(len(data)))
	if err != nil {
		return 0, err
	}

	if err := ctx.WriteMemory(ptr, data); err != nil {
		return 0, err
	}

	return ptr, nil
}

func paramListWrite(ctx types.VMContext, values []reflect.Value) ([]int32, error) {
	if len(values) == 0 {
		return nil, nil
	}

	lastRetVal := values[len(values)-1].Interface()
	if _, ok := lastRetVal.(error); !ok && lastRetVal != nil {
		return nil, errors.New("invalid host func, last return value must be error")
	} else if ok && lastRetVal != nil {
		return nil, lastRetVal.(error)
	}

	valuesLen := len(values) - 1
	int32Ary := make([]int32, valuesLen)
	for i, value := range values[:valuesLen] {
		i2, err := storeValue(ctx, value)
		if err != nil {
			return nil, err
		}
		int32Ary[i] = i2
	}

	return int32Ary, nil
}
