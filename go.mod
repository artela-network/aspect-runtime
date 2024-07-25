module github.com/artela-network/aspect-runtime

go 1.20

require (
	github.com/bytecodealliance/wasmtime-go/v20 v20.0.0
	github.com/ethereum/go-ethereum v1.12.0
	github.com/google/uuid v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bytecodealliance/wasmtime-go/v20 => github.com/artela-network/wasmtime-go/v20 v20.0.3
