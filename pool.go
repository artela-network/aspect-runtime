package runtime

import (
	"context"
	"crypto"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/artela-network/aspect-runtime/types"
	"github.com/google/uuid"
)

type (
	Key  string
	Hash string
)

type Entry struct {
	prev *Entry // point to prev entry
	next *Entry // point to next entry

	sprev *Entry // for sub list
	snext *Entry // for sub list

	key     Key // build with id:hash
	runtime types.AspectRuntime
}

func (e *Entry) destroy() {
	if e.runtime != nil {
		e.runtime.Destroy()
	}
}

type EntryList struct {
	sync.Mutex

	head *Entry
	tail *Entry

	subs map[Hash]*EntryList

	cap int
	len int
}

func NewEntryList(cap int) *EntryList {
	head := &Entry{}
	tail := &Entry{}
	head.next = tail
	tail.prev = head
	head.snext = tail
	tail.sprev = head

	return &EntryList{
		head: head,
		tail: tail,
		subs: make(map[Hash]*EntryList),
		len:  0,
		cap:  cap,
	}
}

func (list *EntryList) Len() int {
	list.Lock()
	defer list.Unlock()

	return list.len
}

func (list *EntryList) PushFront(entry *Entry) {
	list.Lock()
	defer list.Unlock()

	if list.len >= list.cap {
		end := list.tail.prev
		list.remove(end)
		go end.destroy()
	}

	list.len++
	entry.next = list.head.next
	entry.prev = list.head

	entry.prev.next = entry
	entry.next.prev = entry

	key := entry.key
	_, hash := split(key)
	sub, ok := list.subs[hash]
	if !ok {
		sub = NewEntryList(list.cap)
		list.subs[hash] = sub
	}
	sub.len++

	entry.snext = sub.head.snext
	entry.sprev = sub.head

	entry.sprev.snext = entry
	entry.snext.sprev = entry
}

func (list *EntryList) remove(entry *Entry) {
	list.len--
	entry.prev.next = entry.next
	entry.next.prev = entry.prev

	entry.sprev.snext = entry.snext
	entry.snext.sprev = entry.sprev
	key := entry.key
	_, hash := split(key)
	sub := list.subs[hash]
	sub.len--
	if sub.len == 0 {
		delete(list.subs, hash)
	}
}

func (list *EntryList) PopFront(hash Hash) (*Entry, bool) {
	list.Lock()
	defer list.Unlock()

	if sub, ok := list.subs[hash]; ok && sub.len > 0 {
		entry := sub.head.snext
		list.remove(entry)
		return entry, true
	}
	return nil, false
}

type RuntimePool struct {
	sync.Mutex

	cache  *EntryList
	logger types.Logger
	ctx    context.Context
}

func NewRuntimePool(ctx context.Context, logger types.Logger, capacity int) *RuntimePool {
	return &RuntimePool{
		cache:  NewEntryList(capacity),
		logger: logger,
		ctx:    ctx,
	}
}

func (pool *RuntimePool) Len() int {
	return pool.cache.Len()
}

func (pool *RuntimePool) Runtime(ctx context.Context, rtType RuntimeType, code []byte, apis *types.HostAPIRegistry) (string, types.AspectRuntime, error) {
	startTime := time.Now()

	hash := hashOfRuntimeArgs(rtType, code)
	key, rt, err := pool.get(hash)
	if err == nil && rt.ResetStore(ctx, apis) == nil {
		pool.logger.Debug("runtime pool cache hit", "duration", time.Since(startTime).String(),
			"hash", hash, "key", key)
		return string(key), rt, nil
	}

	rt, err = NewAspectRuntime(pool.ctx, pool.logger, rtType, code, apis)

	if err != nil {
		return "", nil, err
	}

	id := uuid.New()
	keyStr := join(id.String(), hash)

	pool.logger.Debug("runtime pool cache miss", "duration", time.Since(startTime).String(),
		"hash", hash, "key", keyStr)
	return keyStr, rt, nil
}

func (pool *RuntimePool) get(hash Hash) (Key, types.AspectRuntime, error) {
	entry, ok := pool.cache.PopFront(hash)
	if !ok {
		return "", nil, errors.New("not found")
	}
	return entry.key, entry.runtime, nil
}

// Return returns a runtime to the pool
func (pool *RuntimePool) Return(key string, runtime types.AspectRuntime) {
	// free the hostapis and ctx injected to types, in case that go runtime GC failed
	pool.logger.Debug("returning runtime", "key", key)

	runtime.Reset()

	pool.logger.Debug("runtime destroyed", "key", key)

	entry := &Entry{
		key:     Key(key),
		runtime: runtime,
	}

	pool.cache.PushFront(entry)

	pool.logger.Debug("runtime returned", "key", key)
}

func hashOfRuntimeArgs(runtimeType RuntimeType, code []byte) Hash {
	h := sha1.New()
	var rttype [1]byte
	rttype[0] = byte(runtimeType)
	h.Write(rttype[:])
	h.Write(code)
	return Hash(hex.EncodeToString(h.Sum(nil)))
}

func hashOf(objs ...interface{}) []byte {
	sha := crypto.SHA256.New()
	for _, obj := range objs {
		fmt.Fprint(sha, reflect.TypeOf(obj))
		fmt.Fprint(sha, obj)
	}
	return sha.Sum(nil)
}

func join(id string, hash Hash) string {
	return fmt.Sprintf("%s:%s", id, hash)
}

func split(key Key) (string, Hash) {
	s := strings.Split(string(key), ":")
	return s[0], Hash(s[1])
}
