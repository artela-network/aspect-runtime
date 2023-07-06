package runtime

import (
	"context"
	"fmt"

	rtypes "github.com/artela-network/runtime/types"
	"github.com/bytecodealliance/wasmtime-go/v9"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
)

const (
	ExpNameMemory = "memory"
	MaxMemorySize = 32 * 1024 * 1024
)

// wasmTimeRuntime is a wrapper for WASMTime runtime
type wasmTimeRuntime struct {
	engine   *wasmtime.Engine
	store    *wasmtime.Store
	module   *wasmtime.Module
	linker   *wasmtime.Linker
	instance *wasmtime.Instance

	ctx  *rtypes.Context
	apis *HostAPIRegistry
}

func NewWASMTimeRuntime(code []byte, apis *HostAPIRegistry) (out AspectRuntime, err error) {
	watvm := &wasmTimeRuntime{engine: wasmtime.NewEngineWithConfig(defaultWASMTimeConfig())}
	watvm.store = wasmtime.NewStore(watvm.engine)
	// limit memory size to 32MB for now
	watvm.store.Limiter(MaxMemorySize, -1, -1, -1, -1)

	watvm.module, err = wasmtime.NewModule(watvm.engine, code)
	if err != nil {
		log.Error("failed to create wasm module", "err", err, "size", len(code))
		return nil, errors.Wrap(err, "unable create wasm module")
	}

	// create host api linker
	watvm.linker = wasmtime.NewLinker(watvm.engine)

	// link all host apis
	watvm.apis = apis
	if err := watvm.linkToHostFns(); err != nil {
		return nil, err
	}

	// add abort function
	if err := watvm.linkAbort(); err != nil {
		return nil, err
	}

	// instantiate module and store
	watvm.instance, err = watvm.linker.Instantiate(watvm.store, watvm.module)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instantiate wasm module")
	}

	watvm.setCtx()
	apis.SetMemory(watvm.ctx.Memory())

	return watvm, err
}

// Call wasm
func (w *wasmTimeRuntime) Call(method string, args ...interface{}) (interface{}, error) {
	run := w.instance.GetFunc(w.store, method)
	if run == nil {
		return "", errors.Errorf("method %s does not exist", method)
	}

	ptrs := make([]interface{}, len(args))
	for i, arg := range args {
		var err error
		typeIndex := rtypes.AssertType(arg)
		rtType, ok := rtypes.TypeObjectMapping[typeIndex]
		if !ok {
			return nil, errors.Errorf("%v is not supported", arg)
		}
		if err := rtType.Set(arg); err != nil {
			return nil, errors.Wrapf(err, "set argument %+v", arg)
		}
		ptrs[i], err = rtType.Store(w.ctx)
		if err != nil {
			return "", errors.Wrapf(err, "write memory %+v", arg)
		}
	}

	val, err := run.Call(w.store, ptrs...)
	if err != nil {
		return "", errors.Wrapf(err, "method %s execution fail", method)
	}

	ptr, ok := val.(int32)
	if !ok {
		return nil, errors.Errorf("read output failed, value: %s", val)
	}

	h := &rtypes.TypeHeader{}
	h.HLoad(w.ctx, ptr)
	resType, ok := rtypes.TypeObjectMapping[h.DataType()]
	if !ok {
		return nil, errors.Errorf("read param failed, type %d not found", resType)
	}

	resType.Load(w.ctx, ptr)
	return resType.Get(), nil
}

// ResetStore reset the whole memory of wasm
func (w *wasmTimeRuntime) ResetStore() (err error) {
	w.store = wasmtime.NewStore(w.engine)

	w.linker = wasmtime.NewLinker(w.engine)

	// reset link all host apis with new store
	if err := w.linkToHostFns(); err != nil {
		return err
	}

	// add abort function
	if err := w.linkAbort(); err != nil {
		return err
	}

	w.instance, err = w.linker.Instantiate(w.store, w.module)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate wasm module")
	}

	// set context
	w.setCtx()

	// set memory instance to apis, for host function ctx.
	w.apis.SetMemory(w.ctx.Memory())

	return nil
}

func (w *wasmTimeRuntime) Destroy() {
	w.apis.SetMemory(nil)
	w.apis = nil
}

func (w *wasmTimeRuntime) linkToHostFns() error {
	for module, namespaces := range w.apis.WrapperFuncs() {
		for ns, methods := range namespaces {
			for method, function := range methods {
				// create a function wrapper for our "hostapi" on the go side
				item := wasmtime.WrapFunc(w.store, function)

				// create linker with host function injected
				if err := w.linker.Define(
					w.store,
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

func (w *wasmTimeRuntime) setCtx() {
	w.ctx = rtypes.NewContext(
		// The context here is reserved for future functionalities,
		// such as logging, tracing, and other purposes.
		context.Background(),
		rtypes.NewMemory(
			func() []byte {
				return w.instance.GetExport(w.store, ExpNameMemory).Memory().UnsafeData(w.store)
			},
			func(size int32) (int32, error) {
				memoryAllocator := w.instance.GetFunc(w.store, "allocate")
				if memoryAllocator == nil {
					return 0, errors.New("function 'allocate' does not exist")
				}

				res, err := memoryAllocator.Call(w.store, size)
				if err != nil {
					return 0, err
				}

				return res.(int32), nil
			},
		),
	)
}

func (w *wasmTimeRuntime) linkAbort() error {
	abort := wasmtime.WrapFunc(w.store, func(a, b, c, d int32) {
		log.Debug("abort!")
	})
	if err := w.linker.Define(w.store, "env", "abort", abort); err != nil {
		return errors.Wrapf(err, "unable to link to abort")
	}
	return nil
}

// defaultWASMTimeConfig provides a default wasmtime config for the runner.
// TODO: currently this is just a very early version, should investigate deeper for each config option.
func defaultWASMTimeConfig() *wasmtime.Config {
	config := wasmtime.NewConfig()
	// we don't quite need this, discuss latefr
	config.SetWasmSIMD(false)
	// affect execution certainty, disable
	config.SetWasmThreads(false)
	// multi-value return is useful, should be enabled
	config.SetWasmMultiValue(true)
	// need to run benchmarks on this and adjust later
	config.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	// disable multi-memory by default
	config.SetWasmMultiMemory(false)
	// enable debug mode for now
	config.SetDebugInfo(true)

	return config
}

func buildModuleMethod(ns Namesapce, method MethodName) string {
	return fmt.Sprintf("%s.%s", ns, method)
}

func buildModuleName(module Module) string {
	return fmt.Sprintf("%s", module)
}
