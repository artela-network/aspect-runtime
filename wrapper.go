package runtime

import (
	"log"
	"reflect"

	rtypes "github.com/artela-network/runtime/types"
	"github.com/pkg/errors"
)

<<<<<<< HEAD
func Wrappers(ctx *rtypes.Context, fn interface{}) (interface{}, error) {
=======
func Wrappers(ctx *Context, fn interface{}) (interface{}, error) {
>>>>>>> b954a1a (add context)
	errNotSupport := errors.New("host function not supported")
	t := reflect.TypeOf(fn)
	if t.NumOut() > 1 {
		return nil, errNotSupport
	}

	switch t.NumIn() {
	case 0:
		return func() int32 {
			return executeWrapper(ctx, fn)
		}, nil

	case 1:
		return func(arg int32) int32 {
			return executeWrapper(ctx, fn, arg)
		}, nil

	case 2:
		return func(arg1 int32, arg2 int32) int32 {
			return executeWrapper(ctx, fn, arg1, arg2)
		}, nil

	case 3:
		return func(arg1 int32, arg2 int32, arg3 int32) int32 {
			return executeWrapper(ctx, fn, arg1, arg2, arg3)
		}, nil
	}

	return nil, errNotSupport
}

<<<<<<< HEAD
func executeWrapper(ctx *rtypes.Context, fn interface{}, ptrs ...int32) int32 {
=======
func executeWrapper(ctx *Context, fn interface{}, ptrs ...int32) int32 {
>>>>>>> b954a1a (add context)
	args, err := paramsRead(ctx, ptrs...)
	if err != nil {
		log.Fatal("read params:", err)
		return -1
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	ptr, err := paramsWrite(ctx, res)
	if err != nil {
		log.Fatal("write params:", err)
	}
	return ptr
}

<<<<<<< HEAD
func paramsRead(ctx *rtypes.Context, ptrs ...int32) ([]reflect.Value, error) {
	args := make([]reflect.Value, len(ptrs))

	for i, ptr := range ptrs {
		h := &rtypes.TypeHeader{}
		h.HLoad(ctx, ptr)
		reqType, ok := rtypes.TypeObjectMapping[h.DataType()]
=======
func paramsRead(ctx *Context, ptrs ...int32) ([]reflect.Value, error) {
	args := make([]reflect.Value, len(ptrs))

	for i, ptr := range ptrs {
		h := &TypeHeader{}
		h.HLoad(ctx, ptr)
		reqType, ok := TypeObjectMapping[h.DataType()]
>>>>>>> b954a1a (add context)
		if !ok {
			log.Fatalf("type index %d is not valid", h.DataType())
			return nil, errors.New("read param failed")
		}
		reqType.Load(ctx, ptr)
		args[i] = reflect.ValueOf(reqType.Get())
	}
	return args, nil
}

<<<<<<< HEAD
func paramsWrite(ctx *rtypes.Context, values []reflect.Value) (int32, error) {
=======
func paramsWrite(ctx *Context, values []reflect.Value) (int32, error) {
>>>>>>> b954a1a (add context)
	if len(values) > 1 {
		return -1, errors.New("values count is expected to 1")
	}

	if len(values) == 1 {
		retIndex := rtypes.AssertType(values[0].Interface())
		resType, ok := rtypes.TypeObjectMapping[retIndex]
		if !ok {
			return 0, errors.Errorf("%v is not supported", values[0])
		}
		resType.Set(values[0].Interface())
		ptr, err := resType.Store(ctx)
		if err != nil {
			return -1, err
		}
		return ptr, nil
	}

	return 0, nil
}
