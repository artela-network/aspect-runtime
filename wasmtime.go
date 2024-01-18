package runtime

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"

	"github.com/artela-network/aspect-runtime/types"
	rtypes "github.com/artela-network/aspect-runtime/types"
)

const (
	ExpNameMemory = "memory"
	MaxMemorySize = 32 * 1024 * 1024

	StoreFuel          = 3579139937
	HostApiConsumeFuel = 1000
)

// wasmTimeRuntime is a wrapper for WASMTime runtime
type wasmTimeRuntime struct {
	sync.Mutex

	engine   *wasmtime.Engine
	store    *wasmtime.Store
	module   *wasmtime.Module
	linker   *wasmtime.Linker
	instance *wasmtime.Instance

	apis *HostAPIRegistry
}

func NewWASMTimeRuntime(code []byte, apis *HostAPIRegistry) (out AspectRuntime, err error) {
	watvm := &wasmTimeRuntime{engine: wasmtime.NewEngineWithConfig(defaultWASMTimeConfig())}
	watvm.store = wasmtime.NewStore(watvm.engine)
	err = watvm.store.AddFuel(StoreFuel)
	if err != nil {
		return nil, err
	}
	// limit memory size to 32MB for now
	watvm.store.Limiter(MaxMemorySize, -1, -1, -1, 100)

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
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
			log.Error("failed to create wasm instance", "err", r, "stack", debug.Stack())
		}
	}()
	watvm.instance, err = watvm.linker.Instantiate(watvm.store, watvm.module)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instantiate wasm module")
	}

	// watvm.instance.GetExport(watvm.store, "memory").Memory().Grow(watvm.store, 10)

	return watvm, err
}

// Call wasm
func (w *wasmTimeRuntime) Call(method string, args ...interface{}) (interface{}, error) {
	w.Lock()
	defer w.Unlock()

	run := w.instance.GetFunc(w.store, method)
	if run == nil {
		return "", errors.Errorf("method %s does not exist", method)
	}

	var memory = func() []byte {
		return w.instance.GetExport(w.store, "memory").Memory().UnsafeData(w.store)
	}
	var alloc = func(size int32) (int32, error) {
		memoryAllocator := w.instance.GetFunc(w.store, "allocate")
		if memoryAllocator == nil {
			return 0, errors.New("function 'allocate' does not exist")
		}

		res, err := memoryAllocator.Call(w.store, size)
		if err != nil {
			return 0, err
		}

		return res.(int32), nil
	}
	w.apis.SetInstance(w.instance, w.store)

	ptrs := make([]interface{}, len(args))
	for i, arg := range args {
		var err error
		typeIndex := rtypes.AssertType(arg)
		rtType, err := rtypes.TypeObjectMapping(typeIndex)
		if err != nil {
			return nil, err
		}

		data := rtType.Marshal(arg)
		ptr, err := alloc(int32(len(data)))
		if err != nil {
			return nil, err
		}
		copy(memory()[ptr:], data)
		ptrs[i] = ptr
	}

	val, err := run.Call(w.store, ptrs...)
	if err != nil {
		return "", errors.Wrapf(err, "method %s execution fail", method)
	}

	ptr, ok := val.(int32)
	if !ok {
		return nil, errors.Errorf("read output failed, value: %s", val)
	}

	if ptr == 0 {
		// void functions this will be 0
		return nil, nil
	}

	h := &types.TypeHeader{}
	header := make([]byte, types.HeaderLen)
	copy(header, memory()[ptr:ptr+types.HeaderLen])
	dataType, dataLen, err := h.Unmarshal(header)
	if err != nil {
		return nil, errors.Errorf("read output failed, %v", err)
	}
	resType, err := types.TypeObjectMapping(dataType)
	if err != nil {
		return nil, errors.Errorf("read param failed, %v", err)
	}

	retData := make([]byte, types.HeaderLen+dataLen)
	copy(retData, memory()[ptr:ptr+types.HeaderLen+dataLen])
	return resType.Unmarshal(retData)
}

// ResetStore reset the whole memory of wasm
func (w *wasmTimeRuntime) ResetStore(apis *HostAPIRegistry) (err error) {
	w.Lock()
	defer w.Unlock()

	w.store = wasmtime.NewStore(w.engine)
	err = w.store.AddFuel(StoreFuel)
	if err != nil {
		return err
	}
	w.store.Limiter(MaxMemorySize, -1, -1, -1, 100)
	// w.instance.GetExport(w.store, "memory").Memory().Grow(w.store, 10)

	w.linker = wasmtime.NewLinker(w.engine)

	w.apis = apis
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

	return nil
}

func (w *wasmTimeRuntime) Destroy() {
	w.Lock()
	defer w.Unlock()

	w.apis.SetInstance(nil, nil)
	w.apis = nil

	// Deallocate resources associated with the instance, linker, and store.
	// These components will be reconstructed before the next invocation.
	w.instance = nil
	w.linker = nil
	w.store = nil
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
	config.SetConsumeFuel(true)

	return config
}

func buildModuleMethod(ns NameSpace, method MethodName) string {
	return fmt.Sprintf("%s.%s", ns, method)
}

func buildModuleName(module Module) string {
	return string(module)
}
