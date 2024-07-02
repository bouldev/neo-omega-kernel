package defines

import (
	"context"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
)

type Values [][]byte
type API func(args Values) (result Values, err error)
type MsgListener func(msg Values)

type RemoteResultHandler interface {
	SetContext(ctx context.Context) RemoteResultHandler
	SetTimeout(timeout time.Duration) RemoteResultHandler
	BlockGetResponse() (Values, error)
	AsyncGetResponse(callback func(Values, error))
}

type APINode interface {
	// Point-to-Point Remote Process Call
	ExposeAPI(apiName string, api API, newGoroutine bool) error
	CallOmitResponse(api string, args Values)
	CallWithResponse(api string, args Values) RemoteResultHandler
}

type TopicNetNode interface {
	// Multi-to-Multi Message Publish & Subscribe
	PublishMessage(topic string, msg Values)
	ListenMessage(topic string, listener MsgListener, newGoroutine bool)
}

type FundamentalNode interface {
	APINode
	TopicNetNode
}

type KVDataNode interface {
	// Global KV data
	GetValue(key string) (val Values, found bool)
	SetValue(key string, val Values)
}

type RolesNode interface {
	// Tags
	SetTags(tags ...string)
	CheckNetTag(tag string) bool
	CheckLocalTag(tag string) bool
}

type TimeLockNode interface {
	// Lock
	TryLock(name string, acquireTime time.Duration) bool
	ResetLockTime(name string, acquireTime time.Duration) bool
	Unlock(name string)
}

type Node interface {
	can_close.CanCloseWithError
	FundamentalNode
	KVDataNode
	RolesNode
	TimeLockNode
}

type AsyncAPI func(args Values, setResult func(rets Values, err error))
