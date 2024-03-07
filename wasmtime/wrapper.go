package wasmtime

import (
	"fmt"
	"reflect"

	"github.com/bytecodealliance/wasmtime-go/v14"

	"github.com/pkg/errors"

	types "github.com/artela-network/aspect-runtime/types"
)

func Wrap(apiRegistry *types.HostAPIRegistry, hostFunc *types.HostFuncWithGasRule) (interface{}, error) {
	errNotSupport := errors.New("host function not supported")

	fn := hostFunc.Func
	gasRule := hostFunc.GasRule

	t := reflect.TypeOf(fn)
	if t.NumOut() > 2 || t.NumOut() == 0 {
		return nil, errNotSupport
	}

	if t.NumOut() == 1 {
		switch t.NumIn() {
		case 0:
			return func() *wasmtime.Trap {
				_, trap := executeWrapper(apiRegistry.Context(), gasRule, fn)
				return trap
			}, nil

		case 1:
			return func(arg int32) *wasmtime.Trap {
				_, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg)
				return trap
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) *wasmtime.Trap {
				_, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg1, arg2)
				return trap
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) *wasmtime.Trap {
				_, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg1, arg2, arg3)
				return trap
			}, nil
		}
	} else if t.NumOut() == 2 {
		switch t.NumIn() {
		case 0:
			return func() (int32, *wasmtime.Trap) {
				out, trap := executeWrapper(apiRegistry.Context(), gasRule, fn)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 1:
			return func(arg int32) (int32, *wasmtime.Trap) {
				out, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) (int32, *wasmtime.Trap) {
				out, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg1, arg2)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) (int32, *wasmtime.Trap) {
				out, trap := executeWrapper(apiRegistry.Context(), gasRule, fn, arg1, arg2, arg3)
				if trap != nil {
					return 0, trap
				}
				return out[0], nil
			}, nil
		}
	}

	return nil, errNotSupport
}

func executeWrapper(ctx types.Context, gasRule types.HostFuncGasRule, fn interface{}, ptrs ...int32) ([]int32, *wasmtime.Trap) {
	gasRule.SetContext(ctx)

	args, paramSize, err := paramsRead(ctx, ptrs...)
	if paramSize > 0 {
		if err := gasRule.ConsumeGas(paramSize); err != nil {
			return nil, wasmtime.NewTrap(fmt.Sprintf("consume gas failed, %v", err))
		}
	}
	if err != nil {
		return nil, wasmtime.NewTrap(fmt.Sprintf("read params failed"))
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	outPtrs, err := paramListWrite(ctx, res)
	if err != nil {
		return nil, wasmtime.NewTrap(fmt.Sprintf("write params failed, %v", err))
	}
	return outPtrs, nil
}

func paramsRead(ctx types.Context, ptrs ...int32) ([]reflect.Value, int64, error) {
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

func storeValue(ctx types.Context, value reflect.Value) (int32, error) {
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

func paramListWrite(ctx types.Context, values []reflect.Value) ([]int32, error) {
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
