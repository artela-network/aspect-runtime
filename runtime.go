package runtime

import "github.com/pkg/errors"

type (
	engine func(code []byte, apis *HostAPIRegistry) (out AspectRuntime, err error)
	// nolint
	RuntimeType int
)

const (
	WASM RuntimeType = iota
)

var enginePool = map[RuntimeType]engine{
	WASM: NewWASMTimeRuntime,
	// only support wasm now
}

type AspectRuntime interface {
	Call(method string, args ...interface{}) (interface{}, error)
	Destroy()
	ResetStore(apis *HostAPIRegistry) error
}

// NewAspectRuntime is the factory method for creating aspect runtime
func NewAspectRuntime(runtimeType RuntimeType, code []byte, apis *HostAPIRegistry) (AspectRuntime, error) {
	engine := enginePool[runtimeType]
	if engine == nil {
		return nil, errors.New("runtime engine not support")
	}

	aspectRuntime, err := engine(code, apis)
	if err != nil {
		return nil, err
	}

	return aspectRuntime, nil
}
