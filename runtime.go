package runtime

import (
	"log"

	"github.com/pkg/errors"
)

type (
	ApiFunc func(string) string
	// WrapperFunc wraps Api func, input arguments, 0: shared memory ptr, 1: memory length
	WrapperFunc       func(int32) int32
	runtimeBuilder    func(code []byte, apis *HostAPICollection) (out WASMRuntime, err error)
	AspectRuntimeType int

	ArgHelper interface {
		Write(msg string) (int32, error)
		Read(ptr int) (string, error)
	}

	HostAPICollection struct {
		// a function defined in module::namesapce::method
		wrapperFuncs map[string]map[string]map[string]WrapperFunc
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

func RegisterAspectRunner(runnerType AspectRuntimeType, builder runtimeBuilder) {
	builders[runnerType] = builder
}

func NewHostAPICollection() *HostAPICollection {
	return &HostAPICollection{wrapperFuncs: make(map[string]map[string]map[string]WrapperFunc, 0)}
}

func (h *HostAPICollection) AddApi(module, ns, method string, function ApiFunc) {
	if h.wrapperFuncs[module] == nil {
		h.wrapperFuncs[module] = make(map[string]map[string]WrapperFunc, 1)
	}

	if h.wrapperFuncs[module][ns] == nil {
		h.wrapperFuncs[module][ns] = make(map[string]WrapperFunc, 1)
	}

	var fn = func(ptr int32) int32 {
		req, err := h.argHelper.Read(int(ptr))
		if err != nil {
			h.logger.Fatal("reading received message", err)
			return -1
		}
		resp := function(req)

		retP, err := h.argHelper.Write(resp)
		if err != nil {
			h.logger.Fatal("write response failed", err)
			return -1
		}

		return retP
	}
	h.wrapperFuncs[module][ns][method] = fn
}

func (h *HostAPICollection) SetArgHelper(helper ArgHelper) {
	h.argHelper = helper
}

func (h *HostAPICollection) WrapperFuncs() map[string]map[string]map[string]WrapperFunc {
	return h.wrapperFuncs
}

type WASMRuntime interface {
	Call(method string, args ...string) (string, error)
}

// NewAspectRuntime is the factory method for creating aspect runtime
func NewAspectRuntime(runtimeType AspectRuntimeType, code []byte, apis *HostAPICollection) (WASMRuntime, error) {
	builder := builders[runtimeType]
	if builder == nil {
		return nil, errors.New("runtime builder does not exist")
	}

	wasmTime, err := builder(code, apis)
	if err != nil {
		return nil, err
	}

	return wasmTime, nil
}
