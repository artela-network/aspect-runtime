package runtime

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

// Helper: init hostAPI collection(@see type script impl :: declare)
func addApis(t *testing.T, hostApis *HostAPIRegistry) error {
	err := hostApis.AddAPI("index", "test", "hello", func(arg string) string {
		return "hello-" + arg + "-hello"
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("index", "test", "hello2", func(arg1 string, arg2 string, arg3 string) string {
		tmp := arg2 + arg3
		return arg1 + "-" + tmp
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("index", "test", "hello3", func(arg string) {
		require.Equal(t, "greet3-hello", arg)
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("index", "test", "hello4", func(arg string) (string, error) {
		return "", errors.New("error")
	})
	if err != nil {
		return err
	}
	return nil
}

func TestAddApi(t *testing.T) {
	hostApis := NewHostAPIRegistry()
	err := hostApis.AddAPI("index", "test", "hello4", func(arg string) (string, error) {
		return "", errors.New("error")
	})
	if err != nil {
		return
	}
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
	err = addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)
		require.Equal(t, "hello-greet--hello-greet", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
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
	err2 := addApis(t, hostApis)
	if err2 != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)

		require.Equal(t, "hello-greet-abcd-hello-greet", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
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

	err = addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet2", arg1, arg2, arg3)
		require.Equal(t, nil, err)
		require.Equal(t, "bonjour-25-over", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: normal case for []byte as arg
func TestBytesNormal(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	testErr := addApis(t, hostApis)
	if testErr != nil {
		return
	}

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
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

func TestCallHostApiNoReturn(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	errapi := addApis(t, hostApis)
	if errapi != nil {
		return
	}

	var (
		arg             string = "hello"
		wasmTimeRuntime AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, err := wasmTimeRuntime.Call("greet3", arg)
	require.Equal(t, nil, err)

	require.Equal(t, "greet3", res.(string))
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: nil case for []byte as arg
func TestBytesNil(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := NewHostAPIRegistry()

	addErr := addApis(t, hostApis)
	if addErr != nil {
		return
	}

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
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
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
	addEr := addApis(t, hostApis)
	if addEr != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)
		output := res.(string)

		require.Equal(t, "hello-greet-"+arg+"-hello-greet", output)
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}
