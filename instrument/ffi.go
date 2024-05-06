package instrument

// #cgo CFLAGS:-I${SRCDIR}/build/include
// #cgo !windows LDFLAGS:-lm -ldl -pthread
// #cgo windows CFLAGS:-DWASM_API_EXTERN= -DWASI_API_EXTERN=
// #cgo windows LDFLAGS:-luserenv -lole32 -lntdll -lws2_32 -lkernel32 -lbcrypt
// #cgo linux,amd64 LDFLAGS:-L${SRCDIR}/build/linux-x86_64 -laspect_wasm_instrument
// #cgo linux,arm64 LDFLAGS:-L${SRCDIR}/build/linux-aarch64 -laspect_wasm_instrument
// #cgo darwin,amd64 LDFLAGS:-L${SRCDIR}/build/darwin-x86_64 -laspect_wasm_instrument
// #cgo darwin,arm64 LDFLAGS:-L${SRCDIR}/build/darwin-aarch64 -laspect_wasm_instrument
// #cgo windows,amd64 LDFLAGS:-L${SRCDIR}/build/windows-x86_64 -laspect_wasm_instrument
// #include "aspect_wasm_instrument.h"
import "C"

import (
	"fmt"
	"unsafe"
)

func WasmInstrument(rawModule []byte) ([]byte, error) {
	cRawModule := (*C.uchar)(unsafe.Pointer(&rawModule[0]))
	len := C.size_t(len(rawModule))

	cResult := C.wasm_instrument(cRawModule, len)
	if cResult.ptr == nil {
		return nil, fmt.Errorf("wasm_instrument failed")
	}

	output := C.GoBytes(cResult.ptr, C.int(cResult.len))
	C.wasm_instrument_free(cResult.ptr)

	return output, nil
}
