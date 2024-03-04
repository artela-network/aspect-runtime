package types

import (
	"math"
)

const (
	EVMGasToWASMGasMultiplier = 100
	MaxGas                    = math.MaxInt64 / EVMGasToWASMGasMultiplier
)
