package nodes

import (
	"bytes"
	"context"
	"neo-omega-kernel/utils/sync_wrapper"
	"os"
	"strings"
	"time"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/uuid"
)

type SimpleZMQServer struct {
	zmq.Socket
	cbs  *sync_wrapper.SyncKVMap[string, func(Values)]
	apis *sync_wrapper.SyncKVMap[string, func(caller ZMQCaller, args Values, setResult func(Values))]
}

func CreateZMQServerSocket(endPoint string) (zmq.Socket, error) {
	if strings.HasPrefix(endPoint, "ipc://") {
		endPoint := endPoint[len("ipc://"):]
		if _, err := os.Stat(endPoint); err == nil {
			if err = os.Remove(endPoint); err != nil {
				return nil, err
			}
		}
	}
	socket := zmq.NewRouter(context.Background(), zmq.WithID(zmq.SocketIdentity("router")))
	if err := socket.Listen(endPoint); err != nil {
		return nil, err
	}
	return socket, nil
}

func NewSimpleZMQServer(socket zmq.Socket) ZMQAPIServer {
	server := &SimpleZMQServer{
		Socket: socket,
		cbs:    sync_wrapper.NewSyncKVMap[string, func(Values)](),
		apis:   sync_wrapper.NewSyncKVMap[string, func(ZMQCaller, Values, func(Values))](),
	}
	return server
}

func (s *SimpleZMQServer) Serve() (err error) {
	// prob := block_prob.NewBlockProb("ZMQ Server Handle Msg Block Prob", time.Second/10)
	for {
		var msg zmq.Msg
		msg, err = s.Socket.Recv()
		// mark := prob.MarkEventStartByTimeout(func() string {
		// 	ev := "Msg: "
		// 	for _, f := range msg.Frames {
		// 		ev += string(f) + " "
		// 	}
		// 	return ev
		// }, time.Second/5)
		if err != nil {
			return err
		}
		frames := msg.Frames
		if len(frames) < 3 {
			//socket.Send(ErrCallFormat)
			continue
		}
		caller := frames[0]
		indexOrAPI := string(frames[1])
		if strings.HasPrefix(indexOrAPI, "/") {
			caller = bytes.Clone(caller)
			index := bytes.Clone(frames[2])
			if apiFn, ok := s.apis.Get(indexOrAPI); ok {
				apiFn(caller, msg.Frames[3:], func(z Values) {
					if len(index) == 0 {
						return
					}
					frames := append([][]byte{caller, index}, z...)
					if s.SendMulti(zmq.NewMsgFrom(frames...)) != nil {
						s.Socket.Close()
					}
				})
			}
		} else {
			if cb, ok := s.cbs.GetAndDelete(indexOrAPI); ok {
				cb(msg.Frames[2:])
			}
		}
		// prob.MarkEventFinished(mark)
	}
}

func (c *SimpleZMQServer) ExposeAPI(apiName string, api ZMQServerAPI, newGoroutine bool) {
	c.apis.Set(apiName, func(caller ZMQCaller, args Values, setResult func(Values)) {
		if newGoroutine {
			go func() {
				ret := api(caller, args)
				setResult(ret)
			}()
		} else {
			ret := api(caller, args)
			setResult(ret)
		}
	})
}

func (c *SimpleZMQServer) ConcealAPI(apiName string) {
	c.apis.Delete(apiName)
}

func (c *SimpleZMQServer) CallOmitResponse(callee ZMQCaller, api string, args Values) {
	frames := append([][]byte{callee, []byte(api), []byte{}}, args...)
	err := c.Socket.SendMulti(zmq.NewMsgFrom(frames...))
	if err != nil {
		c.Socket.Close()
	}
}

type serverRespHandler struct {
	idx    string
	frames [][]byte
	c      *SimpleZMQServer
	ctx    context.Context
}

func (h *serverRespHandler) doSend() {
	err := h.c.Socket.SendMulti(zmq.NewMsgFrom(h.frames...))
	if err != nil {
		h.c.Socket.Close()
	}
}

func (h *serverRespHandler) SetContext(ctx context.Context) ZMQResultHandler {
	h.ctx = ctx
	return h
}

func (h *serverRespHandler) SetTimeout(timeout time.Duration) ZMQResultHandler {
	if h.ctx == nil {
		h.ctx = context.Background()
	}
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *serverRespHandler) BlockGetResponse() Values {
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

func (h *serverRespHandler) AsyncGetResponse(callback func(Values)) {
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

func (c *SimpleZMQServer) CallWithResponse(callee ZMQCaller, api string, args Values) ZMQResultHandler {
	idx := uuid.New().String()
	frames := append([][]byte{callee, []byte(api), []byte(idx)}, args...)
	return &serverRespHandler{
		idx, frames, c, nil,
	}
}
