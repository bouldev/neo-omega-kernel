package nodes

import (
	"context"
	"errors"
	"fmt"
	"neo-omega-kernel/minecraft_neo/can_close"
	"neo-omega-kernel/nodes/defines"
	"neo-omega-kernel/utils/sync_wrapper"
	"time"
)

type alwaysErrorResultHandler struct{ error }

func (h *alwaysErrorResultHandler) SetContext(ctx context.Context) defines.RemoteResultHandler {
	return h
}

func (h *alwaysErrorResultHandler) SetTimeout(timeout time.Duration) defines.RemoteResultHandler {
	return h
}

func (h *alwaysErrorResultHandler) BlockGetResponse() (defines.Values, error) {
	return defines.Empty, h
}

func (h *alwaysErrorResultHandler) AsyncGetResponse(callback func(defines.Values, error)) {
	go func() { callback(defines.Empty, h) }()
}

type LocalAPINode struct {
	RegedApi *sync_wrapper.SyncKVMap[string, defines.AsyncAPI]
}

var ErrAPIExist = errors.New("defines.API already exposed")
var ErrAPINotExist = errors.New("defines.API not exist")

func (n *LocalAPINode) ExposeAPI(apiName string, api defines.API, newGoroutine bool) error {
	if n.HasAPI(apiName) {
		return ErrAPIExist
	}
	n.RegedApi.Set(apiName, func(args defines.Values, setResult func(defines.Values, error)) {
		if newGoroutine {
			go func() {
				rets, err := api(args)
				setResult(rets, err)
			}()
		} else {
			rets, err := api(args)
			setResult(rets, err)
		}
	})
	return nil
}

func (n *LocalAPINode) RemoveAPI(apiName string) {
	n.RegedApi.Delete(apiName)
}

func (n *LocalAPINode) HasAPI(apiName string) bool {
	_, found := n.RegedApi.Get(apiName)
	return found
}

type remoteRespHandler struct {
	asyncAPI defines.AsyncAPI
	args     defines.Values
	ctx      context.Context
}

func (h *remoteRespHandler) SetContext(ctx context.Context) defines.RemoteResultHandler {
	h.ctx = ctx
	return h
}

func (h *remoteRespHandler) SetTimeout(timeout time.Duration) defines.RemoteResultHandler {
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *remoteRespHandler) BlockGetResponse() (defines.Values, error) {
	w := make(chan struct {
		defines.Values
		error
	})
	h.AsyncGetResponse(func(ret defines.Values, err error) {
		w <- struct {
			defines.Values
			error
		}{
			ret, err,
		}
	})
	r := <-w
	return r.Values, r.error
}

func (h *remoteRespHandler) AsyncGetResponse(callback func(defines.Values, error)) {
	resolver := make(chan struct {
		defines.Values
		error
	}, 1)
	h.asyncAPI(h.args, func(ret defines.Values, err error) {
		resolver <- struct {
			defines.Values
			error
		}{
			ret, err,
		}
	})
	go func() {
		select {
		case ret := <-resolver:
			callback(ret.Values, ret.error)
		case <-h.ctx.Done():
			callback(defines.Empty, fmt.Errorf("timeout"))
			return
		}
	}()
}

func (c *LocalAPINode) CallWithResponse(api string, args defines.Values) defines.RemoteResultHandler {
	if asyncAPI, ok := c.RegedApi.Get(api); ok {
		return &remoteRespHandler{
			asyncAPI, args, context.Background(),
		}
	} else {
		return &alwaysErrorResultHandler{ErrAPINotExist}
	}
}

func (c *LocalAPINode) CallOmitResponse(api string, args defines.Values) {
	if asyncAPI, ok := c.RegedApi.Get(api); ok {
		asyncAPI(args, func(defines.Values, error) {})
	}
}

func NewLocalAPINode() *LocalAPINode {
	return &LocalAPINode{
		RegedApi: sync_wrapper.NewSyncKVMap[string, defines.AsyncAPI](),
	}
}

type timeLock struct {
	ctx      context.Context
	reset    chan struct{}
	cancelFN func()
	owner    string
	unlocked bool
}

type LocalLock struct {
	Locks *sync_wrapper.SyncKVMap[string, *timeLock]
}

func (c *LocalLock) tryLock(name string, acquireTime time.Duration, owner string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	if acquireTime > 0 {
		ctx, _ = context.WithTimeout(ctx, acquireTime)
	}
	l := &timeLock{
		ctx:      ctx,
		reset:    make(chan struct{}),
		cancelFN: cancel,
		owner:    owner,
	}
	_, locked := c.Locks.GetOrSet(name, l)
	if locked {
		return false
	} else {
		go func() {
			select {
			case <-l.ctx.Done():
				if !l.unlocked {
					c.unlock(name, owner)
				}
			case <-l.reset:
				return
			}
		}()
	}
	return true
}

func (c *LocalLock) unlock(name string, owner string) bool {
	l, ok := c.Locks.GetAndDelete(name)
	if !ok {
		return false
	}
	if l.owner == owner {
		l.unlocked = true
		l.cancelFN()
		return true
	}
	return false
}

func (c *LocalLock) resetLockTime(name string, acquireTime time.Duration, owner string) bool {
	l, ok := c.Locks.Get(name)
	if !ok {
		return false
	}
	if l.owner == owner {
		// remove previous ctx
		previousReset := l.reset

		// make a new ctx
		ctx, cancel := context.WithTimeout(context.Background(), acquireTime)
		if acquireTime > 0 {

		} else {
			if !l.unlocked {
				c.unlock(name, owner)
			}
			c.Locks.Delete(name)
			cancel()
			return true
		}
		l.ctx = ctx
		l.cancelFN = cancel
		l.reset = make(chan struct{})
		go func() {
			select {
			case <-l.ctx.Done():
				if !l.unlocked {
					c.unlock(name, owner)
				}
			case <-l.reset:
				return
			}
		}()
		close(previousReset)
		return true
	}
	return false
}

func (c *LocalLock) TryLock(name string, acquireTime time.Duration) bool {
	return c.tryLock(name, acquireTime, "")
}

func (c *LocalLock) ResetLockTime(name string, acquireTime time.Duration) bool {
	return c.resetLockTime(name, acquireTime, "")
}

func (c *LocalLock) Unlock(name string) {
	c.unlock(name, "")
}

func NewLocalLock() *LocalLock {
	return &LocalLock{
		Locks: sync_wrapper.NewSyncKVMap[string, *timeLock](),
	}
}

type LocalTags struct {
	tags *sync_wrapper.SyncKVMap[string, struct{}]
}

func (n *LocalTags) SetTags(tags ...string) {
	for _, tag := range tags {
		n.tags.Set(tag, struct{}{})
	}

}
func (n *LocalTags) CheckNetTag(tag string) bool {
	_, ok := n.tags.Get(tag)
	return ok
}

func (n *LocalTags) CheckLocalTag(tag string) bool {
	return n.CheckNetTag(tag)
}

func NewLocalTags() *LocalTags {
	return &LocalTags{
		tags: sync_wrapper.NewSyncKVMap[string, struct{}](),
	}
}

type LocalTopicNet struct {
	listenedTopics *sync_wrapper.SyncKVMap[string, []defines.MsgListener]
}

func NewLocalTopicNet() *LocalTopicNet {
	return &LocalTopicNet{
		listenedTopics: sync_wrapper.NewSyncKVMap[string, []defines.MsgListener](),
	}
}

func (n *LocalTopicNet) ListenMessage(topic string, listener defines.MsgListener, newGoroutine bool) {
	wrappedListener := func(msg defines.Values) {
		if newGoroutine {
			go listener(msg)
		} else {
			listener(msg)
		}
	}
	n.listenedTopics.UnsafeGetAndUpdate(topic, func(currentListeners []defines.MsgListener) []defines.MsgListener {
		if currentListeners == nil {
			return []defines.MsgListener{wrappedListener}
		}
		currentListeners = append(currentListeners, wrappedListener)
		return currentListeners
	})
}

func (n *LocalTopicNet) publishMessage(topic string, msg defines.Values) defines.Values {
	msgWithTopic := defines.FromString(topic).Extend(msg)
	listeners, _ := n.listenedTopics.Get(topic)
	for _, listener := range listeners {
		listener(msg)
	}
	return msgWithTopic
}

func (n *LocalTopicNet) PublishMessage(topic string, msg defines.Values) {
	n.publishMessage(topic, msg)
}

type LocalNode struct {
	*LocalAPINode
	*LocalLock
	*LocalTags
	*LocalTopicNet
	can_close.CanCloseWithError
	values *sync_wrapper.SyncKVMap[string, defines.Values]
}

func (n *LocalNode) GetValue(key string) (val defines.Values, found bool) {
	return n.values.Get(key)
}
func (n *LocalNode) SetValue(key string, val defines.Values) {
	n.values.Set(key, val)
}

func NewLocalNode(ctx context.Context) defines.Node {
	n := &LocalNode{
		LocalAPINode:      NewLocalAPINode(),
		LocalLock:         NewLocalLock(),
		LocalTags:         NewLocalTags(),
		LocalTopicNet:     NewLocalTopicNet(),
		CanCloseWithError: can_close.NewClose(func() {}),
		values:            sync_wrapper.NewSyncKVMap[string, defines.Values](),
	}

	go func() {
		<-ctx.Done()
		n.CloseWithError(ctx.Err())
	}()

	return n
}
