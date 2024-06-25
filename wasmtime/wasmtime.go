package wasmtime

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	wasmtime "github.com/bytecodealliance/wasmtime-go/v20"
	"github.com/pkg/errors"

	"github.com/artela-network/aspect-runtime/types"
)

const (
	MaxMemorySize = 32 * 1024 * 1024
)

// wasmTimeRuntime is a wrapper for WASMTime runtime
type wasmTimeRuntime struct {
	sync.Mutex

	engine *wasmtime.Engine
	module *wasmtime.Module
	linker *wasmtime.Linker

	ctx *Context

	apis *types.HostAPIRegistry

	logger types.Logger
}

func NewWASMTimeRuntime(ctx context.Context, logger types.Logger, code []byte, apis *types.HostAPIRegistry) (out types.AspectRuntime, err error) {
	watvm := &wasmTimeRuntime{
		engine: wasmtime.NewEngineWithConfig(defaultWASMTimeConfig()),
		logger: logger,
	}

	// init wasm module
	watvm.module, err = wasmtime.NewModule(watvm.engine, code)
	if err != nil {
		logger.Error("failed to create wasm module", "err", err, "size", len(code))
		return nil, errors.Wrap(err, "unable create wasm module")
	}

	// init runtime context
	watvm.ctx = NewContext(ctx, logger)
	watvm.ctx.Store = wasmtime.NewStore(watvm.engine)
	// limit memory size to 32MB for now
	watvm.ctx.Store.Limiter(MaxMemorySize, -1, -1, -1, 100)

	// create host api linker
	watvm.linker = wasmtime.NewLinker(watvm.engine)

	// link all host apis
	watvm.apis = apis
	if err := watvm.linkToHostFns(); err != nil {
		logger.Error("failed to link host functions", "err", err)
		return nil, err
	}

	// add abort function
	if err := watvm.linkAbort(); err != nil {
		logger.Error("failed to link abort function", "err", err)
		return nil, err
	}

	// instantiate module and store
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
			logger.Error("failed to create wasm instance", "err", r, "stack", debug.Stack())
		}
	}()

	watvm.ctx.Instance, err = watvm.linker.Instantiate(watvm.ctx.Store, watvm.module)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instantiate wasm module")
	}

	return watvm, err
}

func (w *wasmTimeRuntime) Context() context.Context {
	return w.ctx
}

func (w *wasmTimeRuntime) Logger() types.Logger {
	return w.logger
}

// Call wasm
func (w *wasmTimeRuntime) Call(method string, gas int64, args ...interface{}) (res interface{}, leftover int64, err error) {
	startTime := time.Now()

	defer func() {
		w.logger.Debug("aspect execution done",
			"duration", time.Since(startTime).String(),
			"remainingGas", leftover,
			"err", err)
	}()

	w.Lock()
	defer w.Unlock()

	w.logger.Info("calling aspect", "method", method, "gas", gas)
	w.logger.Debug("initializing aspect")
	if err := w.init(gas); err != nil {
		return nil, 0, errors.Errorf("aspect init failed, %v", err)
	}

	w.logger.Debug("executing aspect")
	val, callErr := w.call(method, args...)

	leftover, gasErr := w.ctx.RemainingEVMGas()
	if gasErr != nil {
		w.logger.Error("failed to get remaining gas", "err", gasErr)
		return nil, 0, gasErr
	}

	w.logger.Info("aspect executed", "method", method, "gas", gas, "leftover", leftover, "result", val, "err", callErr)

	if callErr != nil {
		return nil, leftover, callErr
	}

	ptr, ok := val.(int32)
	if !ok {
		return nil, leftover, errors.Errorf("read output failed, value: %v", val)
	}

	if ptr == 0 {
		// void functions this will be 0
		return nil, leftover, nil
	}

	header, err := w.ctx.ReadMemory(ptr, types.HeaderLen)
	if err != nil {
		w.logger.Error("failed to read return value header", "err", err)
		return nil, leftover, err
	}

	h := &types.TypeHeader{}
	dataType, dataLen, err := h.Unmarshal(header)
	if err != nil {
		w.logger.Error("failed to unmarshal return value header", "err", err)
		return nil, leftover, errors.Errorf("read output failed, %v", err)
	}

	resType, err := types.TypeObjectMapping(dataType)
	if err != nil {
		w.logger.Error("unsupported return value data type", "err", err, "dataType", dataType)
		return nil, leftover, errors.Errorf("unsupported result type, %v", err)
	}

	retData, err := w.ctx.ReadMemory(ptr, types.HeaderLen+dataLen)
	if err != nil {
		w.logger.Error("failed to read return value", "err", err)
		return nil, leftover, errors.Errorf("read output failed, %v", err)
	}

	res, err = resType.Unmarshal(retData)
	if err != nil {
		w.logger.Error("failed to unmarshal return value", "err", err)
		return nil, leftover, errors.Errorf("read output failed, %v", err)
	}

	return res, leftover, nil
}

func (w *wasmTimeRuntime) call(method string, args ...interface{}) (interface{}, error) {
	run := w.ctx.Instance.GetFunc(w.ctx.Store, method)
	if run == nil {
		return nil, errors.Errorf("method %s does not exist", method)
	}

	w.apis.SetContext(w.ctx)

	ptrs := make([]interface{}, len(args))
	for i, arg := range args {
		typeIndex := types.AssertType(arg)
		rtType, err := types.TypeObjectMapping(typeIndex)
		if err != nil {
			return nil, err
		}

		data := rtType.Marshal(arg)
		ptr, err := w.ctx.AllocMemory(int32(len(data)))
		if err != nil {
			return nil, err
		}

		if err := w.ctx.WriteMemory(ptr, data); err != nil {
			return nil, err
		}

		ptrs[i] = ptr
	}

	val, err := run.Call(w.ctx.Store, ptrs...)
	if err != nil {
		if err.Error() == types.OutOfGasError.Error() {
			return nil, types.OutOfGasError
		}

		return nil, errors.Wrapf(err, "method %s execution fail, err: %s", method, err.Error())
	}

	return val, nil
}

func (w *wasmTimeRuntime) init(gas int64) error {
	w.logger.Debug("filling up gas", "gas", gas)
	if err := w.ctx.AddEVMGas(gas); err != nil {
		w.logger.Error("failed to add gas", "err", err)
		return err
	}

	w.logger.Debug("initializing aspect")
	if _, err := w.call("__aspect_start__"); err != nil {
		w.logger.Error("failed to initialize aspect", "err", err)
		return err
	}

	w.logger.Debug("aspect initialized")
	return nil
}

// ResetStore reset the whole memory of wasm
func (w *wasmTimeRuntime) ResetStore(ctx context.Context, apis *types.HostAPIRegistry) (err error) {
	w.Lock()
	defer w.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintln(r))
			w.logger.Error("reset wasm store panic", "err", r, "stack", debug.Stack())
		}
	}()

	w.logger.Debug("resetting wasm store")

	w.ctx = NewContext(ctx, w.logger)
	w.ctx.Store = wasmtime.NewStore(w.engine)
	w.ctx.Store.Limiter(MaxMemorySize, -1, -1, -1, 100)

	w.linker = wasmtime.NewLinker(w.engine)
	w.apis = apis

	// reset link all host apis with new store
	if err := w.linkToHostFns(); err != nil {
		w.logger.Error("failed to link host functions", "err", err)
		return err
	}

	// add abort function
	if err := w.linkAbort(); err != nil {
		w.logger.Error("failed to link abort function", "err", err)
		return err
	}

	w.ctx.Instance, err = w.linker.Instantiate(w.ctx.Store, w.module)
	if err != nil {
		w.logger.Error("failed to instantiate wasm module", "err", err)
		return errors.Wrap(err, "unable to instantiate wasm module")
	}

	w.logger.Debug("wasm store reset")

	return nil
}

func (w *wasmTimeRuntime) Destroy() {
	w.Lock()
	defer w.Unlock()

	w.logger.Debug("destroying wasm runtime")

	w.apis = nil

	// Deallocate resources associated with the instance, linker, and store.
	// These components will be reconstructed before the next invocation.
	w.linker = nil
	w.ctx.Reset()
	w.ctx = nil
}

func (w *wasmTimeRuntime) linkToHostFns() error {
	for module, namespaces := range w.apis.WrapperFuncs() {
		for ns, methods := range namespaces {
			for method, function := range methods {
				// create a function wrapper for our "hostapi" on the go side
				item := wasmtime.WrapFunc(w.ctx.Store, function)

				// create linker with host function injected
				if err := w.linker.Define(
					w.ctx.Store,
					buildModuleName(module),
					buildModuleMethod(ns, method),
					item,
				); err != nil {
					w.logger.Error("failed to link host api", "module", module, "namespace", ns, "method", method, "err", err)
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
	abort := wasmtime.WrapFunc(w.ctx.Store, func(a, b, c, d int32) {
		w.logger.Debug("abort called", "a", a, "b", b, "c", c, "d", d)
	})
	if err := w.linker.Define(w.ctx.Store, "env", "abort", abort); err != nil {
		return errors.Wrapf(err, "unable to link to abort")
	}
	return nil
}

// defaultWASMTimeConfig provides a default wasmtime config for the runner.
// TODO: currently this is just a very early version, should investigate deeper for each config option.
func defaultWASMTimeConfig() *wasmtime.Config {
	config := wasmtime.NewConfig()
	// we don't quite need this, discuss later
	config.SetWasmSIMD(false)
	// cannot enable RelaxedSIMD when SIMD is disabled
	config.SetWasmRelaxedSIMD(false)
	// affect execution certainty, disable
	config.SetWasmThreads(false)
	// multi-value return is useful, should be enabled
	config.SetWasmMultiValue(true)
	// need to run benchmarks on this and adjust later
	config.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	// disable multi-memory by default
	config.SetWasmMultiMemory(false)
	// need to disable bulk memory, parity wasm instrument doesn't support this
	config.SetWasmBulkMemory(false)
	// reference type must be disabled, this relies on bulk memory
	config.SetWasmReferenceTypes(false)

	// enables selecting the "static" option for all linear memories
	config.SetStaticMemoryForced(true)
	// configures the size of linear memory to reserve for each memory in the
	// pooling allocator.
	// lock to 10MB here.
	config.SetStaticMemoryMaximumSize(10 * (1 << 20))
	// configures the size, in bytes, of the guard region used at the end of a
	// static memory's address space reservation.
	// default to 2GB on 64-bit platforms, 64K on 32-bit platforms.
	// lock to 64K here.
	config.SetStaticMemoryGuardSize(0x1_0000)
	// configures the size, in bytes, of the extra virtual memory space
	// reserved after a "dynamic" memory for growing into.
	// for 64-bit platforms this defaults to 2GB, and for 32-bit platforms this
	// defaults to 1MB.
	// lock to 1MB here.
	config.SetDynamicMemoryReservedForGrowth(1 << 20)
	// configures the size, in bytes, of the guard region used at the end of a
	// dynamic memory's address space reservation.
	// This value defaults to 64KB.
	// lock to 64K here.
	config.SetDynamicMemoryGuardSize(0x1_0000)

	return config
}

func buildModuleMethod(ns types.NameSpace, method types.MethodName) string {
	return fmt.Sprintf("%s.%s", ns, method)
}

func buildModuleName(module types.Module) string {
	return string(module)
}
