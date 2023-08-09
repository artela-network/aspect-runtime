package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/artela-network/artelasdk/djpm"
	"github.com/artela-network/evm/vm"
	"github.com/artela-network/runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
)

func BenchWasm(addCount, loopCount int64, jp int64) {

	///
	/// create of wasm instance
	///
	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./release.wasm"))

	name := "joinPointMock"
	// api := "evmCall"

	hostApis := runtime.NewHostAPIRegistry()
	// hostApis.AddApi("index", "__HostApi__", runtime.MethodName(api), func() {
	// 	callStorage()
	// })

	pool := runtime.NewRuntimePool(10)
	_ = pool

	loop := int64(0)
	if jp != 0 {
		loop = loopCount / jp
	}

	wasmCall := func(i, j int64) {
		key, rt, err := pool.Runtime(runtime.WASM, raw, hostApis)
		if err != nil {
			panic(err)
		}

		_, err = rt.Call(name, i, j)
		if err != nil {
			panic(err)
		}
		pool.Return(key, rt)
	}

	t1 := time.Now()
	for i := 0; i < TIMES; i++ {
		if jp == 0 {
			wasmCall(addCount, loop)
		} else {
			for k := int64(0); k < jp; k++ {
				wasmCall(addCount, loop)
			}
		}
	}
	t2 := time.Now()
	duration := t2.Sub(t1)
	fmt.Printf("loops: %d, join point: %d, wasm elapsed time: %d ms\n", TIMES, jp, duration.Milliseconds())
	fmt.Printf("adds: %d, wasm average: %d μs\n\n", addCount*loopCount, duration.Microseconds()/TIMES)
}

func BenchEVM(addCount, loopCount int64, jp int64) {
	djpm.NewAspect(getFunc, getBlockAspectsFunc, checkTxFunc, convertTxFunc)
	byteCode := getFiboCode()
	// input := PackFibo(FIBO_ADD, FIBO_LOOP)

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

	loop := int64(0)
	if jp != 0 {
		loop = loopCount / jp
	}
	var (
		input     = PackFibo(addCount, loop)
		inputData = common.Hex2Bytes(input)
	)

	totalGas := uint64(0)
	t1 := time.Now()
	for i := 0; i < TIMES; i++ {
		if jp == 0 {
			gas := uint64(0)
			_, gas, err = evm.Call(vm.AccountRef(sender), address, inputData, math.MaxUint64, new(big.Int))
			if err != nil {
				panic(err)
			}
			totalGas += (math.MaxUint64 - gas)
		} else {
			for k := int64(0); k < jp; k++ {
				gas := uint64(0)
				_, gas, err = evm.Call(vm.AccountRef(sender), address, inputData, math.MaxUint64, new(big.Int))
				if err != nil {
					panic(err)
				}
				totalGas += (math.MaxUint64 - gas)
			}
		}
	}
	t2 := time.Now()
	duration := t2.Sub(t1)
	fmt.Printf("loops: %d, join point: %d, evm elapsed time: %d ms, total gas: %d\n", TIMES, jp, duration.Milliseconds(), totalGas)
	fmt.Printf("add %d, evm average: %d μs, gas average: %d\n\n", addCount*loopCount, duration.Microseconds()/TIMES, totalGas/TIMES)
}
