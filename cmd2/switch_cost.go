package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/artela-network/artelasdk/djpm"
	"github.com/artela-network/evm/vm"
	"github.com/artela-network/runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
)

const TIME_SWITH = 8000

func BenchCostOfSwithContext() {
	fmt.Println("\nloops: ", TIME_SWITH)
	BenchEVMSwitch()
	BenchWasmSwitch()
}

func BenchEVMSwitch() {
	djpm.NewAspect(getFunc, getBlockAspectsFunc, checkTxFunc, convertTxFunc)
	byteCode := getFiboCode()
	input := PackFibo(90, 0)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	vmConf := vm.Config{
		ExtraEips: []int{3855},
	}
	sender := common.Address{}
	sender.SetBytes([]byte("artela"))

	vmctx := vm.BlockContext{
		Transfer: func(vm.StateDB, common.Address, common.Address, *big.Int) {},
		CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool {
			return true
		},
		BlockNumber: big.NewInt(0),
	}
	evm := vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
	_, address, _, err := evm.Create(vm.AccountRef(sender), common.Hex2Bytes(byteCode), math.MaxUint64, new(big.Int))
	if err != nil {
		panic(err)
	}
	statedb.Finalise(true)

	evm = vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
	inputData := common.Hex2Bytes(input)

	t1 := time.Now()
	for i := 0; i < TIME_SWITH; i++ {
		_, _, err = evm.Call(vm.AccountRef(sender), address, inputData, math.MaxUint64, new(big.Int))
		if err != nil {
			panic(err)
		}
	}
	t2 := time.Now()
	duration := t2.Sub(t1)
	fmt.Println("evm switch context elapsed time: ", duration.Milliseconds(), "ms")
	fmt.Printf("evm average: %dμs\n", duration.Microseconds()/TIME_SWITH)
}

func BenchWasmSwitch() {
	raw, _ := os.ReadFile("/Users/likun/go/src/github.com/test/asexample/build/release.wasm")

	name := "joinPointMock"

	hostApis := runtime.NewHostAPIRegistry()
	// hostApis.AddApi("index", "__HostApi__", runtime.MethodName(api), func() {
	// 	callStorage()
	// })

	pool := runtime.NewRuntimePool(10)
	_ = pool

	wasmCall := func(m, n int64) {
		key, rt, err := pool.Runtime(runtime.WASMTime, raw, hostApis)
		if err != nil {
			panic(err)
		}

		_, err = rt.Call(name, m, n)
		if err != nil {
			panic(err)
		}
		pool.Return(key, rt)
	}

	t1 := time.Now()
	for i := 0; i < TIME_SWITH; i++ {
		wasmCall(90, 0)
	}
	t2 := time.Now()
	duration := t2.Sub(t1)
	fmt.Println("wasm switch context elapsed time: ", duration.Milliseconds(), "ms")
	fmt.Printf("wasm average: %dμs\n", duration.Microseconds()/TIME_SWITH)
}
