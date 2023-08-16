#!/bin/bash

#
# run test1, make m = 10, n = 160
#
echo ====================running test2====================
m=10
n=160
echo -----------wasm-----------
./cmd2 wasm $m $n 1
sleep 10
for i in {1..9}; do
    ./cmd2 wasm $m $n $((6 * $i))
    sleep 10
done

echo -----------evm-----------
./cmd2 evm $m $n 1
sleep 10
for i in {1..9}; do
    ./cmd2 evm $m $n $((6 * $i))
    sleep 10
done

#
# run test2, make jpc = 10
#
echo
echo ====================running test2====================
jpc=10

echo -----------wasm-----------
for j in {0..9}; do
    ./cmd2 wasm 10 $((400 * $j)) $jpc
    sleep 10
done

echo -----------evm-----------
for j in {0..9}; do
    ./cmd2 evm 10 $((400 * $j)) $jpc
    sleep 10
done

#
# run test3, make execution time = 1000 μs
#
