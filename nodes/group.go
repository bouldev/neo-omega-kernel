package nodes

import (
	"strings"
	"time"
)

type group struct {
	Node
	name          string
	allowAbsolute bool
}

func (n *group) translateName(name string) string {
	// absolute path
	if n.allowAbsolute && strings.HasPrefix(name, "/") {
		return name
	} else {
		return n.name + "/" + name
	}
}

func (n *group) ExposeAPI(apiName string, api API) error {
	return n.Node.ExposeAPI(n.translateName(apiName), api)
}
func (n *group) CallOmitResponse(api string, args Values) {
	n.Node.CallOmitResponse(n.translateName(api), args)
}
func (n *group) CallWithResponse(api string, args Values) RemoteResultHandler {
	return n.Node.CallWithResponse(n.translateName(api), args)
}
func (n *group) PublishMessage(topic string, msg Values) {
	n.Node.PublishMessage(n.translateName(topic), msg)
}
func (n *group) ListenMessage(topic string, listener MsgListener) {
	n.Node.ListenMessage(n.translateName(topic), listener)
}
func (n *group) GetValue(key string) (val Values, found bool) {
	return n.Node.GetValue(n.translateName(key))
}
func (n *group) SetValue(key string, val Values) {
	n.Node.SetValue(n.translateName(key), val)
}
func (n *group) SetTags(tags ...string) {
	ttrags := []string{}
	for _, tag := range tags {
		ttrags = append(ttrags, n.translateName(tag))
	}
	n.Node.SetTags(ttrags...)
}
func (n *group) CheckNetTag(tag string) bool {
	return n.Node.CheckNetTag(n.translateName(tag))
}
func (n *group) CheckLocalTag(tag string) bool {
	return n.Node.CheckLocalTag(n.translateName(tag))
}
func (n *group) TryLock(name string, acquireTime time.Duration) bool {
	return n.Node.TryLock(n.translateName(name), acquireTime)
}

func (n *group) ResetLockTime(name string, acquireTime time.Duration) bool {
	return n.Node.ResetLockTime(n.translateName(name), acquireTime)
}

func (n *group) Unlock(name string) {
	n.Node.Unlock(n.translateName(name))
}

func NewGroup(name string, node Node, allowAbsolute bool) Node {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return &group{node, name, allowAbsolute}
}
