package wasmtime

import "github.com/artela-network/runtime"

func init() {
	// register runner
	runtime.RegisterAspectRunner(runtime.WASMTime, NewWASMTimeRuntime)
}
