package multicast

import (
	"errors"
	"sync"
)

func NewMulticastService(handle MulticastHandle) *MulticastService {

	return &MulticastService{
		channels: make(map[string]*superChannel),
		handle:   handle,
	}
}

//MulticastService 组播
type MulticastService struct {
	sync.RWMutex
	id       int
	handle   MulticastHandle
	channels map[string]*superChannel
}

//NewChannel 创建一个频道
func (mc *MulticastService) NewChannel(id string, size int) {

	if size < 0 {

		size = 1
	}

	mc.RLock()
	_, ok := mc.channels[id]
	mc.RUnlock()

	if ok {

		return
	}

	channels := make([]*channel, size)
	for i := 0; i < size; i++ {

		channels[i] = &channel{subs: make(map[uint64]func(interface{}))}
	}

	suChan := &superChannel{size: size, chans: channels}

	mc.Lock()
	mc.channels[id] = suChan
	mc.Unlock()
}

//DelChannel 删除频道
func (mc *MulticastService) DelChannel(id string) {

	suChan, _ := mc.getSuperChannel(id)
	if suChan == nil {

		return
	}

	suChan.allUnSubscribe()

	mc.Lock()
	delete(mc.channels, id)
	mc.Unlock()
}

// Subscribe 订阅
func (mc *MulticastService) Subscribe(chanId string, uid uint64) error {

	sc, err := mc.getSuperChannel(chanId)
	if err != nil {

		return err
	}

	sub := mc.handle.Subscribe(chanId, uid)
	sc.subscribe(uid, sub)

	return nil
}

// UnSubscribe 退订
func (mc *MulticastService) UnSubscribe(chanId string, uid uint64) error {

	sc, err := mc.getSuperChannel(chanId)
	if err != nil {

		return err
	}

	mc.handle.UnSubscribe(chanId, uid)
	sc.unSubscribe(uid)

	return nil
}

//UnSubscribeAll 退订玩家所有频道
func (mc *MulticastService) UnSubscribeAll(uid uint64, subs map[string]bool) error {

	for chanid := range subs {

		sc, err := mc.getSuperChannel(chanid)
		if err != nil {

			return err
		}

		sc.unSubscribe(uid)
	}

	return nil
}

// Publish 发布消息
func (mc *MulticastService) Publish(chanId string, msg interface{}) error {

	sc, err := mc.getSuperChannel(chanId)
	if err != nil {

		return err
	}

	sc.publish(msg)

	return nil
}

func (mc *MulticastService) getSuperChannel(id string) (*superChannel, error) {

	mc.RLock()
	v, ok := mc.channels[id]
	mc.RUnlock()

	if !ok {

		return nil, errors.New("superChannel does not exist")
	}

	return v, nil
}

type superChannel struct {
	size  int
	chans []*channel
}

func (sc *superChannel) getChannel(uid uint64) *channel {

	return sc.chans[uid%uint64(sc.size)]
}

func (sc *superChannel) subscribe(uid uint64, s func(interface{})) {

	channel := sc.getChannel(uid)
	channel.subscribe(uid, s)
}

func (sc *superChannel) unSubscribe(uid uint64) {

	channel := sc.getChannel(uid)
	channel.unSubscribe(uid)
}

func (sc *superChannel) allUnSubscribe() {

	for _, v := range sc.chans {

		go v.allUnSubscribe()
	}
}

func (sc *superChannel) publish(msg interface{}) {

	for _, v := range sc.chans {

		//go v.publish(msg)
		v.publish(msg)
	}
}

type channel struct {
	sync.RWMutex
	subs map[uint64]func(interface{})
}

func (c *channel) subscribe(uid uint64, s func(interface{})) {

	c.Lock()

	c.subs[uid] = s

	c.Unlock()
}

func (c *channel) unSubscribe(uid uint64) {

	c.Lock()

	delete(c.subs, uid)

	c.Unlock()
}

func (c *channel) allUnSubscribe() {

	c.Lock()

	for k := range c.subs {

		delete(c.subs, k)
	}

	c.Unlock()
}

func (c *channel) publish(msg interface{}) {

	c.RLock()

	fs := make([]func(interface{}), 0, len(c.subs))
	for _, v := range c.subs {

		fs = append(fs, v)
	}

	c.RUnlock()

	for _, f := range fs {

		f(msg)
	}
}

type MulticastHandle interface {
	Subscribe(chanId string, uid uint64) func(interface{})
	UnSubscribe(chanId string, uid uint64) error
}
