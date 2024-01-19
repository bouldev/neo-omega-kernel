package neomega

import (
	"context"
	"errors"
	"time"
)

var ErrTimeout = errors.New("time out")

type AsyncResponseHandle[T any] struct {
	deferredActon          func(resolve func(result T, err error))
	cleanUpActionOnTimeout func()
	ctx                    context.Context
	timeoutResult          struct {
		result T
		err    error
	}
	terminated bool
}

func NewAsyncResponseHandle[T any](deferAction func(resolve func(result T, err error)), cleanUpActionOnTimeout func()) *AsyncResponseHandle[T] {
	return &AsyncResponseHandle[T]{
		deferredActon:          deferAction,
		cleanUpActionOnTimeout: cleanUpActionOnTimeout,
		ctx:                    context.Background(),
		timeoutResult: struct {
			result T
			err    error
		}{
			err: ErrTimeout,
		},
	}
}

func (r *AsyncResponseHandle[T]) SetTimeoutResponse(result T, err error) *AsyncResponseHandle[T] {
	r.timeoutResult = struct {
		result T
		err    error
	}{
		result: result,
		err:    err,
	}
	return r
}

func (r *AsyncResponseHandle[T]) SetTimeout(timeout time.Duration) *AsyncResponseHandle[T] {
	ctx, _ := context.WithTimeout(r.ctx, timeout)
	r.ctx = ctx
	return r
}

func (r *AsyncResponseHandle[T]) SetContext(ctx context.Context) *AsyncResponseHandle[T] {
	r.ctx = ctx
	return r
}

func (h *AsyncResponseHandle[T]) BlockGetResult() (result T, err error) {
	if h.terminated {
		panic("program logic error, want double result destination specific")
	}
	h.terminated = true
	resolver := make(chan struct {
		result T
		err    error
	}, 1)
	h.deferredActon(func(r T, e error) {
		resolver <- struct {
			result T
			err    error
		}{r, e}
	})
	select {
	case ret := <-resolver:
		return ret.result, ret.err
	case <-h.ctx.Done():
		h.cleanUpActionOnTimeout()
		return h.timeoutResult.result, h.timeoutResult.err
	}
}

func (h *AsyncResponseHandle[T]) AsyncGetResult(cb func(T, error)) {
	if h.terminated {
		panic("program logic error, want double result destination specific")
	}
	h.terminated = true
	resolver := make(chan struct {
		result T
		err    error
	}, 1)

	go func() {
		select {
		case ret := <-resolver:
			cb(ret.result, ret.err)
			return
		case <-h.ctx.Done():
			h.cleanUpActionOnTimeout()
			cb(h.timeoutResult.result, h.timeoutResult.err)
			return
		}
	}()
	h.deferredActon(func(r T, e error) {
		resolver <- struct {
			result T
			err    error
		}{r, e}
	})
}
