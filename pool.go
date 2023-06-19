package runtime

import (
	"container/list"
	"crypto"
	"encoding/hex"
	"fmt"
	"reflect"
	"sync"
)

type (
	entry struct {
		key     string
		runtime AspectRuntime
	}

	RuntimePool struct {
		sync.Mutex

		capacity int
		// list.Value = AspectRuntime
		keys *list.List

		// key: hash of args to build the AspectRuntime
		cache map[string]*list.Element
	}
)

func NewRuntimePool(capacity int) *RuntimePool {
	return &RuntimePool{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		keys:     list.New(),
	}
}

func (pool *RuntimePool) Capacity() int {
	return pool.capacity
}

func (pool *RuntimePool) Len() int {
	return len(pool.cache)
}

// RuntimeForceRefresh create a new AspectRuntime and force to refresh of runtime pool
func (pool *RuntimePool) RuntimeForceRefresh(rtType RuntimeType, code []byte, apis *HostAPIRegistry, preRun ...string) (string, AspectRuntime, error) {
	return pool.get(rtType, code, apis, true, preRun...)
}

// Get retrieves an aspect runtime from the pool.
// The key used to access the pool is the hash value obtained from combining the runtimeType, code, and APIs.
//
// If the aspect runtime does not exist in the pool, a new runtime is created and cached in the pool.
//
// The preRun parameter refers to the function names used to clear the memory of the previous run, or something else.
// If preRun executes failed, it will continue to create a new runtime and cache in the pool.
func (pool *RuntimePool) Runtime(rtType RuntimeType, code []byte, apis *HostAPIRegistry, preRun ...string) (string, AspectRuntime, error) {
	return pool.get(rtType, code, apis, false, preRun...)
}

// PutBack return a runtime to the pool
func (pool *RuntimePool) PutBack(key string, runtime AspectRuntime) {
	pool.Lock()
	defer pool.Unlock()

	if elem, ok := pool.cache[key]; ok {
		pool.keys.MoveToFront(elem)
		return
	}

	if pool.Len() >= pool.Capacity() {
		// remove the last from the pool
		last := pool.keys.Back()
		pool.remove(last.Value.(*entry).key, last)
	}

	// add new to front
	pool.add(key, runtime)
}

// TODO reset memory instead of pssing preRun.
func (pool *RuntimePool) get(rtType RuntimeType, code []byte, apis *HostAPIRegistry, forceRefresh bool, preRun ...string) (string, AspectRuntime, error) {
	pool.Lock()
	defer pool.Unlock()

	hash := hashOfRuntimeArgs(rtType, code, apis)
	elem, ok := pool.cache[hash]
	if ok {
		rt := elem.Value.(*entry).runtime

		preRunOK := true
		if !forceRefresh {
			for _, pr := range preRun {
				_, err := rt.Call(pr)
				if err != nil {
					preRunOK = false
					break
				}
			}
		}
		// remove from the pool, ignore it is borrowed or removed.
		pool.remove(hash, elem)

		if preRunOK {
			return hash, rt, nil
		}
	}

	rt, err := NewAspectRuntime(rtType, code, apis)
	if err != nil {
		return "", nil, err
	}

	// do not put the runtime to the pool, until after using it and putting it back.
	return hash, rt, nil
}

func (pool *RuntimePool) remove(key string, elem *list.Element) {
	pool.keys.Remove(elem)
	delete(pool.cache, key)
}

func (pool *RuntimePool) add(key string, runtime AspectRuntime) {
	new := pool.keys.PushFront(&entry{key, runtime})
	pool.cache[key] = new
}

func hashOfRuntimeArgs(runtimeType RuntimeType, code []byte, apis *HostAPIRegistry) string {
	return hex.EncodeToString(hash(runtimeType, code, apis))
}

func hash(objs ...interface{}) []byte {
	sha := crypto.SHA256.New()
	for _, obj := range objs {
		fmt.Fprint(sha, reflect.TypeOf(obj))
		fmt.Fprint(sha, obj)
	}
	return sha.Sum(nil)
}
