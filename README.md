# Aspect Runtime
Aspect runtime is a runtime platform for executing aspects. It is a wrapper around the bytecodealliance/wasmtime engine and supports any language that is compatible with WebAssembly (wasm). Building upon the foundation of fixed-length variables supported by wasm, Aspect runtime also enables the use of custom data structures and facilitates communication with the host environment.

# System Overview
- `Register`: defining host functions registered with the runtime.
- `Runtime`: defining interfaces of aspect runtime.
- `wasmtime`: defining wasm implement of interface.
- `wrapper`: wrapping registered host functions and handle complex params.
- `types`: defining all types supported as params.

# Usage
1. Register the host functions
    ```
    hostFns := runtime.NewHostAPIRegistry()
    hostFns.AddApi("index", "test", "hello", func(arg string) string {
        return "greet to " + arg;
    })
    ```
2. Create runtime and pass registry of host function.
    ```
    raw := []byte() // raw bytecode
    
    wasmTimeRuntime, err = runtime.NewAspectRuntime(runtime.WASM, raw, hostFns)
    res, err := wasmTimeRuntime.Call("greet", arg)
    if err != nil {
        // handle the error
    }
    output = res.(string) // string is expected
    // handle the output
    wasmTimeRuntime.Destroy() // to destroy the rt, in case of go runtime gc failed
    ...
    ```
3. Use the pool.
    <br/>To establish a runtime pool with a specific capacity, the pool is designed to cache and provide runtime instances based on their type, bytecode, and host functions.
    The `preRun` function is invoked before returning the instance and serves the purpose of clearing any variables or data from the previous run.
    ```
    capacity := 10
	pool := NewRuntimePool(capacity)

    preRun := "resetFn"
    key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostFns, forceClearCache, "resetFn")

    ```
    After using the runtime instance from the pool, it is important to put it back into the pool. Failing to do so would result in the instance being unavailable for subsequent calls that require the same instance.
    ```
    pool.PutBack(key, wasmTimeRuntime)
    ```