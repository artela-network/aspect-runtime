package runtime

import (
	"context"
	"fmt"

	rtypes "github.com/artela-network/runtime/types"
	"github.com/bytecodealliance/wasmtime-go/v11"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
)

const (
	ExpNameMemory = "memory"
	MaxMemorySize = 32 * 1024 * 1024
)

// WasmTimeRuntime is a wrapper for WASMTime runtime
type WasmTimeRuntime struct {
	Engine   *wasmtime.Engine
	Store    *wasmtime.Store
	Module   *wasmtime.Module
	Linker   *wasmtime.Linker
	Instance *wasmtime.Instance

	Ctx  *rtypes.Context
	Apis *HostAPIRegistry
}

func NewWASMTimeRuntime(code []byte, apis *HostAPIRegistry) (out AspectRuntime, err error) {
	watvm := &WasmTimeRuntime{Engine: wasmtime.NewEngineWithConfig(defaultWASMTimeConfig())}
	watvm.Store = wasmtime.NewStore(watvm.Engine)
	// limit memory size to 32MB for now
	watvm.Store.Limiter(MaxMemorySize, -1, -1, -1, -1)

	watvm.Module, err = wasmtime.NewModule(watvm.Engine, code)
	if err != nil {
		log.Error("failed to create wasm Module", "err", err, "size", len(code))
		return nil, errors.Wrap(err, "unable create wasm Module")
	}

	// create host api Linker
	watvm.Linker = wasmtime.NewLinker(watvm.Engine)

	// link all host Apis
	watvm.Apis = apis
	if err := watvm.linkToHostFns(); err != nil {
		return nil, err
	}

	// add abort function
	if err := watvm.linkAbort(); err != nil {
		return nil, err
	}

	// instantiate Module and Store
	watvm.Instance, err = watvm.Linker.Instantiate(watvm.Store, watvm.Module)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instantiate wasm Module")
	}

	watvm.SetCtx()
	apis.SetMemory(watvm.Ctx.Memory())

	return watvm, err
}

// Call wasm
func (w *WasmTimeRuntime) Call(method string, args ...interface{}) (interface{}, error) {
	run := w.Instance.GetFunc(w.Store, method)
	if run == nil {
		return "", errors.Errorf("method %s does not exist", method)
	}

	// ptrs := make([]interface{}, len(args))
	// for i, arg := range args {
	// 	var err error
	// 	typeIndex := rtypes.AssertType(arg)
	// 	rtType, ok := rtypes.TypeObjectMapping[typeIndex]
	// 	if !ok {
	// 		return nil, errors.Errorf("%v is not supported", arg)
	// 	}
	// 	if err := rtType.Set(arg); err != nil {
	// 		return nil, errors.Wrapf(err, "set argument %+v", arg)
	// 	}
	// 	ptrs[i], err = rtType.Store(w.Ctx)
	// 	if err != nil {
	// 		return "", errors.Wrapf(err, "write memory %+v", arg)
	// 	}
	// }

	val, err := run.Call(w.Store, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "method %s execution fail", method)
	}

	if val == nil {
		return nil, nil
	}

	ptr, ok := val.(int32)
	if !ok {
		return nil, errors.Errorf("read output failed, value: %s", val)
	}

	h := &rtypes.TypeHeader{}
	h.HLoad(w.Ctx, ptr)
	resType, ok := rtypes.TypeObjectMapping[h.DataType()]
	if !ok {
		return nil, errors.Errorf("read param failed, type %d not found", resType)
	}

	resType.Load(w.Ctx, ptr)
	return resType.Get(), nil
}

// ResetStore reset the whole memory of wasm
func (w *WasmTimeRuntime) ResetStore(apis *HostAPIRegistry) (err error) {

	w.Instance, err = w.Linker.Instantiate(w.Store, w.Module)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate wasm Module")
	}

	// set context
	w.SetCtx()

	// set memory Instance to Apis, for host function Ctx.
	w.Apis.SetMemory(w.Ctx.Memory())

	return nil
}

func (w *WasmTimeRuntime) Destroy() {
	w.Apis.SetMemory(nil)
	// w.Linker.Define(w.Store, "", ExpNameMemory, nil)
	// w.Apis = nil
}

func (w *WasmTimeRuntime) linkToHostFns() error {
	for module, namespaces := range w.Apis.WrapperFuncs() {
		for ns, methods := range namespaces {
			for method, function := range methods {
				// create a function wrapper for our "hostapi" on the go side
				item := wasmtime.WrapFunc(w.Store, function)

				// create Linker with host function injected
				if err := w.Linker.Define(
					w.Store,
					buildModuleName(module),
					buildModuleMethod(ns, method),
					item,
				); err != nil {
					return errors.Wrapf(
						err, "unable to link host api %s:%s.%s", module, ns, method,
					)
				}
			}
		}
	}
	return nil
}

func (w *WasmTimeRuntime) SetCtx() {
	w.Ctx = rtypes.NewContext(
		// The context here is reserved for future functionalities,
		// such as logging, tracing, and other purposes.
		context.Background(),
		rtypes.NewMemory(
			func() []byte {
				return w.Instance.GetExport(w.Store, ExpNameMemory).Memory().UnsafeData(w.Store)
			},
			func(size int32) (int32, error) {
				memoryAllocator := w.Instance.GetFunc(w.Store, "allocate")
				if memoryAllocator == nil {
					return 0, errors.New("function 'allocate' does not exist")
				}

				res, err := memoryAllocator.Call(w.Store, size)
				if err != nil {
					return 0, err
				}

				return res.(int32), nil
			},
		),
	)
}

func (w *WasmTimeRuntime) linkAbort() error {
	abort := wasmtime.WrapFunc(w.Store, func(a, b, c, d int32) {
		log.Debug("abort!")
	})
	if err := w.Linker.Define(w.Store, "env", "abort", abort); err != nil {
		return errors.Wrapf(err, "unable to link to abort")
	}
	return nil
}

// defaultWASMTimeConfig provides a default wasmtime config for the runner.
// TODO: currently this is just a very early version, should investigate deeper for each config option.
func defaultWASMTimeConfig() *wasmtime.Config {
	config := wasmtime.NewConfig()
	// we don't quite need this, discuss latefr
	// config.SetWasmSIMD(false)
	// affect execution certainty, disable
	config.SetWasmThreads(false)
	// multi-value return is useful, should be enabled
	config.SetWasmMultiValue(true)
	// need to run benchmarks on this and adjust later
	// config.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	// disable multi-memory by default
	config.SetWasmMultiMemory(false)

	config.SetWasmSIMD(true)
	// config.SetCraneliftOptLevel(wasmtime.OptLevelSpeed)
	config.SetStrategy(wasmtime.StrategyAuto)

	return config
}

func buildModuleMethod(ns Namesapce, method MethodName) string {
	return fmt.Sprintf("%s.%s", ns, method)
}

func buildModuleName(module Module) string {
	return fmt.Sprintf("%s", module)
}
