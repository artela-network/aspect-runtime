package instrument

import (
	"os"
	"testing"
)

func BenchmarkInstrument(b *testing.B) {
	byteCode, err := os.ReadFile("/Users/jack/Projects/go/src/github.com/artela-network/runtime/instrument/testdata/release.wasm")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = WasmInstrument(byteCode)
		if err != nil {
			b.Fatal(err)
		}
	}
}
