package pools

import (
	"fmt"
	"sync"
)

type IdPool struct {
	sync.RWMutex
	used      map[int64]bool
	startId   int64
	maxUsedId int64
}

func NewIdPool(startId int64) *IdPool {

	return &IdPool{
		used:      make(map[int64]bool, 0),
		startId:   startId,
		maxUsedId: startId,
	}
}

func (pool *IdPool) Get() int64 {

	pool.Lock()
	defer pool.Unlock()

	for id := range pool.used {

		delete(pool.used, id)

		return id
	}

	pool.maxUsedId = pool.maxUsedId + 1

	return pool.maxUsedId
}

func (pool *IdPool) Put(id int64) {

	pool.Lock()
	defer pool.Unlock()

	if id <= pool.startId || id > pool.maxUsedId {

		panic(fmt.Errorf("IDPool.Put(%v): invalid value, must be in the range [%v,%v]", id, pool.startId, pool.maxUsedId))
	}

	if pool.used[id] {

		panic(fmt.Errorf("IDPool.Put(%v): can't put value that was already recycled", id))
	}

	pool.used[id] = true
}

func (pool *IdPool) MaxUsedCount() int64 {

	pool.RLock()
	defer pool.RUnlock()

	return pool.maxUsedId - pool.startId
}

func (pool *IdPool) CurrUsedCount() int64 {

	pool.RLock()
	defer pool.RUnlock()

	return pool.maxUsedId - pool.startId - int64(len(pool.used))
}
