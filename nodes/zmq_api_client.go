package nodes

import (
	"context"
	"fmt"
	"neo-omega-kernel/utils/sync_wrapper"
	"strings"
	"time"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/uuid"
)

type SimpleZMQAPIClient struct {
	zmq.Socket
	cbs  *sync_wrapper.SyncKVMap[string, func(Values)]
	apis *sync_wrapper.SyncKVMap[string, func(Values, func(Values))]
}

func CreateZMQClientSocket(endPoint string) (zmq.Socket, error) {
	name := fmt.Sprintf("%v", uuid.New().String())
	id := zmq.SocketIdentity(name)
	socket := zmq.NewDealer(context.Background(), zmq.WithID(id), zmq.WithDialerMaxRetries(-1))
	if err := socket.Dial(endPoint); err != nil {
		return nil, err
	}
	return socket, nil
}

func NewSimpleZMQAPIClient(socket zmq.Socket) (c ZMQAPIClient) {
	client := &SimpleZMQAPIClient{
		socket,
		sync_wrapper.NewSyncKVMap[string, func(Values)](),
		sync_wrapper.NewSyncKVMap[string, func(Values, func(Values))](),
	}
	return client
}

func (c *SimpleZMQAPIClient) Run() (err error) {
	var msg zmq.Msg
	for {
		msg, err = c.Socket.Recv()
		if err != nil {
			return err
		}
		indexOrApi := string(msg.Frames[0])
		if strings.HasPrefix(indexOrApi, "/") {
			index := msg.Frames[1]
			if apiFn, ok := c.apis.Get(indexOrApi); ok {
				apiFn(msg.Frames[2:], func(z Values) {
					if len(index) == 0 {
						return
					}
					frames := append([][]byte{index}, z...)
					if c.SendMulti(zmq.NewMsgFrom(frames...)) != nil {
						c.Socket.Close()
					}
				})
			}
		} else {
			if cb, ok := c.cbs.GetAndDelete(indexOrApi); ok {
				cb(msg.Frames[1:])
			}
		}
	}
}

func (c *SimpleZMQAPIClient) CallOmitResponse(api string, args Values) {
	frames := append([][]byte{[]byte(api), []byte{}}, args...)
	err := c.Socket.SendMulti(zmq.NewMsgFrom(frames...))
	if err != nil {
		c.Socket.Close()
	}
}

type clientRespHandler struct {
	idx    string
	frames [][]byte
	c      *SimpleZMQAPIClient
	ctx    context.Context
}

func (h *clientRespHandler) doSend() {
	err := h.c.Socket.SendMulti(zmq.NewMsgFrom(h.frames...))
	if err != nil {
		h.c.Socket.Close()
	}
}

func (h *clientRespHandler) SetContext(ctx context.Context) ZMQResultHandler {
	h.ctx = ctx
	return h
}

func (h *clientRespHandler) SetTimeout(timeout time.Duration) ZMQResultHandler {
	if h.ctx == nil {
		h.ctx = context.Background()
	}
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *clientRespHandler) BlockGetResponse() Values {
	resolver := make(chan Values, 1)
	h.c.cbs.Set(h.idx, func(ret Values) {
		resolver <- ret
	})
	h.doSend()
	if h.ctx == nil {
		return <-resolver
	}
	select {
	case ret := <-resolver:
		return ret
	case <-h.ctx.Done():
		h.c.cbs.Delete(h.idx)
		return Empty
	}
}

func (h *clientRespHandler) AsyncGetResponse(callback func(Values)) {
	if h.ctx == nil {
		h.c.cbs.Set(h.idx, callback)
	} else {
		resolver := make(chan Values, 1)
		h.c.cbs.Set(h.idx, func(ret Values) {
			resolver <- ret
		})
		go func() {
			select {
			case ret := <-resolver:
				callback(ret)
			case <-h.ctx.Done():
				h.c.cbs.Delete(h.idx)
				callback(Empty)
				return
			}
		}()
	}
	h.doSend()
}

func (c *SimpleZMQAPIClient) CallWithResponse(api string, args Values) ZMQResultHandler {
	idx := uuid.New().String()
	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
	return &clientRespHandler{
		idx, frames, c, nil,
	}
}

func (c *SimpleZMQAPIClient) ExposeAPI(apiName string, api ZMQClientAPI, newGoroutine bool) {
	c.apis.Set(apiName, func(args Values, setResult func(Values)) {
		if newGoroutine {
			go func() {
				ret := api(args)
				setResult(ret)
			}()
		} else {
			ret := api(args)
			setResult(ret)
		}
	})
}
