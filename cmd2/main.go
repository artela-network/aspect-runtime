package main

import (
	"fmt"
	"github.com/artela-network/runtime"
	"log"
	"os"
	"path"
	"time"

	// "net/http"
	// _ "net/http/pprof"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

const (
	TIMES = 100000
)

func main() {
	go func() {
		app := fiber.New()

		app.Use(pprof.New())

		app.Get("/", handleRequest)
		log.Fatal(app.Listen(":8080"))
	}()
	//ConstantJoinPointCount()
	//ConstantComputationalComplexity()
	TestPooledCall()
	TestCall()
}

func handleRequest(c *fiber.Ctx) error {
	log.Printf("handling request from: %s", c.Context().RemoteAddr())
	if _, err := c.WriteString(c.Context().RemoteAddr().String()); err != nil {
		log.Printf("could not write IP: %s", err)
	}
	return nil
}

func TestCall() {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./release.wasm"))

	hostApis := runtime.NewHostAPIRegistry()

	rt, _ := runtime.NewAspectRuntime(runtime.WASM, raw, hostApis)
	wast := rt.(*runtime.WasmTimeRuntime)

	wasmPrepare := func() {
		wast.Instance, _ = wast.Linker.Instantiate(wast.Store, wast.Module)

		// set context
		wast.SetCtx()

		// set memory instance to apis, for host function ctx.
		wast.Apis.SetMemory(wast.Ctx.Memory())
	}

	start := time.Now()
	for i := 0; i < 1_0000; i++ {
		wasmPrepare()
	}
	end := time.Now()

	fmt.Printf("%d ns/op\n", end.Sub(start).Nanoseconds()/1_0000)
}

func TestPooledCall() {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./release.wasm"))

	hostApis := runtime.NewHostAPIRegistry()

	pool := runtime.NewRuntimePool(10)
	key, rt, err := pool.Runtime(runtime.WASM, raw, hostApis)
	if err != nil {
		panic(err)
	}
	pool.Return(key, rt)

	start := time.Now()
	for i := 0; i < 1_0000; i++ {
		key, rt, err := pool.Runtime(runtime.WASM, raw, hostApis)
		if err != nil {
			fmt.Println(i)
			panic(err)
		}

		pool.Return(key, rt)
	}
	end := time.Now()

	fmt.Printf("%d ns/op\n", end.Sub(start).Nanoseconds()/1_0000)
}

func ConstantJoinPointCount() {
	jp := int64(10)
	BenchWasm(10, 0, jp)
	BenchWasm(10, 40, jp)
	BenchWasm(10, 80, jp)
	BenchWasm(10, 120, jp)
	BenchWasm(10, 160, jp)
	BenchWasm(10, 200, jp)
	BenchWasm(10, 240, jp)
	BenchWasm(10, 280, jp)
	BenchWasm(10, 320, jp)
	BenchWasm(10, 360, jp)
	BenchWasm(10, 400, jp)

	BenchEVM(10, 0, jp)
	BenchEVM(10, 40, jp)
	BenchEVM(10, 80, jp)
	BenchEVM(10, 120, jp)
	BenchEVM(10, 160, jp)
	BenchEVM(10, 200, jp)
	BenchEVM(10, 240, jp)
	BenchEVM(10, 280, jp)
	BenchEVM(10, 320, jp)
	BenchEVM(10, 360, jp)
	BenchEVM(10, 400, jp)
}

func ConstantComputationalComplexity() {
	// BenchJoinPoint(true)
	// BenchCallEVM()
	m := int64(10)
	n := int64(160)

	BenchWasm(m, n, 1)
	BenchWasm(m, n, 4)
	BenchWasm(m, n, 8)
	BenchWasm(m, n, 12)
	BenchWasm(m, n, 16)
	BenchWasm(m, n, 20)
	BenchWasm(m, n, 24)
	BenchWasm(m, n, 28)
	BenchWasm(m, n, 32)
	BenchWasm(m, n, 36)
	BenchWasm(m, n, 40)
	// BenchWasm(m, n, 96)

	BenchEVM(m, n, 1)
	BenchEVM(m, n, 4)
	BenchEVM(m, n, 8)
	BenchEVM(m, n, 12)
	BenchEVM(m, n, 16)
	BenchEVM(m, n, 20)
	BenchEVM(m, n, 24)
	BenchEVM(m, n, 28)
	BenchEVM(m, n, 32)
	BenchEVM(m, n, 36)
	BenchEVM(m, n, 40)
	// BenchEVM(m, n, 96)

	// BenchCostOfSwithContext()
}
