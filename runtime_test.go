package runtime

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// Helper: init hostAPI collection(@see type script impl :: declare)
func addApis(hostApis *HostAPIRegistry) {
	hostApis.AddApi("index", "test", "hello", func(arg string) string {
		return "hello-" + arg + "-hello"
	})
	hostApis.AddApi("index", "test", "hello2", func(arg1 string, arg2 string, arg3 string) string {
		tmp := arg2 + arg3
		return arg1 + "-" + tmp
	})
}

// Test Case: empty string arg for addApi and execute
func TestCallEmptyStr(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	var (
		arg             string = ""
		wasmTimeRuntime AspectRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)
		require.Equal(t, "hello-greet--hello-greet", res.(string))
	}
}

// Test Case: normal case for addApi add and execute
func TestCallNormal(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	var (
		arg             string = "abcd"
		wasmTimeRuntime AspectRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)

		require.Equal(t, "hello-greet-abcd-hello-greet", res.(string))
	}
}

// Test Case: for multi args of addApi func and execute
func TestCallMultiArgs(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	var (
		arg1            string = "bonjour"
		arg2            string = "2"
		arg3            string = "5"
		wasmTimeRuntime AspectRuntime
		err             error
	)

	addApis(hostApis)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet2", arg1, arg2, arg3)
		require.Equal(t, nil, err)
		require.Equal(t, "bonjour-25-over", res.(string))
	}
}

// Test Case: normal case for []byte as arg
func TestBytesNormal(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	addApis(hostApis)

	var (
		arg             []byte = []byte{0x1, 0x2, 0x3, 0x4}
		wasmTimeRuntime AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, err := wasmTimeRuntime.Call("testBytes", arg)
	require.Equal(t, nil, err)

	require.Equal(t, true, reflect.DeepEqual([]byte{0x2, 0x3, 0x4, 0x5}, res.([]byte)))
}

// Test Case: nil case for []byte as arg
func TestBytesNil(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	addApis(hostApis)

	var (
		arg             []byte = nil
		wasmTimeRuntime AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, err := wasmTimeRuntime.Call("testBytes", arg)
	require.Equal(t, nil, err)

	require.Equal(t, true, reflect.DeepEqual([]byte{}, res.([]byte)))
}

// Test Case: long string as input/output params
func TestLongString(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	var (
		arg             string = ""
		wasmTimeRuntime AspectRuntime
		err             error
	)
	for i := 1; i <= 10000; i++ {
		arg += fmt.Sprintf("%-6d", i)
	}
	addApis(hostApis)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)
		output := res.(string)

		require.Equal(t, "hello-greet-"+arg+"-hello-greet", output)
	}
}

// Test Case: normal case for addApi add and execute
func TestCallNormalWithPool(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()
	addApis(hostApis)

	pool := NewRuntimePool(10)

	for i := 0; i < 3; i++ {
		wasmTimeRuntime, err := pool.Runtime(WASM, raw, hostApis, false)
		require.Equal(t, nil, err)

		{
			res, err := wasmTimeRuntime.Call("testIncrease")
			require.Equal(t, nil, err)

			require.Equal(t, strconv.Itoa((i+1)*10), res.(string))
		}
	}
}
