package runtime

func init() {
	enginePool[WASMTime] = NewWASMTimeRuntime
	enginePool[WAZero] = NewWazeroRuntime
}
