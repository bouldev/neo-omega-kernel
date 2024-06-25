package underlay_conn

import (
	"context"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/minecraft_neo/can_close"
	conn_defines "neo-omega-kernel/minecraft_neo/cascade_conn/defines"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/utils/sync_wrapper"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FrameAPIClient struct {
	can_close.CanCloseWithError
	FrameConn conn_defines.ByteFrameConn
	cbs       *sync_wrapper.SyncKVMap[string, func(defines.Values)]
	apis      *sync_wrapper.SyncKVMap[string, func(defines.Values, func(defines.Values))]
}

func NewFrameAPIClient(conn conn_defines.ByteFrameConn) *FrameAPIClient {
	c := &FrameAPIClient{
		// close underlay conn on err
		CanCloseWithError: can_close.NewClose(conn.Close),
		FrameConn:         conn,
		cbs:               sync_wrapper.NewSyncKVMap[string, func(defines.Values)](),
		apis:              sync_wrapper.NewSyncKVMap[string, func(defines.Values, func(defines.Values))](),
	}
	conn.EnableCompression(packet.SnappyCompression)
	go func() {
		// close when underlay err
		c.CloseWithError(<-conn.WaitClosed())
	}()
	return c
}

func NewFrameAPIClientWithCtx(conn conn_defines.ByteFrameConn, ctx context.Context) *FrameAPIClient {
	c := NewFrameAPIClient(conn)
	go func() {
		select {
		case <-c.WaitClosed():
		case <-ctx.Done():
			c.CloseWithError(ctx.Err())
		}
	}()
	return c
}

func (c *FrameAPIClient) ExposeAPI(apiName string, api defines.ZMQClientAPI, newGoroutine bool) {
	if !strings.HasPrefix(apiName, "/") {
		apiName = "/" + apiName
	}
	c.apis.Set(apiName, func(args defines.Values, setResult func(defines.Values)) {
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

func (c *FrameAPIClient) Run() (err error) {
	go c.FrameConn.ReadRoutine(func(data []byte) {
		frames := bytesToBytesSlices(data)
		indexOrApi := string(frames[0])
		if strings.HasPrefix(indexOrApi, "/") {
			index := frames[1]
			if apiFn, ok := c.apis.Get(indexOrApi); ok {
				apiFn(frames[2:], func(z defines.Values) {
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
	return <-c.WaitClosed()
}

func (c *FrameAPIClient) CallOmitResponse(api string, args defines.Values) {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	frames := append([][]byte{[]byte(api), {}}, args...)
	c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
}

type clientRespHandler struct {
	idx    string
	frames [][]byte
	c      *FrameAPIClient
	ctx    context.Context
}

func (h *clientRespHandler) doSend() {
	h.c.FrameConn.WriteBytePacket(byteSlicesToBytes(h.frames))
}

func (h *clientRespHandler) SetContext(ctx context.Context) defines.ZMQResultHandler {
	h.ctx = ctx
	return h
}

func (h *clientRespHandler) SetTimeout(timeout time.Duration) defines.ZMQResultHandler {
	if h.ctx == nil {
		h.ctx = context.Background()
	}
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *clientRespHandler) BlockGetResponse() defines.Values {
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

func (h *clientRespHandler) AsyncGetResponse(callback func(defines.Values)) {
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

func (c *FrameAPIClient) CallWithResponse(api string, args defines.Values) defines.ZMQResultHandler {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	idx := uuid.New().String()
	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
	return &clientRespHandler{
		idx, frames, c, nil,
	}
}
