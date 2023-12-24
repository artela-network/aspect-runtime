package runtime

import (
	"log"
	"reflect"

	"github.com/bytecodealliance/wasmtime-go/v14"

	"github.com/pkg/errors"

	rtypes "github.com/artela-network/aspect-runtime/types"
)

func Wrappers(ctx *rtypes.Context, fn interface{}) (interface{}, error) {
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

func executeWrapper(ctx *rtypes.Context, fn interface{}, ptrs ...int32) {
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Panicln("read params:", err)
	}
	v := reflect.ValueOf(fn)
	v.Call(args)
}

func executeWrapperAndReturn2(ctx *rtypes.Context, fn interface{}, ptrs ...int32) ([]int32, *wasmtime.Trap) {
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Panicln("read params:", err)
		return nil, nil
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	ptr, trap, err := paramListWrite(ctx, res)
	if err != nil {
		log.Panicln("write params:", err)
	}
	return ptr, trap
}

func executeWrapperAndReturn(ctx *rtypes.Context, fn interface{}, ptrs ...int32) int32 {
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

func paramsRead(ctx *rtypes.Context, ptrs ...int32) ([]reflect.Value, error) {
	args := make([]reflect.Value, len(ptrs))

	for i, ptr := range ptrs {
		h := &rtypes.TypeHeader{}
		h.HLoad(ctx, ptr)
		reqType, ok := rtypes.TypeObjectMapping[h.DataType()]
		if !ok {
			log.Printf("type index %d is not valid", h.DataType())
			return nil, errors.New("read param failed")
		}
		reqType.Load(ctx, ptr)
		args[i] = reflect.ValueOf(reqType.Get())
	}
	return args, nil
}

func paramsWrite(ctx *rtypes.Context, values []reflect.Value) (int32, error) {
	if len(values) > 1 {
		return -1, errors.New("values count is expected to 1")
	}

	if len(values) == 1 {
		retIndex := rtypes.AssertType(values[0].Interface())
		resType, ok := rtypes.TypeObjectMapping[retIndex]
		if !ok {
			return 0, errors.Errorf("%v is not supported", values[0].Interface())
		}
		err := resType.Set(values[0].Interface())
		if err != nil {
			return 0, err
		}
		ptr, err := resType.Store(ctx)
		if err != nil {
			return -1, err
		}
		return ptr, nil
	}

	return 0, nil
}

func storeValue(ctx *rtypes.Context, value reflect.Value) (int32, *wasmtime.Trap, error) {
	if value.IsNil() {
		return 0, nil, nil
	}
	err, ok := value.Interface().(error)
	if ok && err != nil {
		return 0, wasmtime.NewTrap(err.Error()), nil
	}

	retIndex := rtypes.AssertType(value.Interface())

	resType, ok := rtypes.TypeObjectMapping[retIndex]
	if !ok {
		return 0, nil, errors.Errorf("%v is not supported", value.Interface())
	}
	err = resType.Set(value.Interface())
	if err != nil {
		return 0, nil, err
	}
	ptr, storeErr := resType.Store(ctx)
	if storeErr != nil {
		return 0, nil, storeErr
	}
	return ptr, nil, nil
}

func paramListWrite(ctx *rtypes.Context, values []reflect.Value) ([]int32, *wasmtime.Trap, error) {
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
