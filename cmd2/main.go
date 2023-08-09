package main

const (
	TIMES = 2000
)

func main() {
	ConstantJoinPointCount()
	ConstantComputationalComplexity()
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
	n := int64(240)

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
	// BenchEVM(m, n, 96)

	// BenchCostOfSwithContext()
}
