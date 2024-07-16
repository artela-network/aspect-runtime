package runtime

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/artela-network/aspect-runtime/types"
	"github.com/artela-network/aspect-runtime/wasmtime"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

type mockedHostContext struct {
	gas uint64
}

func (m *mockedHostContext) RemainingGas() uint64 {
	return m.gas
}

func (m *mockedHostContext) SetGas(gas uint64) {
	m.gas = gas
}

func (m *mockedHostContext) SetVMContext(_ types.VMContext) {
}

type mockedLogger struct{}

func (m *mockedLogger) Debug(msg string, keyvals ...interface{}) {
	fmt.Println("DEBUG", msg, keyvals)
}

func (m *mockedLogger) Info(msg string, keyvals ...interface{}) {
	fmt.Println("INFO", msg, keyvals)
}

func (m *mockedLogger) Error(msg string, keyvals ...interface{}) {
	fmt.Println("ERROR", msg, keyvals)
}

func (m *mockedLogger) With(keyvals ...interface{}) types.Logger {
	return m
}

// Helper: init hostAPI collection(@see type script impl :: declare)
func addApis(t *testing.T, hostApis *types.HostAPIRegistry) error {
	err := hostApis.AddAPI("runtime_test", "test", "hello", &types.HostFuncWithGasRule{
		Func: func(arg string) (string, error) {
			return "hello-" + arg + "-hello", nil
		},
		GasRule:     types.NewStaticGasRule(1),
		HostContext: &mockedHostContext{},
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("runtime_test", "test", "hello2", &types.HostFuncWithGasRule{
		Func: func(arg1 string, arg2 string, arg3 string) (string, error) {
			tmp := arg2 + arg3
			return arg1 + "-" + tmp, nil
		},
		GasRule:     types.NewStaticGasRule(1),
		HostContext: &mockedHostContext{},
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("runtime_test", "test", "hello3", &types.HostFuncWithGasRule{
		Func: func(arg string) error {
			require.Equal(t, "greet3-hello", arg)
			return nil
		},
		GasRule:     types.NewStaticGasRule(1),
		HostContext: &mockedHostContext{},
	})
	if err != nil {
		return err
	}
	err = hostApis.AddAPI("runtime_test", "test", "hello4", &types.HostFuncWithGasRule{
		Func: func(arg string) (string, error) {
			return "", errors.New("error")
		},
		GasRule:     types.NewStaticGasRule(1),
		HostContext: &mockedHostContext{},
	})
	if err != nil {
		return err
	}
	return nil
}

func TestAddApi(t *testing.T) {
	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)
	err := hostApis.AddAPI("index", "test", "hello4", &types.HostFuncWithGasRule{
		Func: func(arg string) (string, error) {
			return "", errors.New("error")
		},
		GasRule:     types.NewStaticGasRule(1),
		HostContext: &mockedHostContext{},
	})
	if err != nil {
		return
	}
}

// Test Case: empty string arg for addApi and execute
func TestCallEmptyStr(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	var (
		arg             string = ""
		wasmTimeRuntime types.AspectRuntime
		err             error
	)
	err = addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, leftover, err := wasmTimeRuntime.Call("greet", 10000, arg)
		fmt.Println(leftover)
		require.Equal(t, nil, err)
		require.Equal(t, "hello-greet--hello-greet", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

func TestInfiniteLoop(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	err := addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err := NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	_, leftover, err := wasmTimeRuntime.Call("infiniteLoop", math.MaxInt64)
	fmt.Println(leftover)

	require.Equal(t, nil, err)
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

func TestFib(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))
	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	err := addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err := NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	_, leftover, err := wasmTimeRuntime.Call("fib", math.MaxInt64, uint64(math.MaxInt32), uint64(math.MaxInt32))
	fmt.Println(leftover)

	require.Equal(t, nil, err)
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: normal case for addApi add and execute
func TestCallNormal(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	var (
		arg             string = "abcd"
		wasmTimeRuntime types.AspectRuntime
		err             error
	)
	err2 := addApis(t, hostApis)
	if err2 != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, leftover, err := wasmTimeRuntime.Call("greet", math.MaxInt64, arg)
		fmt.Println(leftover)

		require.Equal(t, nil, err)

		require.Equal(t, "hello-greet-abcd-hello-greet", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: for multi args of addApi func and execute
func TestCallMultiArgs(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	var (
		arg1            string = "bonjour"
		arg2            string = "2"
		arg3            string = "5"
		wasmTimeRuntime types.AspectRuntime
		err             error
	)

	err = addApis(t, hostApis)
	if err != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, leftover, err := wasmTimeRuntime.Call("greet2", types.MaxGas, arg1, arg2, arg3)
		fmt.Println(leftover)

		require.Equal(t, nil, err)
		require.Equal(t, "bonjour-25-over", res.(string))
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: normal case for []byte as arg
func TestBytesNormal(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	testErr := addApis(t, hostApis)
	if testErr != nil {
		return
	}

	var (
		arg             []byte = []byte{0x1, 0x2, 0x3, 0x4}
		wasmTimeRuntime types.AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, leftover, err := wasmTimeRuntime.Call("testBytes", math.MaxInt64, arg)
	fmt.Println(leftover)

	require.Equal(t, nil, err)

	require.Equal(t, true, reflect.DeepEqual([]byte{0x2, 0x3, 0x4, 0x5}, res.([]byte)))
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

func TestCallHostApiNoReturn(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	errapi := addApis(t, hostApis)
	if errapi != nil {
		return
	}

	var (
		arg             string = "hello"
		wasmTimeRuntime types.AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, leftover, err := wasmTimeRuntime.Call("greet3", math.MaxInt64, arg)
	fmt.Println(leftover)

	require.Equal(t, nil, err)

	require.Equal(t, "greet3", res.(string))
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: nil case for []byte as arg
func TestBytesNil(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	addErr := addApis(t, hostApis)
	if addErr != nil {
		return
	}

	var (
		arg             []byte = nil
		wasmTimeRuntime types.AspectRuntime
		err             error
	)

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)
	res, leftover, err := wasmTimeRuntime.Call("testBytes", math.MaxInt64, arg)
	fmt.Println(leftover)

	require.Equal(t, nil, err)

	require.Equal(t, true, reflect.DeepEqual([]byte{}, res.([]byte)))
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}

// Test Case: long string as input/output params
func TestLongString(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./wasmtime/testdata/runtime_test.wasm"))

	hostApis := types.NewHostAPIRegistry(&mockedHostContext{}, wasmtime.Wrap)

	var (
		arg             string = ""
		wasmTimeRuntime types.AspectRuntime
		err             error
	)
	for i := 1; i <= 10000; i++ {
		arg += fmt.Sprintf("%-6d", i)
	}
	addEr := addApis(t, hostApis)
	if addEr != nil {
		return
	}

	wasmTimeRuntime, err = NewAspectRuntime(context.Background(), &mockedLogger{}, WASM, raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, leftover, err := wasmTimeRuntime.Call("greet", math.MaxInt64, arg)
		fmt.Println(leftover)

		require.Equal(t, nil, err)
		output := res.(string)

		require.Equal(t, "hello-greet-"+arg+"-hello-greet", output)
	}
	wasmTimeRuntime.Destroy() // to destroy the rt, in case of memory leak
}
