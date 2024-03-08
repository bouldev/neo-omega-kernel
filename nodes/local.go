package nodes

import (
	"context"
	"errors"
	"fmt"
	"neo-omega-kernel/utils/sync_wrapper"
	"time"
)

type alwaysErrorResultHandler struct{ error }

func (h *alwaysErrorResultHandler) SetContext(ctx context.Context) RemoteResultHandler {
	return h
}

func (h *alwaysErrorResultHandler) SetTimeout(timeout time.Duration) RemoteResultHandler {
	return h
}

func (h *alwaysErrorResultHandler) BlockGetResponse() (Values, error) {
	return Empty, h
}

func (h *alwaysErrorResultHandler) AsyncGetResponse(callback func(Values, error)) {
	go func() { callback(Empty, h) }()
}

type LocalAPINode struct {
	RegedApi *sync_wrapper.SyncKVMap[string, AsyncAPI]
}

var ErrAPIExist = errors.New("API already exposed")
var ErrAPINotExist = errors.New("API not exist")

func (n *LocalAPINode) ExposeAPI(apiName string, api API, newGoroutine bool) error {
	if n.HasAPI(apiName) {
		return ErrAPIExist
	}
	n.RegedApi.Set(apiName, func(args Values, setResult func(Values, error)) {
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
	asyncAPI AsyncAPI
	args     Values
	ctx      context.Context
}

func (h *remoteRespHandler) SetContext(ctx context.Context) RemoteResultHandler {
	h.ctx = ctx
	return h
}

func (h *remoteRespHandler) SetTimeout(timeout time.Duration) RemoteResultHandler {
	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
	return h
}

func (h *remoteRespHandler) BlockGetResponse() (Values, error) {
	w := make(chan struct {
		Values
		error
	})
	h.AsyncGetResponse(func(ret Values, err error) {
		w <- struct {
			Values
			error
		}{
			ret, err,
		}
	})
	r := <-w
	return r.Values, r.error
}

func (h *remoteRespHandler) AsyncGetResponse(callback func(Values, error)) {
	resolver := make(chan struct {
		Values
		error
	}, 1)
	h.asyncAPI(h.args, func(ret Values, err error) {
		resolver <- struct {
			Values
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
			callback(Empty, fmt.Errorf("timeout"))
			return
		}
	}()
}

func (c *LocalAPINode) CallWithResponse(api string, args Values) RemoteResultHandler {
	if asyncAPI, ok := c.RegedApi.Get(api); ok {
		return &remoteRespHandler{
			asyncAPI, args, context.Background(),
		}
	} else {
		return &alwaysErrorResultHandler{ErrAPINotExist}
	}
}

func (c *LocalAPINode) CallOmitResponse(api string, args Values) {
	if asyncAPI, ok := c.RegedApi.Get(api); ok {
		asyncAPI(args, func(Values, error) {})
	}
}

func NewLocalAPINode() *LocalAPINode {
	return &LocalAPINode{
		RegedApi: sync_wrapper.NewSyncKVMap[string, AsyncAPI](),
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
	listenedTopics *sync_wrapper.SyncKVMap[string, []MsgListener]
}

func NewLocalTopicNet() *LocalTopicNet {
	return &LocalTopicNet{
		listenedTopics: sync_wrapper.NewSyncKVMap[string, []MsgListener](),
	}
}

func (n *LocalTopicNet) ListenMessage(topic string, listener MsgListener, newGoroutine bool) {
	wrappedListener := func(msg Values) {
		if newGoroutine {
			go listener(msg)
		} else {
			listener(msg)
		}
	}
	n.listenedTopics.UnsafeGetAndUpdate(topic, func(currentListeners []MsgListener) []MsgListener {
		if currentListeners == nil {
			return []MsgListener{wrappedListener}
		}
		currentListeners = append(currentListeners, wrappedListener)
		return currentListeners
	})
}

func (n *LocalTopicNet) publishMessage(topic string, msg Values) Values {
	msgWithTopic := FromString(topic).Extend(msg)
	listeners, _ := n.listenedTopics.Get(topic)
	for _, listener := range listeners {
		listener(msg)
	}
	return msgWithTopic
}

func (n *LocalTopicNet) PublishMessage(topic string, msg Values) {
	n.publishMessage(topic, msg)
}

type LocalNode struct {
	*LocalAPINode
	*LocalLock
	*LocalTags
	*LocalTopicNet
	errChan chan error
	values  *sync_wrapper.SyncKVMap[string, Values]
}

func (n *LocalNode) Dead() chan error {
	return n.errChan
}

func (n *LocalNode) GetValue(key string) (val Values, found bool) {
	return n.values.Get(key)
}
func (n *LocalNode) SetValue(key string, val Values) {
	n.values.Set(key, val)
}

func NewLocalNode(ctx context.Context) Node {
	errChan := make(chan error)
	go func() {
		<-ctx.Done()
		errChan <- ctx.Err()
	}()
	return &LocalNode{
		LocalAPINode:  NewLocalAPINode(),
		LocalLock:     NewLocalLock(),
		LocalTags:     NewLocalTags(),
		LocalTopicNet: NewLocalTopicNet(),
		errChan:       errChan,
		values:        sync_wrapper.NewSyncKVMap[string, Values](),
	}
}
