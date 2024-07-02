package underlay_conn

import (
	"context"
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	conn_defines "github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/defines"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"

	"github.com/google/uuid"
)

type FrameAPIServer struct {
	apis    *sync_wrapper.SyncKVMap[string, func(defines.ZMQCaller, defines.Values, func(defines.Values))]
	conns   *sync_wrapper.SyncKVMap[string, *FrameAPIServerConn]
	onClose *sync_wrapper.SyncKVMap[string, func()]
	can_close.CanCloseWithError
}

func NewFrameAPIServer(onCloseHook func()) *FrameAPIServer {
	return &FrameAPIServer{
		apis:              sync_wrapper.NewSyncKVMap[string, func(defines.ZMQCaller, defines.Values, func(defines.Values))](),
		conns:             sync_wrapper.NewSyncKVMap[string, *FrameAPIServerConn](),
		onClose:           sync_wrapper.NewSyncKVMap[string, func()](),
		CanCloseWithError: can_close.NewClose(onCloseHook),
	}
}

type FrameAPIServerConn struct {
	identity      string
	identityBytes []byte
	can_close.CanCloseWithError
	*FrameAPIServer
	FrameConn conn_defines.ByteFrameConnBase
	cbs       *sync_wrapper.SyncKVMap[string, func(defines.Values)]
}

func (s *FrameAPIServer) NewFrameAPIServer(conn conn_defines.ByteFrameConnBase) *FrameAPIServerConn {
	identity := uuid.New().String()
	c := &FrameAPIServerConn{
		identity:      identity,
		identityBytes: []byte(identity),
		// close underlay conn on err
		CanCloseWithError: can_close.NewClose(conn.Close),
		FrameConn:         conn,
		cbs:               sync_wrapper.NewSyncKVMap[string, func(defines.Values)](),
		FrameAPIServer:    s,
	}
	s.conns.Set(identity, c)
	go func() {
		// close when underlay err
		c.CloseWithError(<-conn.WaitClosed())
		onClose, ok := s.onClose.Get(identity)
		if ok {
			onClose()
		}
		s.conns.Delete(identity)
	}()
	return c
}

func (c *FrameAPIServer) SetOnCloseCleanUp(callee defines.ZMQCaller, cb func()) {
	c.onClose.Set(string(callee), cb)
}

func (s *FrameAPIServer) NewFrameAPIServerWithCtx(conn conn_defines.ByteFrameConn, apis *FrameAPIServer, ctx context.Context) *FrameAPIServerConn {
	c := s.NewFrameAPIServer(conn)
	go func() {
		select {
		case <-c.WaitClosed():
		case <-ctx.Done():
			c.CloseWithError(ctx.Err())
		}
	}()
	return c
}

func (c *FrameAPIServer) ConcealAPI(apiName string) {
	c.apis.Delete(apiName)
}

func (c *FrameAPIServer) ExposeAPI(apiName string, api defines.ZMQServerAPI, newGoroutine bool) {
	if !strings.HasPrefix(apiName, "/") {
		apiName = "/" + apiName
	}
	c.apis.Set(apiName, func(caller defines.ZMQCaller, args defines.Values, setResult func(defines.Values)) {
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

func (c *FrameAPIServer) CallOmitResponse(callee defines.ZMQCaller, api string, args defines.Values) {
	conn, ok := c.conns.Get(string(callee))
	if !ok {
		return
	}
	conn.CallOmitResponse(api, args)
}

type serverVoidRespHandler struct{}

func (h *serverVoidRespHandler) SetContext(ctx context.Context) defines.ZMQResultHandler   { return h }
func (h *serverVoidRespHandler) SetTimeout(timeout time.Duration) defines.ZMQResultHandler { return h }
func (h *serverVoidRespHandler) BlockGetResponse() defines.Values                          { return defines.Empty }
func (h *serverVoidRespHandler) AsyncGetResponse(callback func(defines.Values)) {
	go callback(defines.Empty)
}

func (c *FrameAPIServer) CallWithResponse(callee defines.ZMQCaller, api string, args defines.Values) defines.ZMQResultHandler {
	conn, ok := c.conns.Get(string(callee))
	if !ok {
		return &serverVoidRespHandler{}
	}
	return conn.CallWithResponse(api, args)
}

func (c *FrameAPIServerConn) Run() {
	c.FrameConn.ReadRoutine(func(data []byte) {
		frames := bytesToBytesSlices(data)
		indexOrApi := string(frames[0])
		if strings.HasPrefix(indexOrApi, "/") {
			index := frames[1]
			if apiFn, ok := c.apis.Get(indexOrApi); ok {
				apiFn(defines.ZMQCaller(c.identity), frames[2:], func(z defines.Values) {
					if len(index) == 0 {
						return
					}
					frames := append([][]byte{index}, z...)
					c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
				})
			}
		} else {
			if cb, ok := c.cbs.GetAndDelete(indexOrApi); ok {
				cb(frames[1:])
			}
		}
	})
}

func (c *FrameAPIServerConn) CallOmitResponse(api string, args defines.Values) {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	frames := append([][]byte{[]byte(api), {}}, args...)
	c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
}

type serverRespHandler struct {
	idx    string
	frames [][]byte
	c      *FrameAPIServerConn
	ctx    context.Context
}

func (h *serverRespHandler) doSend() {
	h.c.FrameConn.WriteBytePacket(byteSlicesToBytes(h.frames))
}

func (h *serverRespHandler) SetContext(ctx context.Context) defines.ZMQResultHandler {
	h.ctx = ctx
	return h
}

func (h *serverRespHandler) SetTimeout(timeout time.Duration) defines.ZMQResultHandler {
	if h.ctx == nil {
		h.ctx = context.Background()
	}
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *serverRespHandler) BlockGetResponse() defines.Values {
	resolver := make(chan defines.Values, 1)
	h.c.cbs.Set(h.idx, func(ret defines.Values) {
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
		return defines.Empty
	}
}

func (h *serverRespHandler) AsyncGetResponse(callback func(defines.Values)) {
	if h.ctx == nil {
		h.c.cbs.Set(h.idx, callback)
	} else {
		resolver := make(chan defines.Values, 1)
		h.c.cbs.Set(h.idx, func(ret defines.Values) {
			resolver <- ret
		})
		go func() {
			select {
			case ret := <-resolver:
				callback(ret)
			case <-h.ctx.Done():
				h.c.cbs.Delete(h.idx)
				callback(defines.Empty)
				return
			}
		}()
	}
	h.doSend()
}

func (c *FrameAPIServerConn) CallWithResponse(api string, args defines.Values) defines.ZMQResultHandler {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	idx := uuid.New().String()
	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
	return &serverRespHandler{
		idx, frames, c, nil,
	}
}
