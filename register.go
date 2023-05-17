package runtime

import (
	"context"
	"log"
)

type HostAPICollection struct {
	// a function defined in module::namesapce::method
	wrapperFuncs map[string]map[string]map[string]interface{}
	logger       log.Logger
	ctx          *Context
}

func NewHostAPICollection() *HostAPICollection {
	return &HostAPICollection{
		wrapperFuncs: make(map[string]map[string]map[string]interface{}, 0),
		ctx:          NewContext(context.Background(), nil),
	}
}

func (h *HostAPICollection) AddApi(module, ns, method string, fn interface{}) error {
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

func (h *HostAPICollection) WrapperFuncs() map[string]map[string]map[string]interface{} {
	return h.wrapperFuncs
}

func (h *HostAPICollection) SetMemory(mem *Memory) {
	if h.ctx == nil {
		h.ctx = NewContext(context.Background(), mem)
		return
	}
	h.ctx.memory = mem
}
