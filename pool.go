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
		ready   bool
		key     string
		runtime AspectRuntime
	}

	// nolint
	RuntimePool struct {
		sync.Mutex

		capacity int

		// list.Value = &entry
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
func (pool *RuntimePool) RuntimeForceRefresh(rtType RuntimeType, code []byte, apis *HostAPIRegistry) (string, AspectRuntime, error) {
	return pool.get(rtType, code, apis, true)
}

// Runtime retrieves an aspect runtime from the pool.
// The key used to access the pool is the hash value obtained from combining the runtimeType, code, and APIs.
//
// If the aspect runtime does not exist in the pool, a new runtime is created and cached in the pool.
func (pool *RuntimePool) Runtime(rtType RuntimeType, code []byte, apis *HostAPIRegistry) (string, AspectRuntime, error) {
	return pool.get(rtType, code, apis, false)
}

// Return returns a runtime to the pool
func (pool *RuntimePool) Return(key string, runtime AspectRuntime) {
	pool.Lock()
	defer pool.Unlock()

	wfuncs := runtime.HostFuncs()

	// free the hostapis and ctx injected to types, in case that go runtime GC failed
	runtime.Destroy()

	ready := false
	registry := &HostAPIRegistry{}
	registry.AddFuncs(wfuncs)
	if err := runtime.ResetStore(registry); err == nil {
		ready = true
	}

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
	pool.add(key, runtime, ready)
}

func (pool *RuntimePool) get(rtType RuntimeType, code []byte, apis *HostAPIRegistry, forceRefresh bool) (string, AspectRuntime, error) {
	pool.Lock()
	defer pool.Unlock()

	hash := hashOfRuntimeArgs(rtType, code)
	elem, ok := pool.cache[hash]
	if ok {
		// remove from the pool, either it is borrowed or removed.
		pool.remove(hash, elem)

		if !forceRefresh {
			rt := elem.Value.(*entry).runtime
			if elem.Value.(*entry).ready {
				return hash, rt, nil
			}

			if err := rt.ResetStore(apis); err == nil {
				return hash, rt, nil
			}
			// if call reset failed, continue to create a new one
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

func (pool *RuntimePool) add(key string, runtime AspectRuntime, ready bool) {
	new := pool.keys.PushFront(&entry{
		ready:   ready,
		key:     key,
		runtime: runtime,
	})
	pool.cache[key] = new
}

func hashOfRuntimeArgs(runtimeType RuntimeType, code []byte) string {
	return hex.EncodeToString(hash(runtimeType, code))
}

func hash(objs ...interface{}) []byte {
	sha := crypto.SHA256.New()
	for _, obj := range objs {
		fmt.Fprint(sha, reflect.TypeOf(obj))
		fmt.Fprint(sha, obj)
	}
	return sha.Sum(nil)
}
