package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/nodes/defines"
)

type ZMQSlaveNode struct {
	client        defines.ZMQAPIClient
	localAPI      *LocalAPINode
	localTags     *LocalTags
	localTopicNet *LocalTopicNet
	can_close.CanCloseWithError
}

func (n *ZMQSlaveNode) IsMaster() bool {
	return false
}

// func (n *ZMQSlaveNode) heatBeat() {
// 	go func() {
// 		for {
// 			if !n.client.CallWithResponse("/ping", defines.Empty).SetTimeout(time.Second).BlockGetResponse().EqualString("pong") {
// 				n.CloseWithError(fmt.Errorf("disconnected"))
// 				break
// 			}
// 			time.Sleep(time.Second * 5)
// 		}
// 	}()
// }

func (n *ZMQSlaveNode) ListenMessage(topic string, listener defines.MsgListener, newGoroutine bool) {
	n.client.CallWithResponse("/subscribe", defines.FromString(topic)).BlockGetResponse()
	n.localTopicNet.ListenMessage(topic, listener, newGoroutine)
}

func (n *ZMQSlaveNode) PublishMessage(topic string, msg defines.Values) {
	n.client.CallOmitResponse("/publish", defines.FromString(topic).Extend(msg))
	n.localTopicNet.publishMessage(topic, msg)
}

func (n *ZMQSlaveNode) ExposeAPI(apiName string, api defines.API, newGoroutine bool) error {
	r := n.client.CallWithResponse("/reg_api", defines.FromString(apiName)).SetTimeout(time.Second).BlockGetResponse()
	if r.EqualString("ok") {
		// salve to master & salve (other call)
		n.client.ExposeAPI(apiName, func(args defines.Values) defines.Values {
			result, err := api(args)
			return defines.WrapOutput(result, err)
		}, false)
		// salve to salve (self) call
		n.localAPI.ExposeAPI(apiName, api, newGoroutine)
		return nil
	} else {
		return fmt.Errorf(r.ToString())
	}
}

func (c *ZMQSlaveNode) CallOmitResponse(api string, args defines.Values) {
	if c.localAPI.HasAPI(api) {
		c.localAPI.CallOmitResponse(api, args)
	} else {
		c.client.CallOmitResponse(api, args)
	}
}

type salveResultUpdateHandler struct {
	defines.ZMQResultHandler
}

func (h *salveResultUpdateHandler) SetContext(ctx context.Context) defines.RemoteResultHandler {
	h.ZMQResultHandler.SetContext(ctx)
	return h
}

func (h *salveResultUpdateHandler) SetTimeout(timeout time.Duration) defines.RemoteResultHandler {
	h.ZMQResultHandler.SetTimeout(timeout)
	return h
}

func (h *salveResultUpdateHandler) BlockGetResponse() (defines.Values, error) {
	nestedRet := h.ZMQResultHandler.BlockGetResponse()
	return defines.UnwrapOutput(nestedRet)
}

func (h *salveResultUpdateHandler) AsyncGetResponse(callback func(defines.Values, error)) {
	h.ZMQResultHandler.AsyncGetResponse(func(nestedRet defines.Values) {
		rets, err := defines.UnwrapOutput(nestedRet)
		callback(rets, err)
	})
}

func (c *ZMQSlaveNode) CallWithResponse(api string, args defines.Values) defines.RemoteResultHandler {
	if c.localAPI.HasAPI(api) {
		return c.localAPI.CallWithResponse(api, args)
	} else {
		return &salveResultUpdateHandler{c.client.CallWithResponse(api, args)}
	}
}

func (c *ZMQSlaveNode) GetValue(key string) (val defines.Values, found bool) {
	v, err := c.CallWithResponse("/get-value", defines.FromString(key)).BlockGetResponse()
	if err != nil || v.IsEmpty() {
		return nil, false
	} else {
		return v, true
	}
}

func (c *ZMQSlaveNode) SetValue(key string, val defines.Values) {
	c.CallOmitResponse("/set-value", defines.FromString(key).Extend(val))
}

func (c *ZMQSlaveNode) SetTags(tags ...string) {
	c.CallOmitResponse("/set-tags", defines.FromStrings(tags...))
	c.localTags.SetTags(tags...)
}

func (c *ZMQSlaveNode) CheckNetTag(tag string) bool {
	if c.localTags.CheckLocalTag(tag) {
		return true
	}
	rest, err := c.CallWithResponse("/check-tag", defines.FromString(tag)).BlockGetResponse()
	if err != nil || rest.IsEmpty() {
		return false
	}
	hasTag, err := rest.ToBool()
	if err != nil {
		return false
	}
	return hasTag
}

func (n *ZMQSlaveNode) CheckLocalTag(tag string) bool {
	return n.localTags.CheckLocalTag(tag)
}

func (c *ZMQSlaveNode) TryLock(name string, acquireTime time.Duration) bool {
	rest, err := c.CallWithResponse("/try-lock", defines.FromString(name).Extend(defines.FromInt64(acquireTime.Milliseconds()))).BlockGetResponse()
	if err != nil || rest.IsEmpty() {
		return false
	}
	locked, err := rest.ToBool()
	if err != nil {
		return false
	}
	return locked
}

func (c *ZMQSlaveNode) ResetLockTime(name string, acquireTime time.Duration) bool {
	rest, err := c.CallWithResponse("/reset-lock-time", defines.FromString(name).Extend(defines.FromInt64(acquireTime.Milliseconds()))).BlockGetResponse()
	if err != nil || rest.IsEmpty() {
		return false
	}
	locked, err := rest.ToBool()
	if err != nil {
		return false
	}
	return locked
}

func (c *ZMQSlaveNode) Unlock(name string) {
	c.CallOmitResponse("/unlock", defines.FromString(name))
}

func NewZMQSlaveNode(client defines.ZMQAPIClient) (defines.Node, error) {
	slave := &ZMQSlaveNode{
		client:            client,
		localAPI:          NewLocalAPINode(),
		localTags:         NewLocalTags(),
		localTopicNet:     NewLocalTopicNet(),
		CanCloseWithError: can_close.NewClose(client.Close),
	}
	client.ExposeAPI("/ping", func(args defines.Values) defines.Values {
		return defines.Values{[]byte("pong")}
	}, false)
	client.ExposeAPI("/on_new_msg", func(args defines.Values) defines.Values {
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty
		}
		msg := args.ConsumeHead()
		slave.localTopicNet.PublishMessage(topic, msg)
		return defines.Empty
	}, false)

	go func() {
		slave.CloseWithError(<-client.WaitClosed())
	}()
	// go slave.heatBeat()

	if !slave.client.CallWithResponse("/new_client", defines.Empty).BlockGetResponse().EqualString("ok") {
		return nil, fmt.Errorf("version mismatch")
	} else {
		return slave, nil
	}
}
