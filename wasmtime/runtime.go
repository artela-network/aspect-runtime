package wasmtime

import (
	"fmt"
	"reflect"

	"github.com/artela-network/runtime"
	"github.com/bytecodealliance/wasmtime-go/v7"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
)

const (
	ExpNameMemory = "memory"
)

// wasmTimeRuntime is a wrapper for WASMTime runtime
type wasmTimeRuntime struct {
	engine   *wasmtime.Engine
	store    *wasmtime.Store
	module   *wasmtime.Module
	linker   *wasmtime.Linker
	instance *wasmtime.Instance

	memory *Memory
}

func NewWASMTimeRuntime(code []byte, apis *runtime.HostAPICollection) (out runtime.WASMRuntime, err error) {
	watvm := &wasmTimeRuntime{engine: wasmtime.NewEngineWithConfig(defaultWASMTimeConfig())}
	watvm.store = wasmtime.NewStore(watvm.engine)
	watvm.module, err = wasmtime.NewModule(watvm.engine, code)
	if err != nil {
		return nil, errors.Wrap(err, "unable create wasm module")
	}

	// create host api linker
	watvm.linker = wasmtime.NewLinker(watvm.engine)

	// link all host apis
	for module, namespaces := range apis.WrapperFuncs() {
		for ns, methods := range namespaces {
			for method, function := range methods {
				// create a function wrapper for our "hostapi" on the go side
				item := wasmtime.WrapFunc(watvm.store, function)

				// create linker with host function injected
				if err = watvm.linker.Define(watvm.store, module, buildModuleMethod(ns, method), item); err != nil {
					return nil, errors.Wrapf(err, "unable to link host api %s:%s.%s", module, ns, method)
				}
			}
		}
	}

	// add abort function
	abort := wasmtime.WrapFunc(watvm.store, func(a, b, c, d int32) {
		log.Debug("abort!")
	})
	if err = watvm.linker.Define(watvm.store, "env", "abort", abort); err != nil {
		return nil, errors.Wrapf(err, "unable to link to abort")
	}

	// log function
	log := wasmtime.WrapFunc(watvm.store, func(ptr int32) {
		fmt.Println(string(watvm.memory.data[ptr : ptr+100]))
	})
	if err = watvm.linker.Define(watvm.store, "index", "test.log", log); err != nil {
		return nil, errors.Wrapf(err, "unable to link to abort")
	}

	// instantiate module and store
	if watvm.instance, err = watvm.linker.Instantiate(watvm.store, watvm.module); err != nil {
		return nil, errors.Wrap(err, "unable to instantiate wasm module")
	}

	watvm.memory = &Memory{
		data: watvm.instance.GetExport(watvm.store, ExpNameMemory).Memory().UnsafeData(watvm.store),
		allocate: func(size int32) (int32, error) {
			memoryAllocator := watvm.instance.GetFunc(watvm.store, "allocate")
			if memoryAllocator == nil {
				return 0, errors.Wrap(err, "unable to allocate memory in wasm")
			}

			res, err := memoryAllocator.Call(watvm.store, size)
			if err != nil {
				return 0, err
			}

			return res.(int32), nil
		},
		refresh: func(m *Memory) {
			m.data = watvm.instance.GetExport(watvm.store, ExpNameMemory).Memory().UnsafeData(watvm.store)
		},
	}

	runtime.NewMemory(
		func() []byte {
			return watvm.instance.GetExport(watvm.store, ExpNameMemory).Memory().UnsafeData(watvm.store)
		},
		func(size int32) (int32, error) {
			return watvm.memory.allocate(size)
		},
	)
	apis.SetArgHelper(watvm.memory)

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
		typeIndex, ok := runtime.TypeMapping[reflect.TypeOf(arg).Name()]
		if !ok {
			return nil, errors.Errorf("%v is not supported", arg)
		}
		rtType := runtime.TypeObjectMapping[typeIndex]
		if err := rtType.Set(arg); err != nil {
			return nil, errors.Wrapf(err, "set argument %+v", arg)
		}
		ptrs[i], err = rtType.Store()
		if err != nil {
			return "", errors.Wrapf(err, "write memory %+v", arg)
		}
	}

	fmt.Println("call ptrs: ", ptrs)
	val, err := run.Call(w.store, ptrs...)
	if err != nil {
		return "", errors.Wrapf(err, "method %s execution fail", method)
	}

	ptr, ok := val.(int32)
	if !ok {
		return nil, errors.Errorf("read output failed, value: %s", val)
	}

	h := &runtime.TypeHeader{}
	h.HLoad(ptr)
	resType, ok := runtime.TypeObjectMapping[h.DataType()]
	if !ok {
		return nil, errors.Errorf("read param failed, type %d not found", resType)
	}

	resType.Load(ptr)
	return resType.Get(), nil
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

	return config
}

func buildModuleMethod(ns, method string) string {
	return fmt.Sprintf("%s.%s", ns, method)
}
