package main

import (
	"testing"
)

func Test1(t *testing.T) {
	BenchJoinPoint(true)
	CallProxyAndStorage()
}

func Test2(t *testing.T) {
	// BenchEVM()
	BenchWasm(90, 20, 1)

}
