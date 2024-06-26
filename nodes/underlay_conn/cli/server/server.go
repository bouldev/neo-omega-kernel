package main

import (
	"fmt"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/nodes/underlay_conn"
)

func Server() {
	fmt.Println("server start")
	server, err := underlay_conn.NewServerFromBasicNet("tcp://0.0.0.0:7333")
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

func main() {
	Server()
}
