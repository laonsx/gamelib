package rpc

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	callTimeout = 5 * time.Second
	client      *Client

	mux                sync.RWMutex
	streamClientCaches = make(map[string]*StreamClientCache)
)

// InitClient 初始化客户端
func InitClient(cluster map[string]string, services []*ServiceConf, opts []grpc.DialOption) {

	client = new(Client)
	client.clients = make(map[string]GameClient)
	client.cluster = cluster

	serviceMap := make(map[string]*ServiceConf)
	servicesNumMap := make(map[uint16]*ServiceConf)

	for _, v := range services {

		serviceMap[v.Sname] = v
		servicesNumMap[v.Pnum] = v
	}
	client.servicesMap = serviceMap
	client.servicesNumMap = servicesNumMap
	client.opts = opts
}

func ReloadMethodConf(services []*ServiceConf) {

	client.mux.Lock()
	defer client.mux.Unlock()

	serviceMap := make(map[string]*ServiceConf)
	servicesNumMap := make(map[uint16]*ServiceConf)

	for _, v := range services {

		serviceMap[v.Sname] = v
		servicesNumMap[v.Pnum] = v
	}

	client.servicesMap = serviceMap
	client.servicesNumMap = servicesNumMap
}

// GetName 根据协议号 获取节点名称和服务名
func GetName(pnum uint16) (node, sname string, err error) {

	if s, ok := client.servicesNumMap[pnum]; ok {

		node = s.Node
		sname = s.Sname
	} else {

		err = fmt.Errorf("service not found by pnum(%d)", pnum)
	}

	return
}

// GetPNum 根据服务名 获取协议号
func GetPNum(service string) (node string, pnum uint16, err error) {

	if s, ok := client.servicesMap[service]; ok {

		node = s.Node
		pnum = s.Pnum
	} else {

		err = fmt.Errorf("pnum not found by service(%s)", service)
	}

	return
}

// Stream 获取一个流
func Stream(node string, md map[string]string) (Game_StreamClient, context.CancelFunc, error) {

	var c GameClient

	c, err := client.newClient(node)
	if err != nil {

		return nil, nil, err
	}

	//var ctx = context.Background()
	ctx, cancel := context.WithCancel(context.Background())
	if md != nil {

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))
	}

	stream, err := c.Stream(ctx)

	return stream, cancel, err
}

func StreamCall(node string, service string, data []byte, session *Session) ([]byte, error) {

	mux.RLock()
	streamCache, ok := streamClientCaches[node]
	mux.RUnlock()

	var err error

	if !ok || streamCache.stream == nil {

		streamCache = &StreamClientCache{}
		streamCache.stream, streamCache.cancel, err = Stream(node, nil)
		if err != nil {

			return nil, err
		}

		mux.Lock()
		streamClientCaches[node] = streamCache
		mux.Unlock()
	}

	err = streamCache.stream.Send(&GameMsg{ServiceName: service, Msg: data, Session: session})
	if err == io.EOF {

		_ = streamCache.stream.CloseSend()
		streamCache.cancel()
		streamCache.stream, streamCache.cancel, err = Stream(node, nil)
		if err == nil {

			err = streamCache.stream.Send(&GameMsg{ServiceName: service, Msg: data, Session: session})
		}
	}
	if err != nil {

		return nil, err
	}

	ret, err := streamCache.stream.Recv()
	if err != nil {

		return nil, err
	}

	return ret.Msg, err
}

// Call 简单的grpc调用
func Call(node string, service string, data []byte, session *Session) ([]byte, error) {

	c, err := client.newClient(node)
	if err != nil {

		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	ret, err := c.Call(ctx, &GameMsg{ServiceName: service, Msg: data, Session: session})
	if err != nil {

		return nil, err
	}

	return ret.Msg, err
}

// Client rpc Client结构
type Client struct {
	mux            sync.Mutex
	clients        map[string]GameClient
	cluster        map[string]string
	servicesMap    map[string]*ServiceConf
	servicesNumMap map[uint16]*ServiceConf
	opts           []grpc.DialOption
}

type ServiceConf struct {
	Pnum  uint16 `json:"pnum"`
	Sname string `json:"sname"`
	Node  string `json:"node"`
}

func (c *Client) newClient(node string) (GameClient, error) {

	if v, ok := c.clients[node]; ok {

		return v, nil
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if addr, ok := c.cluster[node]; ok {

		conn, err := grpc.Dial(addr, c.opts...)
		if err != nil {

			return nil, err
		}

		gameClient := NewGameClient(conn)
		c.clients[node] = gameClient

		return gameClient, nil
	}

	return nil, errors.New("node conf not found")
}

type StreamClientCache struct {
	stream Game_StreamClient
	cancel context.CancelFunc
}
