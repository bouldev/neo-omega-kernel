package nodes

import (
	"context"
	"fmt"
	"time"
)

type ZMQSlaveNode struct {
	client        ZMQAPIClient
	localAPI      *LocalAPINode
	localTags     *LocalTags
	localTopicNet *LocalTopicNet
	runErrs       chan error
}

func (n *ZMQSlaveNode) IsMaster() bool {
	return false
}

func (n *ZMQSlaveNode) run() {
	go func() {
		for {
			if !n.client.CallWithResponse("/ping", Empty).SetTimeout(time.Second).BlockGetResponse().EqualString("pong") {
				n.runErrs <- fmt.Errorf("disconnected")
				break
			}
			time.Sleep(time.Second)
		}
	}()
	go func() {
		n.client.Run()
		n.runErrs <- fmt.Errorf("disconnected")
	}()
}

func (n *ZMQSlaveNode) Dead() chan error {
	return n.runErrs
}

func (n *ZMQSlaveNode) ListenMessage(topic string, listener MsgListener) {
	n.client.CallOmitResponse("/subscribe", FromString(topic))
	n.localTopicNet.ListenMessage(topic, listener)
}

func (n *ZMQSlaveNode) PublishMessage(topic string, msg Values) {
	n.client.CallOmitResponse("/publish", FromString(topic).Extend(msg))
	n.localTopicNet.publishMessage(topic, msg)
}

func (n *ZMQSlaveNode) ExposeAPI(apiName string, api API) error {
	r := n.client.CallWithResponse("/reg_api", FromString(apiName)).SetTimeout(time.Second).BlockGetResponse()
	if r.EqualString("ok") {
		// salve to master & salve (other call)
		n.client.ExposeAPI(apiName, func(args Values) Values {
			result, err := api(args)
			return wrapOutput(result, err)
		})
		// salve to salve (self) call
		n.localAPI.ExposeAPI(apiName, api)
		return nil
	} else {
		return fmt.Errorf(r.ToString())
	}
}

func (c *ZMQSlaveNode) CallOmitResponse(api string, args Values) {
	if c.localAPI.HasAPI(api) {
		c.localAPI.CallOmitResponse(api, args)
	} else {
		c.client.CallOmitResponse(api, args)
	}
}

type salveResultUpdateHandler struct {
	ZMQResultHandler
}

func (h *salveResultUpdateHandler) SetContext(ctx context.Context) RemoteResultHandler {
	h.ZMQResultHandler.SetContext(ctx)
	return h
}

func (h *salveResultUpdateHandler) SetTimeout(timeout time.Duration) RemoteResultHandler {
	h.ZMQResultHandler.SetTimeout(timeout)
	return h
}

func (h *salveResultUpdateHandler) BlockGetResponse() (Values, error) {
	nestedRet := h.ZMQResultHandler.BlockGetResponse()
	return unwrapOutput(nestedRet)
}

func (h *salveResultUpdateHandler) AsyncGetResponse(callback func(Values, error)) {
	h.ZMQResultHandler.AsyncGetResponse(func(nestedRet Values) {
		rets, err := unwrapOutput(nestedRet)
		callback(rets, err)
	})
}

func (c *ZMQSlaveNode) CallWithResponse(api string, args Values) RemoteResultHandler {
	if c.localAPI.HasAPI(api) {
		return c.localAPI.CallWithResponse(api, args)
	} else {
		return &salveResultUpdateHandler{c.client.CallWithResponse(api, args)}
	}
}

func (c *ZMQSlaveNode) GetValue(key string) (val Values, found bool) {
	v, err := c.CallWithResponse("/get-value", FromString(key)).BlockGetResponse()
	if err != nil || v.IsEmpty() {
		return nil, false
	} else {
		return v, true
	}
}

func (c *ZMQSlaveNode) SetValue(key string, val Values) {
	c.CallOmitResponse("/set-value", FromString(key).Extend(val))
}

func (c *ZMQSlaveNode) SetTags(tags ...string) {
	c.CallOmitResponse("/set-tags", FromStrings(tags...))
	c.localTags.SetTags(tags...)
}

func (c *ZMQSlaveNode) CheckNetTag(tag string) bool {
	if c.localTags.CheckLocalTag(tag) {
		return true
	}
	rest, err := c.CallWithResponse("/check-tag", FromString(tag)).BlockGetResponse()
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
	rest, err := c.CallWithResponse("/try-lock", FromString(name).Extend(FromInt64(acquireTime.Milliseconds()))).BlockGetResponse()
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
	rest, err := c.CallWithResponse("/reset-lock-time", FromString(name).Extend(FromInt64(acquireTime.Milliseconds()))).BlockGetResponse()
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
	c.CallOmitResponse("/unlock", FromString(name))
}

func NewZMQSlaveNode(client ZMQAPIClient) (Node, error) {
	slave := &ZMQSlaveNode{
		client:        client,
		localAPI:      NewLocalAPINode(),
		localTags:     NewLocalTags(),
		localTopicNet: NewLocalTopicNet(),
		runErrs:       make(chan error, 2),
	}
	client.ExposeAPI("/ping", func(args Values) Values {
		return Values{[]byte("pong")}
	})
	client.ExposeAPI("/on_new_msg", func(args Values) Values {
		topic, err := args.ToString()
		if err != nil {
			return Empty
		}
		msg := args.ConsumeHead()
		slave.localTopicNet.PublishMessage(topic, msg)
		return Empty
	})

	go slave.run()

	if !slave.client.CallWithResponse("/new_client", Empty).BlockGetResponse().EqualString("ok") {
		return nil, fmt.Errorf("version mismatch")
	} else {
		return slave, nil
	}
}
