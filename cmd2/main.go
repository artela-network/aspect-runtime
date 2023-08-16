package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	TIMES = 10000
)

func main() {
	args := os.Args
	fmt.Println(args)
	if len(args) != 5 {
		fmt.Println("command error, format: cmd2 [evm/wasm] [m] [n] [join point count]")
	}
	args = args[1:]

	m, err := strconv.Atoi(args[1])
	if err != nil {
		panic(err)
	}

	n, err := strconv.Atoi(args[2])
	if err != nil {
		panic(err)
	}

	jpc, err := strconv.Atoi(args[3])
	if err != nil {
		panic(err)
	}

	if args[0] == "evm" {
		BenchEVM(int64(m), int64(n), int64(jpc))
	} else if args[0] == "wasm" {
		BenchWasm(int64(m), int64(n), int64(jpc))
	}
}

func Test() {
	BenchWasm(10, 0, 1)
	BenchEVM(10, 0, 1)
}

func Test_3() {
	// execution time = 1000 micro seconds
	BenchWasm(10, 1, 36)
}

func Test_2() {
	jp := int64(10)
	BenchWasm(10, 0, jp)
	GC()

	BenchWasm(10, 40, jp)
	GC()

	BenchWasm(10, 80, jp)
	GC()

	BenchWasm(10, 120, jp)
	GC()

	BenchWasm(10, 160, jp)
	GC()

	BenchWasm(10, 200, jp)
	GC()

	BenchWasm(10, 240, jp)
	GC()

	BenchWasm(10, 280, jp)
	GC()

	BenchWasm(10, 320, jp)
	GC()

	BenchWasm(10, 360, jp)
	GC()

	// EVM

	BenchEVM(10, 0, jp)
	GC()

	BenchEVM(10, 40, jp)
	GC()

	BenchEVM(10, 80, jp)
	GC()

	BenchEVM(10, 120, jp)
	GC()

	BenchEVM(10, 160, jp)
	GC()

	BenchEVM(10, 200, jp)
	GC()

	BenchEVM(10, 240, jp)
	GC()

	BenchEVM(10, 280, jp)
	GC()

	BenchEVM(10, 320, jp)
	GC()

	BenchEVM(10, 360, jp)
	GC()
}

func Test_1() {
	// BenchJoinPoint(true)
	// BenchCallEVM()
	m := int64(10)
	n := int64(160)

	BenchWasm(m, n, 1)
	GC()

	BenchWasm(m, n, 4)
	GC()

	BenchWasm(m, n, 8)
	GC()

	BenchWasm(m, n, 12)
	GC()

	BenchWasm(m, n, 16)
	GC()

	BenchWasm(m, n, 20)
	GC()

	BenchWasm(m, n, 24)
	GC()

	BenchWasm(m, n, 28)
	GC()

	BenchWasm(m, n, 32)
	GC()

	BenchWasm(m, n, 36)
	GC()

	// BenchWasm(m, n, 40)
	// BenchWasm(m, n, 96)

	time.Sleep(5 * time.Second)

	BenchEVM(m, n, 1)
	GC()

	BenchEVM(m, n, 4)
	GC()

	BenchEVM(m, n, 8)
	GC()

	BenchEVM(m, n, 12)
	GC()

	BenchEVM(m, n, 16)
	GC()

	BenchEVM(m, n, 20)
	GC()

	BenchEVM(m, n, 24)
	GC()

	BenchEVM(m, n, 28)
	GC()

	BenchEVM(m, n, 32)
	GC()

	BenchEVM(m, n, 36)
	GC()

	// BenchEVM(m, n, 40)
	// BenchEVM(m, n, 96)

	// BenchCostOfSwithContext()
}

func GC() {
	runtime.GC()
	time.Sleep(2 * time.Second)
}
