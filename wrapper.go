package runtime

import (
	"log"
	"reflect"

	"github.com/pkg/errors"
)

func Wrappers(fn interface{}) (interface{}, error) {
	errNotSupport := errors.New("host function not supported")
	t := reflect.TypeOf(fn)
	if t.NumOut() > 1 {
		return nil, errNotSupport
	}

	switch t.NumIn() {
	case 0:
		return func() int32 {
			return executeWrapper(fn)
		}, nil

	case 1:
		return func(arg int32) int32 {
			return executeWrapper(fn, arg)
		}, nil

	case 2:
		return func(arg1 int32, arg2 int32) int32 {
			return executeWrapper(fn, arg1, arg2)
		}, nil

	case 3:
		return func(arg1 int32, arg2 int32, arg3 int32) int32 {
			return executeWrapper(fn, arg1, arg2, arg3)
		}, nil
	}

	return nil, errNotSupport
}

func executeWrapper(fn interface{}, ptrs ...int32) int32 {
	args, err := paramsRead(ptrs...)
	if err != nil {
		log.Fatal("read params:", err)
		return -1
	}
	v := reflect.ValueOf(fn)
	res := v.Call(args)
	ptr, err := paramsWrite(res)
	if err != nil {
		log.Fatal("write params:", err)
	}
	return ptr
}

func paramsRead(ptrs ...int32) ([]reflect.Value, error) {
	args := make([]reflect.Value, len(ptrs))

	for i, ptr := range ptrs {
		h := &TypeHeader{}
		h.HLoad(ptr)
		reqType, ok := TypeObjectMapping[h.DataType()]
		if !ok {
			log.Fatalf("type index %d is not valid", h.dataType)
			return nil, errors.New("read param failed")
		}
		reqType.Load(ptr)
		args[i] = reflect.ValueOf(reqType.Get())
	}
	return args, nil
}

func paramsWrite(values []reflect.Value) (int32, error) {
	if len(values) > 1 {
		return -1, errors.New("values count is expected to 1")
	}

	if len(values) == 1 {
		retIndex := AssertType(values[0].Interface())
		resType, ok := TypeObjectMapping[retIndex]
		if !ok {
			return 0, errors.Errorf("%v is not supported", values[0])
		}
		resType.Set(values[0].Interface())
		ptr, err := resType.Store()
		if err != nil {
			return -1, err
		}
		return ptr, nil
	}

	return 0, nil
}
