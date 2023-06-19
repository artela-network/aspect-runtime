package runtime

import (
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test Case: normal case for addApi add and execute
func TestCallNormalWithPool(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()
	addApis(t, hostApis)

	pool := NewRuntimePool(10)

	for i := 0; i < 12; i++ {
		key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, err := wasmTimeRuntime.Call("testIncrease")
		require.Equal(t, nil, err)

		require.Equal(t, strconv.Itoa((i+1)*10), res.(string))
		pool.PutBack(key, wasmTimeRuntime)
	}
}

func TestCallNormalWithPool2(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	pool := NewRuntimePool(10)

	for i := 0; i < 12; i++ {
		hostApis := NewHostAPIRegistry()
		addApis(t, hostApis)
		key, wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis)
		require.Equal(t, nil, err)
		res, err := wasmTimeRuntime.Call("testIncrease")
		require.Equal(t, nil, err)

		require.Equal(t, strconv.Itoa(10), res.(string))
		pool.PutBack(key, wasmTimeRuntime)

		expectLen := i + 1
		if i >= 10 {
			expectLen = 10
		}
		if i <= 10 {
			require.Equal(t, expectLen, pool.Len())
		}
	}
}
