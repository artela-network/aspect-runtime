package runtime

import (
	"github.com/pkg/errors"
)

type (
	engine      func(code []byte, apis *HostAPIRegister) (out AspectRuntime, err error)
	RuntimeType int
)

const (
	WASM RuntimeType = iota
)

var (
	enginePool = make(map[RuntimeType]engine)
)

type AspectRuntime interface {
	Call(method string, args ...interface{}) (interface{}, error)
}

// NewAspectRuntime is the factory method for creating aspect runtime
func NewAspectRuntime(runtimeType RuntimeType, code []byte, apis *HostAPIRegister) (AspectRuntime, error) {
	if runtimeType == WASM {
		// only support wasm now
		enginePool[runtimeType] = NewWASMTimeRuntime
	}

	engine := enginePool[runtimeType]
	if engine == nil {
		return nil, errors.New("runtime engine does not exist")
	}

	aspectRuntime, err := engine(code, apis)
	if err != nil {
		return nil, err
	}

	return aspectRuntime, nil
}
