package types

type HostFuncGasRule interface {
	SetContext(ctx VMContext)
	ConsumeGas(dataSize int64) error
}

type BaseGasRule struct {
	ctx VMContext
}

func (r *BaseGasRule) SetContext(ctx VMContext) {
	r.ctx = ctx
}

type StaticGasRule struct {
	BaseGasRule

	cost int64
}

func (s *StaticGasRule) ConsumeGas(_ int64) error {
	return s.ctx.ConsumeWASMGas(s.cost)
}

func NewStaticGasRule(cost int64) *StaticGasRule {
	return &StaticGasRule{
		cost: cost,
	}
}

type DynamicGasRule struct {
	BaseGasRule

	multiplier int64
	fixedCost  int64
}

func NewDynamicGasRule(fixedCost int64, multiplier int64) *DynamicGasRule {
	return &DynamicGasRule{
		multiplier: multiplier,
		fixedCost:  fixedCost,
	}
}

func (d *DynamicGasRule) ConsumeGas(dataSize int64) error {
	return d.ctx.ConsumeWASMGas(d.fixedCost + dataSize*d.multiplier)
}
