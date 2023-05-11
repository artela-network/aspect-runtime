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
		arg             string = "abcd"
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)
	hostApis.AddApi("index", "test", "hello", func(arg string) string {
		return "hello-" + arg + "-hello"
	})

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)

		require.Equal(t, "hello-greet-abcd-hello-greet", res.(string))
	}
}
