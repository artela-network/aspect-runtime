package wasmtime

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/artela-network/runtime"
	"github.com/stretchr/testify/require"
)

// Helper: init hostAPI collection(@see type script impl :: declare)
func addApis(hostApis *runtime.HostAPICollection) {
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
	raw, _ := os.ReadFile(path.Join(cwd, "./testdata/runtime_test.wasm"))

	hostApis := runtime.NewHostAPICollection()

	var (
		arg             string = ""
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
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
	raw, _ := os.ReadFile(path.Join(cwd, "./testdata/runtime_test.wasm"))

	hostApis := runtime.NewHostAPICollection()

	var (
		arg             string = "abcd"
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)
	addApis(hostApis)

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
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
	raw, _ := os.ReadFile(path.Join(cwd, "./testdata/runtime_test.wasm"))

	hostApis := runtime.NewHostAPICollection()

	var (
		arg1            string = "bonjour"
		arg2            string = "2"
		arg3            string = "5"
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)

	addApis(hostApis)

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
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
	raw, _ := os.ReadFile(path.Join(cwd, "./testdata/runtime_test.wasm"))

	hostApis := runtime.NewHostAPICollection()

	addApis(hostApis)

	var (
		arg             []byte = []byte{0x1, 0x2, 0x3, 0x4}
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
	require.Equal(t, nil, err)
	res, err := wasmTimeRuntime.Call("testBytes", arg)
	require.Equal(t, nil, err)

	require.Equal(t, true, reflect.DeepEqual([]byte{0x2, 0x3, 0x4, 0x5}, res.([]byte)))
}
