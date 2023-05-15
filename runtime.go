package runtime

import (
	"github.com/pkg/errors"
)

type (
<<<<<<< HEAD
	engine      func(code []byte, apis *HostAPIRegister) (out AspectRuntime, err error)
	RuntimeType int
=======
	runtimeBuilder    func(code []byte, apis *HostAPICollection) (out WASMRuntime, err error)
	AspectRuntimeType int

	Ctx *Context

	HostAPICollection struct {
		// a function defined in module::namesapce::method
		wrapperFuncs map[string]map[string]map[string]interface{}
		logger       log.Logger
		argHelper    ArgHelper
	}
>>>>>>> b954a1a (add context)
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
