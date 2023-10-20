package runtime

import (
	"context"
	"log"

	rtypes "github.com/artela-network/runtime/types"
)

type (
	Module     string
	NameSpace  string
	MethodName string
)

type HostAPIRegistry struct {
	// a function defined in Module::Namespace::MethodName
	wrapperFuncs map[Module]map[NameSpace]map[MethodName]interface{}
	//nolint
	logger log.Logger
	ctx    *rtypes.Context
}

func NewHostAPIRegistry() *HostAPIRegistry {
	return &HostAPIRegistry{
		wrapperFuncs: make(map[Module]map[NameSpace]map[MethodName]interface{}, 0),
		ctx:          rtypes.NewContext(context.Background(), nil),
	}
}

func (h *HostAPIRegistry) AddAPI(module Module, ns NameSpace, method MethodName, fn interface{}) error {
	wrapper, err := Wrappers(h.ctx, fn)
	if err != nil {
		return err
	}

	if h.wrapperFuncs[module] == nil {
		h.wrapperFuncs[module] = make(map[NameSpace]map[MethodName]interface{}, 1)
	}

	if h.wrapperFuncs[module][ns] == nil {
		h.wrapperFuncs[module][ns] = make(map[MethodName]interface{}, 1)
	}

	h.wrapperFuncs[module][ns][method] = wrapper
	return nil
}

func (h *HostAPIRegistry) WrapperFuncs() map[Module]map[NameSpace]map[MethodName]interface{} {
	return h.wrapperFuncs
}

func (h *HostAPIRegistry) SetMemory(mem *rtypes.Memory) {
	if h.ctx == nil {
		h.ctx = rtypes.NewContext(context.Background(), mem)
		return
	}
	h.ctx.SetMemory(mem)
}
