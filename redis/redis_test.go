package redis

import (
	"reflect"
	"testing"
	"time"

	"gamelib/codec"
)

func init() {

	InitRedis(codec.MsgPack, codec.UnMsgPack, NewRedisConf("main", "127.0.0.1", "6378", 0), NewRedisConf("main", "127.0.0.1", "6379", 64))
}

func TestRedis_GetSet(t *testing.T) {

	id1 := 100

	r, err := UseRedis("main", uint64(id1))
	if err != nil {

		t.Error("UseRedis:", "main", id1, err)
	}

	err = r.Set("first", &Key{id1})
	if err != nil {

		t.Error("Set:", "first", &Key{id1}, err)
	}

	var getKey Key
	err = r.Get("first", &getKey)
	if err != nil {

		t.Error("Get:", "first", err)
	}

	t.Log(getKey)
}

func TestRedis_Del(t *testing.T) {

	id1 := 100

	r, err := UseRedis("main", uint64(id1))
	if err != nil {

		t.Error("UseRedis:", "main", id1, err)
	}

	err = r.Del("first")
	if err != nil {

		t.Error("Del:", "first", err)
	}

	err = r.Del("second")
	if err != nil {

		t.Error("Del:", "second", err)
	}
}

func TestRedis_GetSetExpire(t *testing.T) {

	id1 := 1281

	r, err := UseRedis("main", uint64(id1))
	if err != nil {

		t.Error("UseRedis:", "main", id1, err)
	}

	err = r.Set("second", &Key{id1}, 2)
	if err != nil {

		t.Error("Set:", "second", &Key{id1}, err)
	}

	var getKey1 Key
	err = r.Get("second", &getKey1)
	if err != nil {

		t.Error("Get:", "second", err)
	}

	t.Log(getKey1)

	time.Sleep(time.Second * time.Duration(2))

	var getKey2 Key
	err = r.Get("second", &getKey2)
	if err != nil {

		t.Error("Get:", "second", err)
	}

	t.Log(getKey2)
}

func TestRedis_Incrby(t *testing.T) {

	r, err := UseRedisByName("main")
	if err != nil {

		t.Error("UseRedis:", "main", err)
	}

	err = r.Set("num", 0)
	if err != nil {

		t.Error("Set:", "num", 0, err)
	}

	result, err := r.Incrby("num", 1)
	if err != nil {

		t.Error("Incrby:", "num", 1, err)
	}
	if result != 1 {

		t.Error("Incrby:", "num", 1)
	}

	result, err = r.Incrby("num", 5)
	if err != nil {

		t.Error("Incrby:", "num", 5, err)
	}
	if result != 6 {

		t.Error("Incrby:", "num", 6)
	}

	err = r.Del("num1")
	if err != nil {

		t.Error("Del:", "num1", err)
	}

	result, err = r.Incrby("num1", 5)
	if err != nil {

		t.Error("Incrby:", "num1", 5, err)
	}
	if result != 5 {

		t.Error("Incrby:", "num1", 5, result)
	}
}

type Key struct {
	Id int
}

type User struct {
	Uid  uint64
	Name string
	Cash int64
	Gold int64
}

func (user *User) GetFieldPtr() []interface{} {

	var ptr = []interface{}{
		&user.Uid,
		&user.Name,
		&user.Cash,
		&user.Gold,
	}

	return ptr
}

func GetEntityFields(v interface{}) []interface{} {

	rv := reflect.ValueOf(v)
	erv := rv.Elem()
	rt := erv.Type()
	var fields []interface{}
	for i := 0; i < rt.NumField(); i++ {

		fields = append(fields, rt.Field(i).Name)
	}

	return fields
}

func TestRedis_Hmset_Hmget_Hset_Hget_Hincrby(t *testing.T) {

	var uid uint64 = 1001
	key := "user1"

	r, err := UseRedis("main", uid)
	if err != nil {

		t.Error("UseRedis:", "main", uid, err)
	}

	data := make(map[string]interface{})
	data["Uid"] = uid
	data["Name"] = "userName_Haha"
	data["Cash"] = 100
	data["Gold"] = 10000000
	err = r.Hmset(key, data)
	if err != nil {

		t.Error("Hmset:", "main", uid, err)
	}

	err = r.Hset(key, "Name", "Haha")
	if err != nil {

		t.Error()
	}

	var old int
	err = r.Hget(key, "Cash", &old)
	if err != nil {

		t.Error("Hget", "Cash", err, old)
	}

	num, err := r.Hincrby(key, "Cash", 1000)
	if err != nil {

		t.Error("Hincrby", "Cash", 1000, err, num)
	}
	t.Log("Cash", old, num)

	obj := new(User)
	err = r.Hmget(key, GetEntityFields(obj), obj.GetFieldPtr()...)
	if err != nil {

		t.Errorf("Hmget:", "main", uid, err)
	}

	if obj.Uid != uid {

		t.Errorf("obj.uid:", obj.Uid, uid)
	}
	t.Log(obj)
}
