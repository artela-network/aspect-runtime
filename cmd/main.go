package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	rt "runtime"
	"time"

	_ "net/http/pprof"

	"github.com/artela-network/runtime"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "../wasmtime/testdata/runtime_test.wasm"))
	for i := 0; i < 10000; i++ {
		if i%100 == 0 {
			fmt.Println("call times: ", i)
			// time.Sleep(5 * time.Second)
		}
		callRT(raw)
		rt.GC()
	}
	// callRTLoop(raw)
	time.Sleep(10 * time.Minute)
}

func callPool(raw []byte) {
	hostApis := runtime.NewHostAPIRegistry()
	addApis(hostApis)

	pool := runtime.NewRuntimePool(10)
	key, wasmTimeRuntime, err := pool.Runtime(runtime.WASM, raw, hostApis)
	if err != nil {
		fmt.Println("NewAspectRuntime", err)
	}
	pool.Return(key, wasmTimeRuntime)
}

func callRTLoop(raw []byte) {
	hostApis := runtime.NewHostAPIRegistry()

	var (
		arg             string = "abcd"
		wasmTimeRuntime runtime.AspectRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = runtime.NewAspectRuntime(runtime.WASM, raw, hostApis)
	if err != nil {
		fmt.Println("NewAspectRuntime", err)
	}

	for i := 0; i < 10000; i++ {
		_ = arg
		_ = wasmTimeRuntime
		{
			res, err := wasmTimeRuntime.Call("greet", arg)
			if err != nil {
				fmt.Println("call greet", err)
			}

			if "hello-greet-abcd-hello-greet" != res.(string) {
				fmt.Println("return result error: ", res)
			}
		}
	}
}

func callRT(raw []byte) {
	hostApis := runtime.NewHostAPIRegistry()

	var (
		arg             string = "abcd"
		wasmTimeRuntime runtime.AspectRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = runtime.NewAspectRuntime(runtime.WASM, raw, hostApis)
	if err != nil {
		fmt.Println("NewAspectRuntime", err)
	}

	_ = arg
	_ = wasmTimeRuntime
	// {
	// 	res, err := wasmTimeRuntime.Call("greet", arg)
	// 	if err != nil {
	// 		fmt.Println("call greet", err)
	// 	}

	// 	if "hello-greet-abcd-hello-greet" != res.(string) {
	// 		fmt.Println("return result error: ", res)
	// 	}
	// }
}

func addApis(hostApis *runtime.HostAPIRegistry) {
	hostApis.AddApi("index", "test", "hello", func(arg string) string {
		return "hello-" + arg + "-hello"
	})
	hostApis.AddApi("index", "test", "hello2", func(arg1 string, arg2 string, arg3 string) string {
		tmp := arg2 + arg3
		return arg1 + "-" + tmp
	})
	hostApis.AddApi("index", "test", "hello3", func(arg string) {
		if "greet3-hello" != arg {
			fmt.Println("hostapi error: '", arg, "' not equals to 'greet3-hello'")
		}
	})
}
