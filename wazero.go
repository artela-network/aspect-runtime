package runtime

import (
	"context"
	"fmt"
	rtypes "github.com/artela-network/runtime/types"
	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type wazeroRuntime struct {
	ctx          *rtypes.Context
	rt           wazero.Runtime
	apis         *HostAPIRegistry
	compiled     wazero.CompiledModule
	moduleConfig wazero.ModuleConfig
	instance     api.Module
}

// New creates a new Wazero Runtime
func NewWazeroRuntime(code []byte, apis *HostAPIRegistry) (out AspectRuntime, err error) {
	ctx := context.Background()

	config := wazero.NewRuntimeConfig().
		WithCompilationCache(wazero.NewCompilationCache()).
		WithCoreFeatures(api.CoreFeaturesV2)

	wvm := &wazeroRuntime{rt: wazero.NewRuntimeWithConfig(ctx, config)}

	wvm.compiled, err = wvm.rt.CompileModule(ctx, code)
	if err != nil {
		return
	}

	wvm.moduleConfig = wazero.NewModuleConfig()

	wvm.apis = apis

	// link host apis
	hostAPIs := apis.WrapperFuncs()
	for moduleName, module := range hostAPIs {
		hostModuleBuilder := wvm.rt.NewHostModuleBuilder(string(moduleName))
		for nsName, ns := range module {
			for methodName, method := range ns {
				hostModuleBuilder.
					NewFunctionBuilder().
					WithFunc(method).
					Export(buildModuleMethod(nsName, methodName))
			}
		}
		if _, err = hostModuleBuilder.Instantiate(ctx); err != nil {
			return
		}
	}

	// link env
	if _, err = wvm.rt.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(a, b, c, d int32) {
			fmt.Println("abort!")
		}).
		Export("abort").
		Instantiate(ctx); err != nil {
		return
	}

	if wvm.instance, err = wvm.rt.InstantiateModule(ctx, wvm.compiled, wvm.moduleConfig); err != nil {
		return
	}

	// set context
	wvm.setCtx()

	// set memory instance to apis, for host function ctx.
	wvm.apis.SetMemory(wvm.ctx.Memory())

	return wvm, nil
}

func (w *wazeroRuntime) Call(method string, args ...interface{}) (interface{}, error) {
	// No need for GetFunc and Call separate as in Wasmtime
	// In Wazero we can directly get the exported function and call it

	u64Args := make([]uint64, 0, len(args))
	for _, arg := range args {
		switch value := arg.(type) {
		case string:
			ptr, err := w.ctx.Memory().Allocate(int32(len(value)))
			if err != nil {
				return nil, err
			}
			if ok := w.instance.Memory().WriteString(uint32(ptr), value); !ok {
				return nil, errors.New("write to mem fail")
			}
			u64Args = append(u64Args, uint64(ptr))
			break
		case []byte:
			ptr, err := w.ctx.Memory().Allocate(int32(len(value)))
			if err != nil {
				return nil, err
			}
			if ok := w.instance.Memory().WriteString(uint32(ptr), string(value)); !ok {
				return nil, errors.New("write to mem fail")
			}
			u64Args = append(u64Args, uint64(ptr))
			break
		case int:
			u64Args = append(u64Args, uint64(value))
			break
		case int8:
			u64Args = append(u64Args, uint64(value))
			break
		case uint8:
			u64Args = append(u64Args, uint64(value))
			break
		case int16:
			u64Args = append(u64Args, uint64(value))
			break
		case uint16:
			u64Args = append(u64Args, uint64(value))
			break
		case int32:
			u64Args = append(u64Args, uint64(value))
			break
		case uint32:
			u64Args = append(u64Args, uint64(value))
			break
		case int64:
			u64Args = append(u64Args, uint64(value))
			break
		case uint64:
			u64Args = append(u64Args, value)
			break
		default:
			return nil, errors.New("unsupported type")
		}
	}

	res, err := w.instance.ExportedFunction(method).Call(w.ctx, u64Args...)
	if err != nil {
		return nil, err
	}

	if len(res) <= 0 {
		return nil, nil
	}

	ptr := api.DecodeI32(res[0])

	h := &rtypes.TypeHeader{}
	h.HLoad(w.ctx, ptr)
	resType, ok := rtypes.TypeObjectMapping[h.DataType()]
	if !ok {
		return nil, errors.Errorf("read param failed, type %d not found", resType)
	}

	resType.Load(w.ctx, ptr)
	return resType.Get(), nil
}

func (w *wazeroRuntime) Destroy() {
}

func (w *wazeroRuntime) ResetStore(apis *HostAPIRegistry) (err error) {
	// In Wazero runtime clean-up and re-instantiate can be done together
	if err := w.instance.Close(w.ctx); err != nil {
		return err
	}

	w.moduleConfig = wazero.NewModuleConfig()
	w.instance, err = w.rt.InstantiateModule(w.ctx, w.compiled, w.moduleConfig)

	return
}

func (w *wazeroRuntime) setCtx() {
	w.ctx = rtypes.NewContext(
		// The context here is reserved for future functionalities,
		// such as logging, tracing, and other purposes.
		context.Background(),
		rtypes.NewWAZeroMemory(
			w.instance.Memory(),
			func(size int32) (int32, error) {
				memoryAllocator := w.instance.ExportedFunction("allocate")
				if memoryAllocator == nil {
					return 0, errors.New("function 'allocate' does not exist")
				}

				res, err := memoryAllocator.Call(w.ctx, uint64(size))
				if err != nil {
					return 0, err
				}

				return api.DecodeI32(res[0]), nil
			},
		),
	)
}
