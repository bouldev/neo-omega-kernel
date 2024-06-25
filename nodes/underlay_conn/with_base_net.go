package underlay_conn

import (
	"fmt"
	"neo-omega-kernel/minecraft_neo/cascade_conn/byte_frame_conn"
	"neo-omega-kernel/nodes/defines"
	"net"
	"strings"
	"time"
)

func NewBasicNetServer(addr string) (net.Listener, error) {
	frags := strings.Split(addr, "://")
	if len(frags) != 2 {
		return nil, fmt.Errorf("must be in format of network://address. e.g. tcp://0.0.0.0:2401")
	}
	return net.Listen(frags[0], frags[1])
}

func NewBasicNetClient(addr string, timeout time.Duration) (net.Conn, error) {
	frags := strings.Split(addr, "://")
	if len(frags) != 2 {
		return nil, fmt.Errorf("must be in format of network://address. e.g. tcp://127.0.0.1:2401")
	}
	return net.DialTimeout(frags[0], frags[1], timeout)
}

func NewClientFromBasicNet(addr string, timeout time.Duration) (defines.ZMQAPIClient, error) {
	conn, err := NewBasicNetClient(addr, timeout)
	if err != nil {
		return nil, err
	}
	frameConn := byte_frame_conn.NewConnectionFromNet(conn)
	client := NewFrameAPIClient(frameConn)
	go client.Run()
	return client, nil
}

func NewServerFromBasicNet(addr string) (defines.ZMQAPIServer, error) {
	listen, err := NewBasicNetServer(addr)
	if err != nil {
		return nil, err
	}
	server := NewFrameAPIServer(func() { listen.Close() })
	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Println("Accept() failed, err: ", err)
				continue
			}
			frameConn := byte_frame_conn.NewConnectionFromNet(conn)
			serveConn := server.NewFrameAPIServer(frameConn)
			go serveConn.Run()
		}
	}()
	return server, nil
}