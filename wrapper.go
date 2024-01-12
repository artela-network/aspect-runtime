package runtime

import (
	"log"
	"reflect"

	"github.com/bytecodealliance/wasmtime-go/v14"

	"github.com/pkg/errors"

	types "github.com/artela-network/aspect-runtime/types"
)

func Wrappers(ctx *types.Context, fn interface{}) (interface{}, error) {
	errNotSupport := errors.New("host function not supported")
	t := reflect.TypeOf(fn)
	if t.NumOut() > 2 {
		return nil, errNotSupport
	}

	if t.NumOut() == 0 {
		switch t.NumIn() {
		case 0:
			return func() {
				executeWrapper(ctx, fn)
			}, nil

		case 1:
			return func(arg int32) {
				executeWrapper(ctx, fn, arg)
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) {
				executeWrapper(ctx, fn, arg1, arg2)
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) {
				executeWrapper(ctx, fn, arg1, arg2, arg3)
			}, nil
		}
	} else if t.NumOut() == 1 {
		switch t.NumIn() {
		case 0:
			return func() int32 {
				return executeWrapperAndReturn(ctx, fn)
			}, nil

		case 1:
			return func(arg int32) int32 {
				return executeWrapperAndReturn(ctx, fn, arg)
			}, nil

		case 2:
			return func(arg1 int32, arg2 int32) int32 {
				return executeWrapperAndReturn(ctx, fn, arg1, arg2)
			}, nil

		case 3:
			return func(arg1 int32, arg2 int32, arg3 int32) int32 {
				return executeWrapperAndReturn(ctx, fn, arg1, arg2, arg3)
			}, nil
		}
	} else if t.NumOut() == 2 {
		switch t.NumIn() {
		case 0:
			return func() int32 {
				return executeWrapperAndReturn(ctx, fn)
			}, nil
		case 1:
			return func(arg int32) (int32, *wasmtime.Trap) {
				return2, trap := executeWrapperAndReturn2(ctx, fn, arg)
				if trap != nil {
					return 0, trap
				}
				return return2[0], nil
			}, nil
		}
	}

	return nil, errNotSupport
}

func executeWrapper(ctx *types.Context, fn interface{}, ptrs ...int32) {
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Panicln("read params:", err)
	}
	v := reflect.ValueOf(fn)
	v.Call(args)
}

func executeWrapperAndReturn2(ctx *types.Context, fn interface{}, ptrs ...int32) ([]int32, *wasmtime.Trap) {
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Panicln("read params:", err)
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	ptr, trap, err := paramListWrite(ctx, res)
	if err != nil {
		log.Panicln("write params list:", err)
	}
	return ptr, trap
}

func executeWrapperAndReturn(ctx *types.Context, fn interface{}, ptrs ...int32) int32 {
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Println("read params:", err)
		return -1
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	ptr, err := paramsWrite(ctx, res)
	if err != nil {
		log.Panicln("write params:", err)
	}
	return ptr
}

var memory = func(instance *wasmtime.Instance, store *wasmtime.Store) []byte {
	return instance.GetExport(store, "memory").Memory().UnsafeData(store)
}
var alloc = func(instance *wasmtime.Instance, store *wasmtime.Store, size int32) (int32, error) {
	memoryAllocator := instance.GetFunc(store, "allocate")
	if memoryAllocator == nil {
		return 0, errors.New("function 'allocate' does not exist")
	}

	res, err := memoryAllocator.Call(store, size)
	if err != nil {
		return 0, err
	}

	return res.(int32), nil
}

func paramsRead(ctx *types.Context, ptrs ...int32) ([]reflect.Value, error) {
	args := make([]reflect.Value, len(ptrs))

	for i, ptr := range ptrs {
		h := &types.TypeHeader{}
		buf := memory(ctx.Instance, ctx.Store)
		header := make([]byte, types.HeaderLen)
		copy(header, buf[ptr:ptr+6])
		dataType, dataLen, err := h.Unmarshal(header)
		if err != nil {
			return nil, err
		}
		reqType, err := types.TypeObjectMapping(dataType)
		if err != nil {
			return nil, err
		}
		reqData := make([]byte, types.HeaderLen+dataLen)
		copy(reqData, buf[ptr:ptr+types.HeaderLen+dataLen])
		value, err := reqType.Unmarshal(reqData)
		if err != nil {
			return nil, err
		}
		args[i] = reflect.ValueOf(value)
	}
	return args, nil
}

func paramsWrite(ctx *types.Context, values []reflect.Value) (int32, error) {
	if len(values) > 1 {
		return -1, errors.New("values count is expected to 1")
	}

	if len(values) == 1 {
		retIndex := types.AssertType(values[0].Interface())
		resType, err := types.TypeObjectMapping(retIndex)
		if err != nil {
			return 0, err
		}
		data := resType.Marshal(values[0].Interface())
		ptr, err := alloc(ctx.Instance, ctx.Store, int32(len(data)))
		if err != nil {
			panic(err)
		}
		buf := memory(ctx.Instance, ctx.Store)
		copy(buf[ptr:], data)

		return ptr, nil
	}

	return 0, nil
}

func storeValue(ctx *types.Context, value reflect.Value) (int32, *wasmtime.Trap, error) {
	if value.IsNil() {
		return 0, nil, nil
	}
	err, ok := value.Interface().(error)
	if ok && err != nil {
		return 0, wasmtime.NewTrap(err.Error()), nil
	}

	retIndex := types.AssertType(value.Interface())

	resType, err := types.TypeObjectMapping(retIndex)
	if err != nil {
		return 0, nil, err
	}

	data := resType.Marshal(value.Interface())
	ptr, err := alloc(ctx.Instance, ctx.Store, int32(len(data)))
	if err != nil {
		panic(err)
	}
	buf := memory(ctx.Instance, ctx.Store)
	copy(buf[ptr:], data)

	return ptr, nil, nil
}

func paramListWrite(ctx *types.Context, values []reflect.Value) ([]int32, *wasmtime.Trap, error) {
	int32Ary := make([]int32, len(values))
	var wasmTrap *wasmtime.Trap
	for i, value := range values {
		i2, trap, err := storeValue(ctx, value)
		if err != nil {
			return nil, trap, err
		}
		if trap != nil {
			wasmTrap = trap
		}
		int32Ary[i] = i2
	}

	return int32Ary, wasmTrap, nil
}
