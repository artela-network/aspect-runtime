package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/artela-network/artelasdk/djpm"
	aspectType "github.com/artela-network/artelasdk/types"
	"github.com/artela-network/evm/vm"
	"github.com/artela-network/runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
	"github.com/shirou/gopsutil/mem"
)

const TIMES1000 = 1000

func BenchJoinPoint(usePool bool) {
	aspectType.GetHostApiHook = func() (aspectType.HostApi, error) {
		return nil, errors.New("not init")
	}

	cwd, _ := os.Getwd()
	raw, _ := os.ReadFile(path.Join(cwd, "./release.wasm"))

	name := "onTxReceive"
	api := "evmCall"

	hostApis := runtime.NewHostAPIRegistry()
	hostApis.AddApi("index", "__HostApi__", runtime.MethodName(api), func() {
		callStorage()
	})

	pool := runtime.NewRuntimePool(10)
	_ = pool
	t1 := time.Now()
	for i := 0; i < TIMES1000; i++ {
		if usePool {
			v, _ := mem.VirtualMemory()
			fmt.Println("index: ", i, ", total: ", v.Total/1024/1024, ", used %: ", v.UsedPercent, ", used: ", v.Used/1024/1024, ", free: ", v.Free/1024/1024)
			key, rt, err := pool.Runtime(runtime.WASM, raw, hostApis)
			if err != nil {
				// panic(err)
				fmt.Println(err)
			}
			rt.Call(name)
			pool.Return(key, rt)

		} else {
			rt, err := runtime.NewAspectRuntime(runtime.WASM, raw, hostApis)
			if err != nil {
				panic(err)
			}

			rt.Call(name)
			rt.Destroy()
		}
	}
	t2 := time.Now()
	fmt.Println("wasm elapsed time: ", t2.Sub(t1).Milliseconds(), "ms")
}

func callStorage() {
	djpm.NewAspect(getFunc, getBlockAspectsFunc, checkTxFunc, convertTxFunc)
	byteCode := "608060405234801561000f575f80fd5b50335f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506106998061005c5f395ff3fe608060405234801561000f575f80fd5b506004361061004a575f3560e01c80632e64cec11461004e5780632f54bf6e1461006c578063303904141461009c5780636057361d146100cd575b5f80fd5b6100566100e9565b604051610063919061027a565b60405180910390f35b610086600480360381019061008191906102fe565b6100f2565b6040516100939190610343565b60405180910390f35b6100b660048036038101906100b19190610498565b610158565b6040516100c492919061051f565b60405180910390f35b6100e760048036038101906100e29190610570565b6101b2565b005b5f600154905090565b5f805f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361014f5760019050610153565b5f90505b919050565b6002818051602081018201805184825260208301602085012081835280955050505050505f91509050805f015f9054906101000a900467ffffffffffffffff1690805f0160089054906101000a900463ffffffff16905082565b806001546101c091906105c8565b6001819055506040518060400160405280606567ffffffffffffffff1681526020016126ac63ffffffff1681525060026040516101fc9061064f565b90815260200160405180910390205f820151815f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055506020820151815f0160086101000a81548163ffffffff021916908363ffffffff16021790555090505050565b5f819050919050565b61027481610262565b82525050565b5f60208201905061028d5f83018461026b565b92915050565b5f604051905090565b5f80fd5b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102cd826102a4565b9050919050565b6102dd816102c3565b81146102e7575f80fd5b50565b5f813590506102f8816102d4565b92915050565b5f602082840312156103135761031261029c565b5b5f610320848285016102ea565b91505092915050565b5f8115159050919050565b61033d81610329565b82525050565b5f6020820190506103565f830184610334565b92915050565b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6103aa82610364565b810181811067ffffffffffffffff821117156103c9576103c8610374565b5b80604052505050565b5f6103db610293565b90506103e782826103a1565b919050565b5f67ffffffffffffffff82111561040657610405610374565b5b61040f82610364565b9050602081019050919050565b828183375f83830152505050565b5f61043c610437846103ec565b6103d2565b90508281526020810184848401111561045857610457610360565b5b61046384828561041c565b509392505050565b5f82601f83011261047f5761047e61035c565b5b813561048f84826020860161042a565b91505092915050565b5f602082840312156104ad576104ac61029c565b5b5f82013567ffffffffffffffff8111156104ca576104c96102a0565b5b6104d68482850161046b565b91505092915050565b5f67ffffffffffffffff82169050919050565b6104fb816104df565b82525050565b5f63ffffffff82169050919050565b61051981610501565b82525050565b5f6040820190506105325f8301856104f2565b61053f6020830184610510565b9392505050565b61054f81610262565b8114610559575f80fd5b50565b5f8135905061056a81610546565b92915050565b5f602082840312156105855761058461029c565b5b5f6105928482850161055c565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6105d282610262565b91506105dd83610262565b92508282019050808211156105f5576105f461059b565b5b92915050565b5f81905092915050565b7f746f6d00000000000000000000000000000000000000000000000000000000005f82015250565b5f6106396003836105fb565b915061064482610605565b600382019050919050565b5f6106598261062d565b915081905091905056fea264697066735822122016e99651fa7285eb8418243e3c44c04f00b0ddf7f64e1993ba7c1e76f81ea0ee64736f6c63430008150033"
	input := "6057361d0000000000000000000000000000000000000000000000000000000000000064"

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
	_, _, err = evm.Call(vm.AccountRef(sender), address, common.Hex2Bytes(input), math.MaxUint64, new(big.Int))
	if err != nil {
		panic(err)
	}

}

func BenchCallEVM() {
	// t1 := time.Now()
	// for i := 0; i < 10000; i++ {
	// 	CallProxyAndStorage()
	// }
	// t2 := time.Now()
	// fmt.Println("evm elapsed time: ", t2.Sub(t1).Milliseconds(), "ms")
	CallProxyAndStorage()
}

func CallProxyAndStorage() {
	t1 := time.Now()
	for i := 0; i < TIMES1000; i++ {
		var storageAddress common.Address
		var statedb *state.StateDB
		djpm.NewAspect(getFunc, getBlockAspectsFunc, checkTxFunc, convertTxFunc)
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

		proxyByteCode := "608060405234801561000f575f80fd5b5060f28061001c5f395ff3fe6080604052348015600e575f80fd5b50600436106026575f3560e01c8063f55332ab14602a575b5f80fd5b603960353660046091565b603b565b005b604051636057361d60e01b8152606460048201526001600160a01b03821690636057361d906024015f604051808303815f87803b1580156079575f80fd5b505af1158015608a573d5f803e3d5ffd5b5050505050565b5f6020828403121560a0575f80fd5b81356001600160a01b038116811460b5575f80fd5b939250505056fea2646970667358221220721f30e5c78520f95377cd0860e24108a130e612efd4bace1b4a0fabfe976de464736f6c63430008150033"
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
		proxyEvm := vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
		_, address, _, err := proxyEvm.Create(vm.AccountRef(sender), common.Hex2Bytes(proxyByteCode), math.MaxUint64, new(big.Int))
		if err != nil {
			panic(err)
		}
		statedb.Finalise(true)
		proxyEvm = vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)

		{
			djpm.NewAspect(getFunc, getBlockAspectsFunc, checkTxFunc, convertTxFunc)
			storageByteCode := "608060405234801561000f575f80fd5b50335f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506106998061005c5f395ff3fe608060405234801561000f575f80fd5b506004361061004a575f3560e01c80632e64cec11461004e5780632f54bf6e1461006c578063303904141461009c5780636057361d146100cd575b5f80fd5b6100566100e9565b604051610063919061027a565b60405180910390f35b610086600480360381019061008191906102fe565b6100f2565b6040516100939190610343565b60405180910390f35b6100b660048036038101906100b19190610498565b610158565b6040516100c492919061051f565b60405180910390f35b6100e760048036038101906100e29190610570565b6101b2565b005b5f600154905090565b5f805f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361014f5760019050610153565b5f90505b919050565b6002818051602081018201805184825260208301602085012081835280955050505050505f91509050805f015f9054906101000a900467ffffffffffffffff1690805f0160089054906101000a900463ffffffff16905082565b806001546101c091906105c8565b6001819055506040518060400160405280606567ffffffffffffffff1681526020016126ac63ffffffff1681525060026040516101fc9061064f565b90815260200160405180910390205f820151815f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055506020820151815f0160086101000a81548163ffffffff021916908363ffffffff16021790555090505050565b5f819050919050565b61027481610262565b82525050565b5f60208201905061028d5f83018461026b565b92915050565b5f604051905090565b5f80fd5b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102cd826102a4565b9050919050565b6102dd816102c3565b81146102e7575f80fd5b50565b5f813590506102f8816102d4565b92915050565b5f602082840312156103135761031261029c565b5b5f610320848285016102ea565b91505092915050565b5f8115159050919050565b61033d81610329565b82525050565b5f6020820190506103565f830184610334565b92915050565b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6103aa82610364565b810181811067ffffffffffffffff821117156103c9576103c8610374565b5b80604052505050565b5f6103db610293565b90506103e782826103a1565b919050565b5f67ffffffffffffffff82111561040657610405610374565b5b61040f82610364565b9050602081019050919050565b828183375f83830152505050565b5f61043c610437846103ec565b6103d2565b90508281526020810184848401111561045857610457610360565b5b61046384828561041c565b509392505050565b5f82601f83011261047f5761047e61035c565b5b813561048f84826020860161042a565b91505092915050565b5f602082840312156104ad576104ac61029c565b5b5f82013567ffffffffffffffff8111156104ca576104c96102a0565b5b6104d68482850161046b565b91505092915050565b5f67ffffffffffffffff82169050919050565b6104fb816104df565b82525050565b5f63ffffffff82169050919050565b61051981610501565b82525050565b5f6040820190506105325f8301856104f2565b61053f6020830184610510565b9392505050565b61054f81610262565b8114610559575f80fd5b50565b5f8135905061056a81610546565b92915050565b5f602082840312156105855761058461029c565b5b5f6105928482850161055c565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6105d282610262565b91506105dd83610262565b92508282019050808211156105f5576105f461059b565b5b92915050565b5f81905092915050565b7f746f6d00000000000000000000000000000000000000000000000000000000005f82015250565b5f6106396003836105fb565b915061064482610605565b600382019050919050565b5f6106598261062d565b915081905091905056fea264697066735822122016e99651fa7285eb8418243e3c44c04f00b0ddf7f64e1993ba7c1e76f81ea0ee64736f6c63430008150033"
			storageInput := "6057361d0000000000000000000000000000000000000000000000000000000000000064"

			// statedb1, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
			// _ = statedb1

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
			var err error
			_, storageAddress, _, err = evm.Create(vm.AccountRef(sender), common.Hex2Bytes(storageByteCode), math.MaxUint64, new(big.Int))
			if err != nil {
				panic(err)
			}
			statedb.Finalise(true)

			evm = vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
			_, _, err = evm.Call(vm.AccountRef(sender), storageAddress, common.Hex2Bytes(storageInput), math.MaxUint64, new(big.Int))
			if err != nil {
				panic(err)
			}
		}

		{
			// proxyByteCode := "608060405234801561000f575f80fd5b5060f28061001c5f395ff3fe6080604052348015600e575f80fd5b50600436106026575f3560e01c8063f55332ab14602a575b5f80fd5b603960353660046091565b603b565b005b604051636057361d60e01b8152606460048201526001600160a01b03821690636057361d906024015f604051808303815f87803b1580156079575f80fd5b505af1158015608a573d5f803e3d5ffd5b5050505050565b5f6020828403121560a0575f80fd5b81356001600160a01b038116811460b5575f80fd5b939250505056fea2646970667358221220721f30e5c78520f95377cd0860e24108a130e612efd4bace1b4a0fabfe976de464736f6c63430008150033"
			// proxyInput := Abi.
			proxyInput := hex.EncodeToString(PackProxy(storageAddress))
			// fmt.Println("hex address: ", storageAddress.Hex())
			// fmt.Println("input hex: ", proxyInput)

			// statedb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
			// vmConf := vm.Config{
			// 	ExtraEips: []int{3855},
			// }
			// sender := common.Address{}
			// sender.SetBytes([]byte("artela"))

			// vmctx := vm.BlockContext{
			// 	Transfer: func(vm.StateDB, common.Address, common.Address, *big.Int) {},
			// 	CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool {
			// 		return true
			// 	},
			// 	BlockNumber: big.NewInt(0),
			// }
			// evm := vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
			// _, address, _, err := evm.Create(vm.AccountRef(sender), common.Hex2Bytes(proxyByteCode), math.MaxUint64, new(big.Int))
			// if err != nil {
			// 	panic(err)
			// }
			// statedb.Finalise(true)

			// evm = vm.NewEVM(vmctx, vm.TxContext{Msg: new(MockMessage)}, statedb, params.AllEthashProtocolChanges, vmConf)
			_, _, err = proxyEvm.Call(vm.AccountRef(sender), address, common.Hex2Bytes(proxyInput), math.MaxUint64, new(big.Int))
			if err != nil {
				panic(err)
			}
		}
	}
	t2 := time.Now()
	fmt.Println("evm elapsed time: ", t2.Sub(t1).Milliseconds(), "ms")
}

// go elapsed time of running fibo
func benchFibo() {
	t1 := time.Now()
	var sum uint64 = 0
	for i := 0; i < TIMES1000; i++ {
		m := uint64(90)
		n := uint64(200000)

		sum = 0
		for d := uint64(0); d < n; d++ {
			var a uint64 = 1
			var b uint64 = 1
			for i := uint64(2); i < m; i++ {
				var c uint64 = a + b
				a = b
				b = c
				if c > sum {
					sum = c
				}
			}
			sum++
		}
	}
	fmt.Println(sum)
	t2 := time.Now()
	fmt.Println("go elapsed time: ", t2.Sub(t1).Milliseconds(), "ms")
}
