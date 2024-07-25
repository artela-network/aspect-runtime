package types

import (
	"math"
)

const (
	EVMGasToWASMGasMultiplier = 1000
	MaxGas                    = math.MaxInt64 / EVMGasToWASMGasMultiplier
)
