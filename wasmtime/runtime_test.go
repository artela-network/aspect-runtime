package wasmtime

import (
	"os"
	"testing"

	"github.com/artela-network/runtime"
	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	raw, _ := os.ReadFile("/Users/likun/go/src/github.com/artela-network/wasm-poc-ts/wasm-aspect/build/release.wasm")

	hostApis := runtime.NewHostAPICollection()

	var (
		arg string = "abcd"
		// arg2            string = "1234"
		wasmTimeRuntime runtime.WASMRuntime
		err             error
	)
	hostApis.AddApi("index", "test", "hello", func(arg1 []byte, arg2 string) string {
		return string(arg1) + string(arg2) + "1234"
	})

	wasmTimeRuntime, err = NewWASMTimeRuntime(raw, hostApis)
	require.Equal(t, nil, err)

	{
		res, err := wasmTimeRuntime.Call("greet", arg)
		require.Equal(t, nil, err)

		require.Equal(t, "abcd1234", string(res))
	}
}
