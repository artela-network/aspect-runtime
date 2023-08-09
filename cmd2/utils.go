package main

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/artela-network/artelasdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

/*
|
| mock
|
|
|
|
|
|
*/

func getFunc(int64, common.Address) ([]*types.AspectCode, error) {
	return nil, nil
}

func getBlockAspectsFunc(int64) ([]*types.AspectCode, error) {
	return nil, nil
}

func checkTxFunc(tx sdk.Msg) bool {
	return true
}

func convertTxFunc(tx sdk.Msg) *ethtypes.Transaction {
	return nil
}

type MockMessage struct {
}

func (m MockMessage) From() common.Address {
	return common.Address{}
}

func (m MockMessage) To() *common.Address {
	return &common.Address{}
}

func (m MockMessage) GasPrice() *big.Int {
	return big.NewInt(0)
}

func (m MockMessage) GasFeeCap() *big.Int {
	return big.NewInt(0)
}

func (m MockMessage) GasTipCap() *big.Int {
	return big.NewInt(0)
}

func (m MockMessage) Gas() uint64 {
	return 0
}

func (m MockMessage) Value() *big.Int {
	return big.NewInt(0)
}

func (m MockMessage) Nonce() uint64 {
	return 0
}

func (m MockMessage) IsFake() bool {
	return true
}

func (m MockMessage) Data() []byte {
	return nil
}

func PackProxy(addr common.Address) []byte {
	const definition = `[{"inputs":[{"internalType":"address","name":"callee","type":"address"}],"name":"call","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	abi, err := abi.JSON(strings.NewReader(definition))
	if err != nil {
		panic(err)
	}
	out, err := abi.Pack("call", addr)
	if err != nil {
		panic(err)
	}

	return out
}

func PackFibo(m int64, n int64) string {
	const definition = `[{"inputs":[{"internalType":"uint256","name":"m","type":"uint256"},{"internalType":"uint256","name":"n","type":"uint256"}],"name":"fibU64Max","outputs":[],"stateMutability":"pure","type":"function"}]`

	abi, err := abi.JSON(strings.NewReader(definition))
	if err != nil {
		panic(err)
	}
	arg1 := big.NewInt(m)
	arg2 := big.NewInt(n)
	out, err := abi.Pack("fibU64Max", arg1, arg2)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(out)
}

func getFiboCode() string {
	return "608060405234801561000f575f80fd5b506101188061001d5f395ff3fe6080604052348015600e575f80fd5b50600436106026575f3560e01c806316a2271c14602a575b5f80fd5b603960353660046084565b603b565b005b5f5b81811015607f5760018060025b85811015606e575f605a838560b7565b92935081905060678160cd565b915050604a565b5050508060799060cd565b9050603d565b505050565b5f80604083850312156094575f80fd5b50508035926020909101359150565b634e487b7160e01b5f52601160045260245ffd5b8082018082111560c75760c760a3565b92915050565b5f6001820160db5760db60a3565b506001019056fea264697066735822122044cff7e990bb4ee694071f4bc121443479e70cb6f74df157dc6e223fe2f3466b64736f6c63430008150033"
}
