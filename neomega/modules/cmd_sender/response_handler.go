package cmd_sender

import (
	"context"
	"neo-omega-kernel/minecraft/protocol/packet"
	"neo-omega-kernel/neomega"
	"neo-omega-kernel/utils/sync_wrapper"
	"time"
)

type CmdResponseHandle struct {
	deferredActon         func()
	ctx                   context.Context
	timeoutSpecificResult *packet.CommandOutput
	terminated            bool
	uuidStr               string
	cbByUUID              *sync_wrapper.SyncKVMap[string, func(*packet.CommandOutput)]
}

func (h *CmdResponseHandle) SetTimeoutResponse(timeoutResponse *packet.CommandOutput) neomega.ResponseHandle {
	h.timeoutSpecificResult = timeoutResponse
	return h
}

func (h *CmdResponseHandle) SetTimeout(timeout time.Duration) neomega.ResponseHandle {
	ctx, _ := context.WithTimeout(h.ctx, timeout)
	h.ctx = ctx
	return h
}

func (h *CmdResponseHandle) SetContext(ctx context.Context) neomega.ResponseHandle {
	h.ctx = ctx
	return h
}

func (h *CmdResponseHandle) BlockGetResult() *packet.CommandOutput {
	if h.terminated {
		panic("program logic error, want double result destination specific")
	}
	h.terminated = true
	resolver := make(chan *packet.CommandOutput, 1)
	h.cbByUUID.Set(h.uuidStr, func(co *packet.CommandOutput) {
		resolver <- co
	})
	h.deferredActon()
	select {
	case ret := <-resolver:
		return ret
	case <-h.ctx.Done():
		h.cbByUUID.Delete(h.uuidStr)
		return h.timeoutSpecificResult
	}
}

func (h *CmdResponseHandle) AsyncGetResult(cb func(output *packet.CommandOutput)) {
	if h.terminated {
		panic("program logic error, want double result destination specific")
	}
	h.terminated = true
	resolver := make(chan *packet.CommandOutput, 1)
	h.cbByUUID.Set(h.uuidStr, func(co *packet.CommandOutput) {
		resolver <- co
	})
	go func() {
		select {
		case ret := <-resolver:
			cb(ret)
			return
		case <-h.ctx.Done():
			h.cbByUUID.Delete(h.uuidStr)
			cb(h.timeoutSpecificResult)

			return
		}
	}()
	h.deferredActon()
}
