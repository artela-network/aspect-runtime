package wasmtime

import (
	"os"
	"path"
	"testing"

	"github.com/artela-network/runtime"
	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./testdata/runtime_test.wasm"))

	hostApis := runtime.NewHostAPICollection()

	var (
		arg1 string = "abcd"
		// arg2            string = "1234"
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)
	hostApis.AddApi("index", "test", "hello", func(arg string) string {
		return arg + "1234"
	})

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("test", arg1)
		require.Equal(t, nil, err)

		require.Equal(t, "abcd1234", string(res))
	}
}
