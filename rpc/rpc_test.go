package rpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laonsx/gamelib/graceful"
	"github.com/laonsx/gamelib/zookeeper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

var nodes = map[string]string{
	"node1": "127.0.0.1:10000",
	"node2": "127.0.0.1:10000",
}

var methods = []*ServiceConf{
	{1001, "TestRpc1.HelloWorld1", "node1"},
}

func init() {

	RegisterService(&TestRpc1{})
	RegisterService(&TestRpc2{})
}

type TestRpc1 struct {
}

func (testRpc *TestRpc1) HelloWorld1(data []byte, session *Session) []byte {

	fmt.Println("request data:", string(data), ";session:", session)

	return []byte("return from testrpc1 helloworld1.")
}

type TestRpc2 struct {
}

func (testRpc *TestRpc2) HelloWorld2(data []byte, session *Session) []byte {

	fmt.Println("request data......:", string(data), ";session:", session)

	return []byte("return from testrpc2 helloworld2......")

}

func TestRpc_GetName(t *testing.T) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	InitClient(nodes, methods, opts)

	node, sname, err := GetName(uint16(1001))

	t.Log(node, sname, err)
}

func TestRpcWithZk(t *testing.T) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name), grpc.WithInsecure(), grpc.WithBlock())
	InitClient(map[string]string{"node1": zookeeper.InitGrpcDialUrl("localhost:2181", "node1")}, methods, opts)

	var serverOpts []grpc.ServerOption
	lis, err := net.Listen("tcp", "127.0.0.1:10000")
	if err != nil {
		panic(err)
	}
	rpcServer := NewServer("node1", lis, serverOpts)
	go rpcServer.Start()

	stream, _, err := Stream("node1", map[string]string{SESSIONUID: "123321123321"})
	if err != nil {

		t.Error(err)
	}

	in := &GameMsg{
		ServiceName: "TestRpc1.HelloWorld1",
		Msg:         []byte("hello world"),
	}
	_ = stream.Send(in)
	result, err := stream.Recv()
	if err != nil {

		t.Error(err)
	}

	t.Log(result)

	resp, err := Call("node1", "TestRpc2.HelloWorld2", []byte("ahahahhahaha"), nil)
	if err != nil {

		t.Error(err)
	}

	t.Log(string(resp))

	in2 := &GameMsg{
		ServiceName: "TestRpc2.HelloWorld2",
		Msg:         []byte("hello rpc2"),
	}

	_ = stream.Send(in2)

	result, err = stream.Recv()
	if err != nil {

		t.Error(err)
	}

	t.Log(result)

	zookeeper.UnRegister()
}

func TestRpc(t *testing.T) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	InitClient(nodes, methods, opts)

	var serverOpts []grpc.ServerOption
	lis, err := net.Listen("tcp", "127.0.0.1:10000")
	if err != nil {
		panic(err)
	}
	rpcServer := NewServer("node1", lis, serverOpts)
	go rpcServer.Start()

	stream, _, err := Stream("node1", map[string]string{SESSIONUID: "123321123321"})
	if err != nil {

		t.Error(err)
	}

	in := &GameMsg{
		ServiceName: "TestRpc1.HelloWorld1",
		Msg:         []byte("hello world"),
	}
	_ = stream.Send(in)
	result, err := stream.Recv()
	if err != nil {

		t.Error(err)
	}

	t.Log(result)

	resp, err := Call("node1", "TestRpc2.HelloWorld2", []byte("ahahahhahaha"), nil)
	if err != nil {

		t.Error(err)
	}

	t.Log(string(resp))

	in2 := &GameMsg{
		ServiceName: "TestRpc2.HelloWorld2",
		Msg:         []byte("hello rpc2"),
	}

	_ = stream.Send(in2)

	result, err = stream.Recv()
	if err != nil {

		t.Error(err)
	}

	t.Log(result)

	stream, _, err = Stream("node2", nil)
	if err != nil {

		t.Error(err)
	}

	in = &GameMsg{
		ServiceName: "TestRpc1.HelloWorld1",
		Msg:         []byte("hello world,no session"),
	}
	err = stream.Send(in)
	if err != nil {

		t.Error(".------------.", err)
	}

	result, err = stream.Recv()
	if err != nil {

		t.Error("..............", err)
		//err = nil
	}

	t.Log(result)

	//ins := &GameMsg{
	//	ServiceName: "TestRpc1.HelloWorld1",
	//	Msg:         []byte("hello world,no session,but set session"),
	//	Session:     &Session{Uid: uint64(1234567890)},
	//}
	//err = stream.Send(ins)
	//if err != nil {
	//
	//	if err == io.EOF {
	//
	//		stream.CloseSend()
	//		stream, err = Stream("node2", nil)
	//		if err != nil {
	//
	//			t.Error("reset..rpc..stream", err)
	//		}
	//
	//		t.Log("resend err=", stream.Send(ins))
	//	}
	//	t.Error(".------------.", err)
	//}
	//
	//result, errs := stream.Recv()
	//if errs != nil {
	//
	//	t.Error(err)
	//}
	//
	//t.Log(result)
	//
	//result, err = stream.Recv()
	//if err != nil {
	//
	//	t.Error(err)
	//}
	//
	//t.Log(result)
	//
	//resp, err = Call("node1", "TestRpc2.HelloWorld2", []byte("ahahahhahaha"), &Session{Uid: uint64(10000001)})
	//if err != nil {
	//
	//	t.Error(err)
	//}
	//
	//t.Log(string(resp))
}

func TestServer_GatewayHandler(t *testing.T) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	InitClient(nodes, methods, opts)

	var serverOpts []grpc.ServerOption
	lis, err := net.Listen("tcp", "127.0.0.1:10000")
	if err != nil {

		panic(err)
	}

	rpcServer := NewServer("node1", lis, serverOpts)

	go rpcServer.Start()

	router := gin.New()
	rpcServer.GatewayHandler(router, DefaultSessionFunc)

	go graceful.ListenAndServe(":10001", router)

	time.Sleep(time.Second)

	resp, err := Call("node1", "TestRpc1.HelloWorld1", []byte("ahahahhahaha"), nil)
	if err != nil {

		t.Error(err)
	}

	t.Log(string(resp))

	req, _ := http.NewRequest("POST", "http://127.0.0.1:10001/TestRpc1/HelloWorld1", bytes.NewBufferString("httprequesthahahaha"))
	req.Header.Set("token", "123456")

	client := &http.Client{}

	resps, _ := client.Do(req)
	b, _ := ioutil.ReadAll(resps.Body)
	t.Log(string(b))

	req, _ = http.NewRequest("POST", "http://127.0.0.1:10001/TestRpc2/HelloWorld2", bytes.NewBufferString("httprequesthahahaha"))
	req.Header.Set("token", "654321")

	resps, _ = client.Do(req)
	b, _ = ioutil.ReadAll(resps.Body)
	t.Log(string(b))

	time.Sleep(time.Second)
}

func TestUnmarshal(t *testing.T) {

	type Test struct {
		Haha string
	}

	teststruct := &Test{Haha: "haha"}

	data, err := Marshal(CodecType_Json, teststruct)
	fmt.Println(string(data), err)

	teststruct2 := &Test{}

	err = Unmarshal(CodecType_Json, data, teststruct2)
	fmt.Println(teststruct2, err)

	s := &Session{
		Uid:   1111,
		Codec: CodecType_ProtoBuf,
	}

	data1, err1 := Marshal(CodecType_ProtoBuf, s)
	fmt.Println(err1)

	ss := &Session{}
	err1 = Unmarshal(CodecType_ProtoBuf, data1, ss)
	fmt.Println(ss, err1)
}
