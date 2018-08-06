package pools

import (
	"reflect"
	"strings"
	"testing"
)

func (p *IdPool) want(want *IdPool, t *testing.T) {

	if p.maxUsedId != want.maxUsedId {

		t.Errorf("pool.maxUsed = %#v, want %#v", p.maxUsedId, want.maxUsedId)
	}

	if !reflect.DeepEqual(p.used, want.used) {

		t.Errorf("pool.used = %#v, want %#v", p.used, want.used)
	}
}

func TestIDPoolFirstGet(t *testing.T) {

	pool := NewIdPool(0)

	if got := pool.Get(); got != 1 {

		t.Errorf("pool.Get() = %v, want 1", got)
	}

	pool.want(&IdPool{used: map[int64]bool{}, maxUsedId: 1}, t)
}

func TestIDPoolSecondGet(t *testing.T) {

	pool := NewIdPool(100)
	pool.Get()

	if got := pool.Get(); got != 2 {

		t.Errorf("pool.Get() = %v, want 2", got)
	}

	pool.want(&IdPool{used: map[int64]bool{}, maxUsedId: 1, startId: 100}, t)
}

func TestIDPoolPutToUsedSet(t *testing.T) {

	pool := NewIdPool(100)
	id1 := pool.Get()
	pool.Get()
	pool.Put(id1)

	pool.want(&IdPool{used: map[int64]bool{101: true}, maxUsedId: 102, startId: 100}, t)
}

func TestIDPoolPutMaxUsed(t *testing.T) {
	pool := NewIdPool(0)
	id1 := pool.Get()
	pool.Put(id1)

	pool.want(&IdPool{used: map[int64]bool{1: true}, maxUsedId: 1}, t)
}

func TestIDPoolMaxUsedCount(t *testing.T) {
	pool := NewIdPool(0)
	pool.Get()
	id2 := pool.Get()
	pool.Put(id2)
	pool.Get()
	id3 := pool.Get()
	pool.Put(id3)

	if hisMax := pool.MaxUsedCount(); hisMax != 3 {

		t.Errorf("pool.MaxUsedCount() = %v, want 3", hisMax)
	}

	if curr := pool.CurrUsedCount(); curr != 2 {

		t.Errorf("pool.CurrUsedCount() = %v, want 2", curr)
	}

	pool.want(&IdPool{used: map[int64]bool{3: true}, maxUsedId: 3}, t)
}

func TestIDPoolGetFromUsedSet(t *testing.T) {

	pool := NewIdPool(0)
	id1 := pool.Get()
	pool.Get()
	pool.Put(id1)

	if got := pool.Get(); got != 1 {

		t.Errorf("pool.Get() = %v, want 1", got)
	}

	pool.want(&IdPool{used: map[int64]bool{}, maxUsedId: 2}, t)
}

func wantError(want string, t *testing.T) {

	rec := recover()
	if rec == nil {

		t.Errorf("expected panic, but there wasn't one")
	}

	err, ok := rec.(error)
	if !ok || !strings.Contains(err.Error(), want) {

		t.Errorf("wrong error, got '%v', want '%v'", err, want)
	}
}

func TestIDPoolPut0(t *testing.T) {

	pool := NewIdPool(0)

	defer wantError("invalid value", t)

	pool.Put(0)
}

func TestIDPoolPutInvalid(t *testing.T) {

	pool := NewIdPool(0)
	pool.Get()

	defer wantError("invalid value", t)

	pool.Put(5)
}

func TestIDPoolPutDuplicate(t *testing.T) {

	pool := NewIdPool(0)
	pool.Get()
	pool.Get()
	pool.Put(1)

	defer wantError("already recycled", t)

	pool.Put(1)
}
