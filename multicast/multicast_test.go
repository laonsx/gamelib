package multicast

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMulticast(t *testing.T) {

	handle := Handle{}

	multicas := NewMulticastService(&handle)

	multicas.NewChannel("test1", 1)
	multicas.NewChannel("test2", 2)
	multicas.NewChannel("test3", 8)

	for i := 100; i < 200; i++ {
		multicas.Subscribe("test1", uint64(i))
	}

	for i := 120; i < 150; i++ {
		multicas.Subscribe("test2", uint64(i))
	}

	for i := 140; i < 200; i++ {
		multicas.Subscribe("test3", uint64(i))
	}

	time.Sleep(1)

	for k, v := range handle {

		t.Log(k, "=>", v.Data)
	}

	for k, v := range multicas.channels {

		for s, w := range v.chans {

			for v, _ := range w.subs {

				t.Log(k, "==>", s, "==>", v)

			}
		}
	}

	multicas.Publish("test1", "coming test1")
	multicas.Publish("test2", "coming test2")
	multicas.Publish("test3", "coming test3")

	time.Sleep(10)

	go func() {

		for i := 20000; true; i++ {

			switch i % 3 {

			case 0:

				multicas.Subscribe("test1", uint64(i))

			case 1:

				multicas.Subscribe("test2", uint64(i))

			case 2:

				multicas.Subscribe("test3", uint64(i))
			}

			time.Sleep(time.Microsecond * 800)
		}
	}()

	go func() {

		for i := 200; true; i++ {

			switch i % 3 {

			case 0:

				multicas.Publish("test1", "coming test1")

			case 1:

				multicas.Publish("test2", "coming test2")

			case 2:

				multicas.Publish("test3", "coming test3")
			}

			time.Sleep(time.Microsecond * 1800)
		}
	}()

	time.Sleep(10 * time.Second)
}

type Handle map[uint64]UserInfo

func (h *Handle) Subscribe(chanId string, uid uint64) func(interface{}) {

	handle := *h
	chans, ok := handle[uid]
	if !ok {

		chans = UserInfo{
			Data: make(map[string]bool),
		}
	}

	chans.Data[chanId] = true

	handle[uid] = chans

	return func(i interface{}) {

		fmt.Println(chanId, "==>", uid, "==send==>", i)
	}
}
func (h *Handle) UnSubscribe(chanId string, uid uint64) error {

	handle := *h
	chans, ok := handle[uid]
	if !ok {

		return errors.New("not have chans")
	}

	delete(chans.Data, chanId)

	handle[uid] = chans

	return nil
}

type UserInfo struct {
	sync.RWMutex
	Data map[string]bool
}
