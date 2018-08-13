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
	defer mc.Unlock()

	mc.channels[id] = suChan
}

//DelChannel 删除频道
func (mc *MulticastService) DelChannel(id string) {

	suChan, _ := mc.getsuperChannel(id)
	if suChan == nil {

		return
	}

	suChan.allUnsubscribe()

	mc.Lock()
	delete(mc.channels, id)
	mc.Unlock()
}

// Subscribe 订阅
func (mc *MulticastService) Subscribe(chanId string, uid uint64) error {

	sc, err := mc.getsuperChannel(chanId)
	if err != nil {

		return err
	}

	sub := mc.handle.Subscribe(chanId, uid)
	if err != nil {

		return err
	}

	sc.subscribe(uid, sub)

	//if agent := agentService.getAgent(data.Uid); agent != nil {
	//	if !agent.isOnline() {
	//		return nil
	//	}
	//
	//	if id := agent.getSubId(data.ChanId); id > 0 {
	//		return nil
	//	}
	//
	//	mc.mux.Lock()
	//	mc.id++
	//	mc.mux.Unlock()
	//
	//	sub := new(subscriber)
	//	sub.id = mc.id
	//	sub.handle = func(msg *MulticastMsg) {
	//		agent.output(msg.PNum, msg.Msg)
	//	}
	//
	//	if agent.addSub(data.ChanId, sub.id) {
	//		c.subscribe(sub.id, sub)
	//		log.Infof("multicast sub, id=%s uid=%d", data.ChanId, data.Uid)
	//	}
	//}

	return nil
}

// UnSubscribe 退订
func (mc *MulticastService) UnSubscribe(chanId string, uid uint64) error {
	sc, err := mc.getsuperChannel(chanId)
	if err != nil {
		return err
	}

	mc.handle.UnSubscribe(chanId, uid)

	sc.unsubscribe(uid)
	//if agent, ok := agentService.agents[data.Uid]; ok {
	//	if subid := agent.getSubId(data.ChanId); subid > 0 {
	//		c.unsubscribe(subid)
	//		agent.delSub(data.ChanId)
	//	}
	//
	//	log.Infof("multicast unsub, id=%s uid=%d", data.ChanId, data.Uid)
	//}

	return nil
}

//UnSubscribeAll 退订玩家所有频道
func (mc *MulticastService) UnSubscribeAll(uid uint64, subs map[string]bool) error {

	for chanid := range subs {

		sc, err := mc.getsuperChannel(chanid)
		if err != nil {

			return err
		}

		sc.unsubscribe(uid)
	}

	return nil
}

// Publish 发布消息
func (mc *MulticastService) Publish(chanId string, msg interface{}) error {

	sc, err := mc.getsuperChannel(chanId)
	if err != nil {

		return err
	}

	sc.publish(msg)

	return nil
}

func (mc *MulticastService) getsuperChannel(id string) (*superChannel, error) {

	mc.RLock()
	defer mc.RUnlock()

	v, ok := mc.channels[id]
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

func (sc *superChannel) unsubscribe(uid uint64) {

	channel := sc.getChannel(uid)
	channel.unsubscribe(uid)
}

func (sc *superChannel) allUnsubscribe() {

	for _, v := range sc.chans {

		go v.allUnsubscribe()
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
	defer c.Unlock()

	c.subs[uid] = s
}

func (c *channel) unsubscribe(uid uint64) {

	c.Lock()
	defer c.Unlock()

	delete(c.subs, uid)
}

func (c *channel) allUnsubscribe() {

	c.Lock()
	defer c.Unlock()

	for k := range c.subs {

		delete(c.subs, k)
	}

}

func (c *channel) publish(msg interface{}) {

	c.RLock()
	defer c.RUnlock()

	for _, v := range c.subs {

		v(msg)
	}
}

type MulticastHandle interface {
	Subscribe(chanId string, uid uint64) func(interface{})
	UnSubscribe(chanId string, uid uint64) error
}
