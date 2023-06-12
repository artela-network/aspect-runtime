package runtime

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"reflect"
)

type RuntimePool struct {
	cap     int
	engines map[string]AspectRuntime
}

func NewRuntimePool(capacity int) *RuntimePool {
	return &RuntimePool{
		cap:     capacity,
		engines: make(map[string]AspectRuntime, capacity),
	}
}

// The Runtime retrieves an aspect runtime from the pool.
// The key used to access the pool is the hash value obtained from combining the runtimeType, code, and APIs.
//
// If the aspect runtime does not exist in the pool, a new runtime is created and cached in the pool.
//
// The preRun parameter refers to the function names used to clear the memory of the previous run, or something else.
// If preRun executes failed, it will continue to create a new runtime and cache in the pool.
func (pool *RuntimePool) Runtime(runtimeType RuntimeType, code []byte, apis *HostAPIRegistry, forceRefresh bool, preRun ...string) (AspectRuntime, error) {
	hash := hashOfRuntimeArgs(runtimeType, code, apis)
	engine, ok := pool.engines[hash]
	if ok {
		if !forceRefresh {
			preRunOK := true
			for _, pr := range preRun {
				_, err := engine.Call(pr)
				if err != nil {
					preRunOK = false
					break
				}
			}
			if preRunOK {
				return engine, nil
			}
		}
		// call preRun error, abandon then engine
		// create a new one instead
		delete(pool.engines, hash)
	}

	engine, err := NewAspectRuntime(runtimeType, code, apis)
	if err != nil {
		return nil, err
	}
	pool.engines[hash] = engine
	return engine, nil
}

func hashOfRuntimeArgs(runtimeType RuntimeType, code []byte, apis *HostAPIRegistry) string {
	return hex.EncodeToString(Hash(runtimeType, code, apis))
}

func Hash(objs ...interface{}) []byte {
	sha := crypto.SHA256.New()
	for _, obj := range objs {
		fmt.Fprint(sha, reflect.TypeOf(obj))
		fmt.Fprint(sha, obj)
	}
	return sha.Sum(nil)
}
