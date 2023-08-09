package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/artela-network/runtime"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	// cwd, _ := os.Getwd()
	// raw, _ := os.ReadFile(path.Join(cwd, "../wasmtime/testdata/runtime_test.wasm"))
	raw, _ := os.ReadFile("/Users/likun/go/src/github.com/test/asexample/build/release.wasm")
	// for i := 0; i < 100000; i++ {
	// 	if i%100 == 0 {
	// 		fmt.Println("call times: ", i)
	// 		// time.Sleep(5 * time.Second)
	// 		time.Sleep(1 * time.Millisecond)
	// 	}
	// 	callRT(raw)
	// 	// rt.GC()
	// }
	callPoolLoop(raw)
	// callRTLoop(raw)
	fmt.Println("Done!")
	wg.Wait()
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

func callPoolLoop(raw []byte) {
	hostApis := runtime.NewHostAPIRegistry()
	addApis(hostApis)

	pool := runtime.NewRuntimePool(10)
	for i := 0; i < 1000000; i++ {
		if i%100 == 0 {
			fmt.Println("call times: ", i)
			time.Sleep(1 * time.Second)
			// time.Sleep(1 * time.Millisecond)
		}
		key, wasmTimeRuntime, err := pool.Runtime(runtime.WASM, raw, hostApis)
		if err != nil {
			fmt.Println("NewAspectRuntime", err)
			panic(err)
		}
		pool.Return(key, wasmTimeRuntime)
	}
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
	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		if err != nil {
			fmt.Println("call greet", err)
		}

		if "hello-greet-abcd-hello-greet" != res.(string) {
			fmt.Println("return result error: ", res)
		}
	}
	wasmTimeRuntime.Destroy()
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
