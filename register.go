package runtime

import (
	"context"
	"log"

	rtypes "github.com/artela-network/runtime/types"
)

type HostAPIRegister struct {
	// a function defined in module::namesapce::method
	wrapperFuncs map[string]map[string]map[string]interface{}
	logger       log.Logger
	ctx          *rtypes.Context
}

func NewHostAPIRegister() *HostAPIRegister {
	return &HostAPIRegister{
		wrapperFuncs: make(map[string]map[string]map[string]interface{}, 0),
		ctx:          rtypes.NewContext(context.Background(), nil),
	}
}

func (h *HostAPIRegister) AddApi(module, ns, method string, fn interface{}) error {
	wrapper, err := Wrappers(h.ctx, fn)
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

func (h *HostAPIRegister) WrapperFuncs() map[string]map[string]map[string]interface{} {
	return h.wrapperFuncs
}

func (h *HostAPIRegister) SetMemory(mem *rtypes.Memory) {
	if h.ctx == nil {
		h.ctx = rtypes.NewContext(context.Background(), mem)
		return
	}
	h.ctx.SetMemory(mem)
}
