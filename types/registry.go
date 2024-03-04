package types

type (
	Module     string
	NameSpace  string
	MethodName string
)

type HostFuncWrapper func(apiRegistry *HostAPIRegistry, fn interface{}, gasRule HostFuncGasRule) (interface{}, error)

type HostFuncWithGasRule struct {
	Func    interface{}
	GasRule HostFuncGasRule
}

type HostAPIRegistry struct {
	// a function defined in Module::Namespace::MethodName
	wrapperFuncs map[Module]map[NameSpace]map[MethodName]interface{}

	hostFuncWrapper HostFuncWrapper

	// nolint
	ctx Context
}

func NewHostAPIRegistry(hostFuncWrapper HostFuncWrapper) *HostAPIRegistry {
	return &HostAPIRegistry{
		wrapperFuncs:    make(map[Module]map[NameSpace]map[MethodName]interface{}),
		hostFuncWrapper: hostFuncWrapper,
	}
}

func (h *HostAPIRegistry) Context() Context {
	return h.ctx
}

func (h *HostAPIRegistry) AddAPI(module Module, ns NameSpace, method MethodName, fn interface{}, gasRule HostFuncGasRule) error {
	wrapper, err := h.hostFuncWrapper(h, fn, gasRule)
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

func (h *HostAPIRegistry) SetContext(ctx Context) {
	h.ctx = ctx
}
