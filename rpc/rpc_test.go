package rpc

import (
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"
)

var nodes map[string]string = map[string]string{
	"node1": "127.0.0.1:10000",
	"node2": "127.0.0.1:10000",
}

var methods [][]string = [][]string{
	[]string{"1001", "TestRpc1.HelloWorld1"},
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

	sname, err := GetName(uint16(1001))
	t.Log(sname, err)
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

	stream, err := Stream("node1", map[string]string{SESSIONUID: "123321123321"})
	if err != nil {

		t.Error(err)
	}

	in := &GameMsg{
		ServiceName: "TestRpc1.HelloWorld1",
		Msg:         []byte("hello world"),
	}
	stream.Send(in)
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

	stream.Send(in2)

	result, err = stream.Recv()
	if err != nil {

		t.Error(err)
	}

	t.Log(result)

	stream, err = Stream("node2", nil)
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
		err = nil
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
