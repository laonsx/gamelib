package rpc

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gamelib/codec"
	"gamelib/gofunc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type methodType struct {
	sync.Mutex
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
	numCalls  uint
}

type service struct {
	name   string
	rcvr   reflect.Value
	typ    reflect.Type
	method map[string]*methodType
}

//处理客户端发送的数据包
func (s *service) handle(methodName string, in *GameMsg, session *Session) (*GameMsg, error) {

	defer gofunc.PrintPanic()

	method, ok := s.method[methodName]
	if !ok {

		return nil, errors.New("method not found")
	}

	if session == nil {

		return s.call(methodName, in)
	}

	function := method.method.Func
	rvs := []reflect.Value{s.rcvr, reflect.ValueOf(session), reflect.ValueOf(in.Msg)}
	ret := function.Call(rvs)
	resp := ret[0].Bytes()

	gameMsg := &GameMsg{ServiceName: in.ServiceName, Msg: resp}

	return gameMsg, nil
}

func (s *service) call(methodName string, in *GameMsg) (*GameMsg, error) {

	defer gofunc.PrintPanic()

	method, ok := s.method[methodName]
	if !ok {

		return nil, errors.New("method not found")
	}

	function := method.method.Func
	var rvs []reflect.Value
	argIsValue := false
	var argv reflect.Value

	if method.argType.Kind() == reflect.Ptr {

		argv = reflect.New(method.argType.Elem())
	} else {

		argv = reflect.New(method.argType)
		argIsValue = true
	}

	err := codec.UnMsgPack(in.Msg, argv.Interface())
	if err != nil {

		return nil, err
	}

	if argIsValue {

		argv = argv.Elem()
	}

	var replyv reflect.Value
	if method.replyType.Kind() == reflect.Ptr {

		replyv = reflect.New(method.replyType.Elem())
	} else {

		replyv = reflect.New(method.replyType).Elem()
	}

	rvs = []reflect.Value{s.rcvr, argv, replyv}
	retV := function.Call(rvs)
	errInter := retV[0].Interface()
	errmsg := ""
	var b []byte
	if errInter != nil {

		errmsg = errInter.(error).Error()
	} else {

		b, err = codec.MsgPack(replyv.Interface())
		if err != nil {

			return nil, err
		}
	}

	gameMsg := &GameMsg{ServiceName: in.ServiceName, Msg: b, Error: errmsg}

	return gameMsg, nil
}

var server *Server

func init() {

	server = new(Server)
	server.serviceMap = make(map[string]*service)
	server.stream = make(map[string]Game_StreamServer)
}

// Session 存储用户信息
type Session struct {
	UserId uint64
}

// Server 对象
type Server struct {
	Name string

	listener   net.Listener
	opts       []grpc.ServerOption
	mux        sync.RWMutex
	serviceMap map[string]*service
	grpcServer *grpc.Server
	stream     map[string]Game_StreamServer
}

// NewServer 创建Server对象
func NewServer(name string, lis net.Listener, opts []grpc.ServerOption) *Server {

	server.Name = name
	server.listener = lis
	server.opts = opts
	return server
}

// Start 启动rpc服务
func (s *Server) Start() {

	grpcServer := grpc.NewServer(s.opts...)
	s.grpcServer = grpcServer

	RegisterGameServer(grpcServer, s)

	log.Printf("rpcserver listening on %s", s.listener.Addr().String())
	grpcServer.Serve(s.listener)
}

// Stop 停止rpc服务
func (s *Server) Close() {

	log.Printf("rpcserver closing")
	s.grpcServer.Stop()
}

// Call grpc server接口实现
func (s *Server) Call(ctx context.Context, in *GameMsg) (resp *GameMsg, err error) {

	dot := strings.LastIndex(in.ServiceName, ".")
	sname := in.ServiceName[:dot]
	mname := in.ServiceName[dot+1:]
	serv, ok := s.serviceMap[sname]
	if !ok {

		return nil, errors.New("service not found")
	}

	resp, err = serv.call(mname, in)

	return
}

// Stream grpc server接口实现
func (s *Server) Stream(stream Game_StreamServer) error {

	gameMsg := make(chan *GameMsg, 1)
	quit := make(chan int)

	go func() {

		defer close(quit)

		for {

			in, err := stream.Recv()
			if err == io.EOF {

				return
			}

			if err != nil {

				return
			}

			gameMsg <- in
		}
	}()

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {

		log.Println("rpcserver stream ctx err")

		return errors.New("stream ctx error")
	}

	name := md["name"][0]

	var session *Session
	if name == "agent" {

		userID, _ := strconv.ParseUint(md["uid"][0], 10, 64)
		session = new(Session)
		session.UserId = userID
	} else {

		s.stream[name] = stream
	}

	defer func() {

		if name != "agent" {

			delete(s.stream, name)
		}

		close(gameMsg)
	}()

	for {

		select {

		case in := <-gameMsg:

			dot := strings.LastIndex(in.ServiceName, ".")
			sname := in.ServiceName[:dot]
			mname := in.ServiceName[dot+1:]
			serv, ok := s.serviceMap[sname]
			if !ok {

				return errors.New("service not found")
			}

			resp, err := serv.handle(mname, in, session)
			if err != nil {

				log.Printf("rpcserver handle %v", err)
				return err
			}

			if err := stream.Send(resp); err != nil {

				log.Printf("rpcserver streamsend, err=%s", err.Error())

				return err
			}

		case <-quit:

			return nil
		}
	}
}

// RegisterService 注册一个服务
func RegisterService(v interface{}) error {

	server.mux.Lock()
	defer server.mux.Unlock()

	if server.serviceMap == nil {

		server.serviceMap = make(map[string]*service)
	}

	s := new(service)
	s.typ = reflect.TypeOf(v)
	s.rcvr = reflect.ValueOf(v)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {

		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Println(s)

		return errors.New(s)
	}

	if _, present := server.serviceMap[sname]; present {

		return errors.New("rpc: service already defined: " + sname)
	}

	s.name = sname
	s.method = suitableMethods(s.typ, true)
	server.serviceMap[s.name] = s

	return nil
}

func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {

	methods := make(map[string]*methodType)

	for m := 0; m < typ.NumMethod(); m++ {

		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		if method.PkgPath != "" {

			continue
		}

		argType := mtype.In(1)
		replyType := mtype.In(2)

		if mtype.NumOut() != 1 {

			if reportErr {

				log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			}

			continue
		}

		methods[mname] = &methodType{method: method, argType: argType, replyType: replyType}
	}

	return methods
}
