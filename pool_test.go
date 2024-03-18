package runtime

import (
	"context"
	"fmt"
	"github.com/artela-network/aspect-runtime/types"
	"github.com/artela-network/aspect-runtime/wasmtime"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test Case: normal case for addApi add and execute
// Same hostapi instance
func TestCallNormalWithPool(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
	err := addApis(t, hostApis)
	if err != nil {
		return
	}

	pool := NewRuntimePool(context.Background(), log.New(), 10)

	for i := 0; i < 12; i++ {
		key, wasmTimeRuntime, err := pool.Runtime(context.Background(), WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, _, err := wasmTimeRuntime.Call("testIncrease", types.MaxGas)
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

	pool := NewRuntimePool(context.Background(), log.New(), 10)

	for i := 0; i < 12; i++ {
		hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
		// nolint
		err := addApis(t, hostApis)
		if err != nil {
			return
		}

		key, wasmTimeRuntime, err := pool.Runtime(context.Background(), WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, _, err := wasmTimeRuntime.Call("testIncrease", types.MaxGas)
		require.Equal(t, nil, err)

		require.Equal(t, "10", res.(string))
		pool.Return(key, wasmTimeRuntime)

		require.Equal(t, 1, pool.Len())
	}
}

func TestPoolPerformance(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))
	// call without pool
	t1 := time.Now()
	for i := 0; i < 100; i++ {
		hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
		err := addApis(t, hostApis)
		if err != nil {
			return
		}

		wasmTimeRuntime, err := NewAspectRuntime(context.Background(), log.New(), WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, _, err := wasmTimeRuntime.Call("testIncrease", types.MaxGas)
		require.Equal(t, nil, err)

		require.Equal(t, "10", res.(string))
		wasmTimeRuntime.Destroy()
		wasmTimeRuntime = nil
	}
	cost1 := time.Now().Sub(t1).Microseconds()
	fmt.Printf("total cost without pool: %dμs\n", cost1)

	// call with pool
	pool := NewRuntimePool(context.Background(), log.New(), 10)
	t2 := time.Now()
	for i := 0; i < 100; i++ {
		hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
		err := addApis(t, hostApis)
		if err != nil {
			return
		}

		key, wasmTimeRuntime, err := pool.Runtime(context.Background(), WASM, raw, hostApis)
		require.Equal(t, nil, err)

		res, _, err := wasmTimeRuntime.Call("testIncrease", types.MaxGas)
		require.Equal(t, nil, err)

		require.Equal(t, "10", res.(string))
		pool.Return(key, wasmTimeRuntime)

		require.Equal(t, 1, pool.Len())
	}
	cost2 := time.Now().Sub(t2).Microseconds()
	fmt.Printf("total cost with pool: %dμs\n", cost2)
	fmt.Printf("cost with pool / cost without pool: %.2f%%\n", float32(cost2)/float32(cost1)*100) // it is 0.2606396 in one test
}

func TestPoolParallelPerformance(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	times := 500
	poolsize := 30
	totalCost1 := 0
	totalCost2 := 0

	for n := 0; n < 1; n++ {
		// call without pool
		t1 := time.Now()
		var wg1 sync.WaitGroup
		for i := 1; i < times; i++ {
			wg1.Add(1)

			// if i%poolsize == 0 {
			// 	time.Sleep(time.Duration(poolsize) * time.Millisecond)
			// }
			go func() {
				hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
				err := addApis(t, hostApis)
				if err != nil {
					return
				}

				wasmTimeRuntime, err := NewAspectRuntime(context.Background(), log.New(), WASM, raw, hostApis)
				require.Equal(t, nil, err)
				res, _, err := wasmTimeRuntime.Call("greet", types.MaxGas, "abc")
				require.Equal(t, nil, err)

				require.Equal(t, "hello-greet-abc-hello-greet", res.(string))
				wasmTimeRuntime.Destroy()
				wasmTimeRuntime = nil
				wg1.Done()
			}()

		}
		wg1.Wait()
		// fmt.Println("total used: ", poolUsed)
		cost1 := time.Now().Sub(t1).Milliseconds()
		fmt.Printf("cost without pool: %dms\n", cost1)
		time.Sleep(1 * time.Second)
		totalCost1 += int(cost1)

		// call with pool
		pool := NewRuntimePool(context.Background(), log.New(), poolsize)
		t2 := time.Now()
		var wg2 sync.WaitGroup
		for i := 1; i < times; i++ {
			wg2.Add(1)

			// if i%poolsize == 0 {
			// 	time.Sleep(time.Duration(poolsize) * time.Millisecond)
			// }
			go func() {
				hostApis := types.NewHostAPIRegistry(wasmtime.Wrap)
				err := addApis(t, hostApis)
				if err != nil {
					return
				}

				key, wasmTimeRuntime, err := pool.Runtime(context.Background(), WASM, raw, hostApis)
				require.Equal(t, nil, err)
				res, _, err := wasmTimeRuntime.Call("greet", types.MaxGas, "abc")
				require.Equal(t, nil, err)

				require.Equal(t, "hello-greet-abc-hello-greet", res.(string))
				pool.Return(key, wasmTimeRuntime)

				// require.Equal(t, 1, pool.Len())
				wg2.Done()
			}()
		}
		wg2.Wait()
		// fmt.Println("total used: ", poolUsed)
		cost2 := time.Now().Sub(t2).Milliseconds()
		fmt.Printf("total cost with pool: %dms\n", cost2)
		totalCost2 += int(cost2)
		runtime.GC()
	}
	fmt.Printf("cost with pool / cost without pool: %.2f%%\n", float32(totalCost2)/float32(totalCost1)*100) // it is 0.2606396 in one test
}
