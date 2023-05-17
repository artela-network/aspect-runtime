package runtime

import (
	"log"

	"github.com/pkg/errors"
)

type (
	runtimeBuilder    func(code []byte, apis *HostAPICollection) (out AspectRuntime, err error)
	AspectRuntimeType int

	Ctx *Context

	HostAPICollection struct {
		// a function defined in module::namesapce::method
		wrapperFuncs map[string]map[string]map[string]interface{}
		logger       log.Logger
		argHelper    ArgHelper
	}
)

const (
	WASMTime AspectRuntimeType = iota
)

var (
	builders = make(map[AspectRuntimeType]runtimeBuilder)
)

type AspectRuntime interface {
	Call(method string, args ...interface{}) (interface{}, error)
}

// NewAspectRuntime is the factory method for creating aspect runtime
func NewAspectRuntime(runtimeType AspectRuntimeType, code []byte, apis *HostAPICollection) (AspectRuntime, error) {
	if runtimeType == WASMTime {
		// only support wasm now
		builders[runtimeType] = NewWASMTimeRuntime
	}

	builder := builders[runtimeType]
	if builder == nil {
		return nil, errors.New("runtime builder does not exist")
	}

	aspectRuntime, err := builder(code, apis)
	if err != nil {
		return nil, err
	}

	return aspectRuntime, nil
}
