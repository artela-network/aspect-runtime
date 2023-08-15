package main

import "github.com/artela-network/runtime"

const (
	TIMES = 10000
)

func main() {
	//ConstantJoinPointCount()
	ConstantComputationalComplexity2()
	ConstantComputationalComplexity()
}

func ConstantJoinPointCount() {
	jp := int64(10)
	BenchWasm(runtime.WASMTime, 10, 0, jp)
	BenchWasm(runtime.WASMTime, 10, 40, jp)
	BenchWasm(runtime.WASMTime, 10, 80, jp)
	BenchWasm(runtime.WASMTime, 10, 120, jp)
	BenchWasm(runtime.WASMTime, 10, 160, jp)
	BenchWasm(runtime.WASMTime, 10, 200, jp)
	BenchWasm(runtime.WASMTime, 10, 240, jp)
	BenchWasm(runtime.WASMTime, 10, 280, jp)
	BenchWasm(runtime.WASMTime, 10, 320, jp)
	BenchWasm(runtime.WASMTime, 10, 360, jp)
	BenchWasm(runtime.WASMTime, 10, 400, jp)

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

	//BenchWasm(runtime.WASMTime, m, n, 1)
	//BenchWasm(runtime.WASMTime, m, n, 4)
	BenchWasm(runtime.WASMTime, m, n, 8)
	//BenchWasm(runtime.WASMTime, m, n, 12)
	//BenchWasm(runtime.WASMTime, m, n, 16)
	//BenchWasm(runtime.WASMTime, m, n, 20)
	//BenchWasm(runtime.WASMTime, m, n, 24)
	//BenchWasm(runtime.WASMTime, m, n, 28)
	//BenchWasm(runtime.WASMTime, m, n, 32)
	//BenchWasm(runtime.WASMTime, m, n, 36)
	//BenchWasm(runtime.WASMTime, m, n, 40)
	// BenchWasm(m, n, 96)

	//BenchEVM(m, n, 1)
	//BenchEVM(m, n, 4)
	//BenchEVM(m, n, 8)
	//BenchEVM(m, n, 12)
	//BenchEVM(m, n, 16)
	//BenchEVM(m, n, 20)
	//BenchEVM(m, n, 24)
	//BenchEVM(m, n, 28)
	//BenchEVM(m, n, 32)
	//BenchEVM(m, n, 36)
	//BenchEVM(m, n, 40)
	// BenchEVM(m, n, 96)

	// BenchCostOfSwithContext()
}

func ConstantComputationalComplexity2() {
	// BenchJoinPoint(true)
	// BenchCallEVM()
	m := int64(10)
	n := int64(160)

	//BenchWasm(runtime.WAZero, m, n, 1)
	//BenchWasm(runtime.WAZero, m, n, 4)
	BenchWasm(runtime.WAZero, m, n, 8)
	//BenchWasm(runtime.WAZero, m, n, 12)
	//BenchWasm(runtime.WAZero, m, n, 16)
	//BenchWasm(runtime.WAZero, m, n, 20)
	//BenchWasm(runtime.WAZero, m, n, 24)
	//BenchWasm(runtime.WAZero, m, n, 28)
	//BenchWasm(runtime.WAZero, m, n, 32)
	//BenchWasm(runtime.WAZero, m, n, 36)
	//BenchWasm(runtime.WAZero, m, n, 40)
	// BenchWasm(m, n, 96)

	//BenchEVM(m, n, 1)
	//BenchEVM(m, n, 4)
	//BenchEVM(m, n, 8)
	//BenchEVM(m, n, 12)
	//BenchEVM(m, n, 16)
	//BenchEVM(m, n, 20)
	//BenchEVM(m, n, 24)
	//BenchEVM(m, n, 28)
	//BenchEVM(m, n, 32)
	//BenchEVM(m, n, 36)
	//BenchEVM(m, n, 40)
	// BenchEVM(m, n, 96)

	// BenchCostOfSwithContext()
}
