package types

type (
	Module     string
	NameSpace  string
	MethodName string
)

type HostFuncWrapper func(ctx VMContext, module Module, ns NameSpace, method MethodName, hostFunc *HostFuncWithGasRule) (interface{}, error)

type HostFuncWithGasRule struct {
	Func        interface{}
	GasRule     HostFuncGasRule
	HostContext HostContext
}

type HostAPIRegistry struct {
	// a function defined in Module::Namespace::MethodName
	wrapperFuncs map[Module]map[NameSpace]map[MethodName]interface{}

	hostFuncWrapper HostFuncWrapper

	// nolint
	ctx VMContext
}

func NewHostAPIRegistry(hostFuncWrapper HostFuncWrapper) *HostAPIRegistry {
	return &HostAPIRegistry{
		wrapperFuncs:    make(map[Module]map[NameSpace]map[MethodName]interface{}),
		hostFuncWrapper: hostFuncWrapper,
	}
}

func (h *HostAPIRegistry) Context() VMContext {
	return h.ctx
}

func (h *HostAPIRegistry) AddAPI(module Module, ns NameSpace, method MethodName, hostFunc *HostFuncWithGasRule) error {
	wrapper, err := h.hostFuncWrapper(h.ctx, module, ns, method, hostFunc)
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

func (h *HostAPIRegistry) SetContext(ctx VMContext) {
	h.ctx = ctx
}

func (h *HostAPIRegistry) Destroy() {
	h.wrapperFuncs = nil
}
