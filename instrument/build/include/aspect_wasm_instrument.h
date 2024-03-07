#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    void* ptr;
    size_t len;
} WasmInstrumentResult;

WasmInstrumentResult wasm_instrument(const unsigned char* raw_module, size_t len);
void wasm_instrument_free(void* ptr);

#ifdef __cplusplus
}
#endif
