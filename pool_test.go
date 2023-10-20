package runtime

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test Case: normal case for addApi add and execute
// Same hostapi instance
func TestCallNormalWithPool(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()
	err := addApis(t, hostApis)
	if err != nil {
		return
	}

	pool := NewRuntimePool(10)

	for i := 0; i < 12; i++ {
		key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, err := wasmTimeRuntime.Call("testIncrease")
		require.Equal(t, nil, err)

		// global: let sum = 0;
		// in function: sum = sum + 10;
		// here res relies on the global variable, but it is still expected to 10
		// because the global value should not be exist after last calling.
		require.Equal(t, "10", res.(string))
		pool.Return(key, wasmTimeRuntime)
	}
}

// Different hostapi instance
func TestCallNormalWithPool2(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	pool := NewRuntimePool(10)

	for i := 0; i < 12; i++ {
		hostApis := NewHostAPIRegistry()
		// nolint
		err := addApis(t, hostApis)
		if err != nil {
			return
		}

		key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, err := wasmTimeRuntime.Call("testIncrease")
		require.Equal(t, nil, err)

		require.Equal(t, "10", res.(string))
		pool.Return(key, wasmTimeRuntime)

		require.Equal(t, 1, pool.Len())
	}
}

func TestPoolPerformance(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()
	err := addApis(t, hostApis)
	if err != nil {
		return
	}

	// call without pool
	t1 := time.Now()
	for i := 0; i < 100; i++ {
		wasmTimeRuntime, err := NewAspectRuntime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		_ = wasmTimeRuntime
	}
	t2 := time.Now()
	cost1 := t2.Sub(t1).Microseconds()
	fmt.Printf("total cost without pool: %dμs\n", cost1)

	// call with pool
	pool := NewRuntimePool(10)
	t3 := time.Now()
	for i := 0; i < 100; i++ {
		key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		pool.Return(key, wasmTimeRuntime)
	}
	t4 := time.Now()
	cost2 := t4.Sub(t3).Microseconds()
	fmt.Printf("total cost with pool: %dμs\n", cost2)
	fmt.Println("cost with pool / cost without pool: ", float32(cost2)/float32(cost1)) // it is 0.2606396 in one test
}
