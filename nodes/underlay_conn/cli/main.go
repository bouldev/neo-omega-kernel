package main

import (
	"fmt"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/nodes/underlay_conn"
	"time"
)

func Server() {
	fmt.Println("server start")
	server, err := underlay_conn.NewServerFromBasicNet("tcp://0.0.0.0:7241")
	if err != nil {
		panic(err)
	}
	server.ExposeAPI("echo", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		fmt.Println("server recv:", caller, args.ToStrings())
		go func() {
			ret := server.CallWithResponse(caller, "client-echo", defines.FromStrings("server", "hi")).BlockGetResponse()
			fmt.Println("server get response ", ret.ToStrings())
			server.CallOmitResponse(caller, "client-echo", defines.FromStrings("server", "hi", "no resp"))
		}()
		return defines.FromString("server echo").Extend(args)
	}, false)
	<-server.WaitClosed()
	fmt.Println("server closed")
}

func Client(id string) {
	fmt.Println("client start")
	client, err := underlay_conn.NewClientFromBasicNet("tcp://127.0.0.1:7241", time.Second)
	if err != nil {
		panic(err)
	}
	go func() {
		ret := client.CallWithResponse("echo", defines.FromStrings("hello", "world")).BlockGetResponse()
		fmt.Println("client get response ", ret.ToStrings())
		client.CallOmitResponse("echo", defines.FromStrings("hello", "world", "no resp"))
	}()
	client.ExposeAPI("client-echo", func(args defines.Values) defines.Values {
		fmt.Println(fmt.Sprintf("client %v recv:", id), args.ToStrings())
		return defines.FromString(fmt.Sprintf("client %v echo", id)).Extend(args)
	}, false)
	<-client.WaitClosed()
	fmt.Printf("client %v closed\n", id)
}

func main() {
	go Server()
	time.Sleep(time.Second)
	go Client("1")
	go Client("2")
	c := make(chan struct{})
	<-c
}
