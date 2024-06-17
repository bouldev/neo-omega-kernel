package base_net

import (
	"context"
	"net"

	"github.com/sandertv/go-raknet"
)

// RakNet is an implementation of a RakNet v10 Network.
type rakNet struct{}

// DialContext ...
func (r rakNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return raknet.DialContext(ctx, address)
}

// PingContext ...
func (r rakNet) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return raknet.PingContext(ctx, address)
}

// Listen ...
func (r rakNet) Listen(address string) (NetworkListener, error) {
	return raknet.Listen(address)
}

var RakNet Network

func init() {
	RakNet = rakNet{}
}
