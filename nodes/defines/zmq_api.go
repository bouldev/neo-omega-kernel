package defines

import (
	"context"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
)

type ZMQCaller string

type ZMQClientAPI func(args Values) Values
type ZMQServerAPI func(caller ZMQCaller, args Values) Values

type ZMQResultHandler interface {
	SetContext(ctx context.Context) ZMQResultHandler
	SetTimeout(timeout time.Duration) ZMQResultHandler
	BlockGetResponse() Values
	AsyncGetResponse(callback func(Values))
}

type ZMQAPIClient interface {
	CallOmitResponse(api string, args Values)
	CallWithResponse(api string, args Values) ZMQResultHandler
	ExposeAPI(apiName string, api ZMQClientAPI, newGoroutine bool)
	can_close.CanClose
}

type ZMQAPIServer interface {
	ExposeAPI(apiName string, api ZMQServerAPI, newGoroutine bool)
	ConcealAPI(apiName string)
	CallOmitResponse(callee ZMQCaller, api string, args Values)
	CallWithResponse(callee ZMQCaller, api string, args Values) ZMQResultHandler
	SetOnCloseCleanUp(callee ZMQCaller, cb func())
	can_close.CanClose
}
