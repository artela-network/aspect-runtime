package runtime

import (
	"log"

	"github.com/pkg/errors"
)

type (
	runtimeBuilder    func(code []byte, apis *HostAPICollection) (out WASMRuntime, err error)
	AspectRuntimeType int

	ArgHelper interface {
		Write(msg string) (int32, error)
		Read(ptr int) (string, error)
	}

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

func RegisterAspectRunner(runnerType AspectRuntimeType, builder runtimeBuilder) {
	builders[runnerType] = builder
}

func NewHostAPICollection() *HostAPICollection {
	return &HostAPICollection{wrapperFuncs: make(map[string]map[string]map[string]interface{}, 0)}
}

func (h *HostAPICollection) AddApi(module, ns, method string, fn interface{}) error {
	wrapper, err := Wrappers(fn)
	if err != nil {
		return err
	}

	if h.wrapperFuncs[module] == nil {
		h.wrapperFuncs[module] = make(map[string]map[string]interface{}, 1)
	}

	if h.wrapperFuncs[module][ns] == nil {
		h.wrapperFuncs[module][ns] = make(map[string]interface{}, 1)
	}

	h.wrapperFuncs[module][ns][method] = wrapper
	return nil
}

func (h *HostAPICollection) SetArgHelper(helper ArgHelper) {
	h.argHelper = helper
}

func (h *HostAPICollection) WrapperFuncs() map[string]map[string]map[string]interface{} {
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
