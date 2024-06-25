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

type SlaveNodeInfo struct {
	Ctx              context.Context
	cancelFn         func()
	MsgToPub         chan defines.Values
	SubScribedTopics *sync_wrapper.SyncKVMap[string, struct{}]
	ExposedApis      *sync_wrapper.SyncKVMap[string, struct{}]
	AcquiredLocks    *sync_wrapper.SyncKVMap[string, struct{}]
	Tags             *sync_wrapper.SyncKVMap[string, struct{}]
}

var ErrNotReg = errors.New("need reg node first")

type ZMQMasterNode struct {
	server defines.ZMQAPIServer
	*LocalAPINode
	*LocalLock
	*LocalTags
	*LocalTopicNet
	slaves                *sync_wrapper.SyncKVMap[string, *SlaveNodeInfo]
	slaveSubscribedTopics *sync_wrapper.SyncKVMap[string, *sync_wrapper.SyncKVMap[string, chan defines.Values]]
	ApiProvider           *sync_wrapper.SyncKVMap[string, string]
	values                *sync_wrapper.SyncKVMap[string, defines.Values]
	can_close.CanCloseWithError
}

func (n *ZMQMasterNode) IsMaster() bool {
	return true
}

func (n *ZMQMasterNode) onNewNode(id string) *SlaveNodeInfo {
	// fmt.Println("new client: ", id)
	ctx, cancelFn := context.WithCancel(context.Background())
	nodeInfo := &SlaveNodeInfo{
		Ctx:              ctx,
		cancelFn:         cancelFn,
		MsgToPub:         make(chan defines.Values, 1024),
		SubScribedTopics: sync_wrapper.NewSyncKVMap[string, struct{}](),
		ExposedApis:      sync_wrapper.NewSyncKVMap[string, struct{}](),
		AcquiredLocks:    sync_wrapper.NewSyncKVMap[string, struct{}](),
		Tags:             sync_wrapper.NewSyncKVMap[string, struct{}](),
	}
	n.slaves.Set(id, nodeInfo)
	return nodeInfo
}

func (n *ZMQMasterNode) onNodeOffline(id string, info *SlaveNodeInfo) {
	info.cancelFn()

	info.SubScribedTopics.Iter(func(topic string, v struct{}) (continueInter bool) {
		if slaves, ok := n.slaveSubscribedTopics.Get(topic); ok {
			slaves.Delete(id)
		}
		return true
	})
	info.ExposedApis.Iter(func(apiName string, v struct{}) (continueInter bool) {
		if providerID, ok := n.ApiProvider.Get(apiName); ok && (providerID == id) {
			n.server.ConcealAPI(apiName)
			n.LocalAPINode.RemoveAPI(apiName)
		}
		return true
	})
	info.AcquiredLocks.Iter(func(k string, v struct{}) (continueInter bool) {
		n.unlock(k, id)
		return true
	})
	// close(info.MsgToPub)
	n.slaves.Delete(id)
	// fmt.Printf("node %v offline\n", string(id))
}

func (n *ZMQMasterNode) publishMessage(source string, topic string, msg defines.Values) {
	msgWithTopic := n.LocalTopicNet.publishMessage(topic, msg)
	subScribers, ok := n.slaveSubscribedTopics.Get(topic)
	if ok {
		subScribers.Iter(func(receiver string, msgC chan defines.Values) (continueInter bool) {
			if receiver == source {
				return true
			}
			select {
			case msgC <- msgWithTopic:
			default:
				fmt.Println("communication between nodes too slow, msg queued!")
				go func() {
					msgC <- msgWithTopic
				}()
			}
			return true
		})
	}
}

func (n *ZMQMasterNode) PublishMessage(topic string, msg defines.Values) {
	n.publishMessage("", topic, msg)
}

func (n *ZMQMasterNode) ExposeAPI(apiName string, api defines.API, newGoroutine bool) error {
	// master to master call
	n.LocalAPINode.ExposeAPI(apiName, api, newGoroutine)
	// slave to master call
	n.server.ExposeAPI(apiName, func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		result, err := api(args)
		return defines.WrapOutput(result, err)
	}, newGoroutine)
	return nil
}

func (c *ZMQMasterNode) GetValue(key string) (val defines.Values, found bool) {
	v, ok := c.values.Get(key)
	return v, ok
}

func (c *ZMQMasterNode) SetValue(key string, val defines.Values) {
	c.values.Set(key, val)
}

func (c *ZMQMasterNode) CheckNetTag(tag string) bool {
	ok := c.LocalTags.CheckLocalTag(tag)
	if ok {
		return true
	}
	found := false
	c.slaves.Iter(func(k string, v *SlaveNodeInfo) bool {
		_, ok := v.Tags.Get(tag)
		if ok {
			found = true
		}
		return !found
	})
	return found
}

func (c *ZMQMasterNode) tryLock(name string, acquireTime time.Duration, owner string) bool {
	if !c.LocalLock.tryLock(name, acquireTime, owner) {
		return false
	}
	if owner != "" {
		slaveInfo, ok := c.slaves.Get(owner)
		if ok {
			slaveInfo.AcquiredLocks.Set(name, struct{}{})
		}
	}
	return true
}

func (c *ZMQMasterNode) unlock(name string, owner string) {
	if c.LocalLock.unlock(name, owner) {
		if owner != "" {
			slaveInfo, ok := c.slaves.Get(owner)
			if ok {
				slaveInfo.AcquiredLocks.Delete(name)
			}
		}
	}
}

func NewZMQMasterNode(server defines.ZMQAPIServer) defines.Node {
	master := &ZMQMasterNode{
		server:                server,
		LocalAPINode:          NewLocalAPINode(),
		LocalLock:             NewLocalLock(),
		LocalTags:             NewLocalTags(),
		LocalTopicNet:         NewLocalTopicNet(),
		slaves:                sync_wrapper.NewSyncKVMap[string, *SlaveNodeInfo](),
		slaveSubscribedTopics: sync_wrapper.NewSyncKVMap[string, *sync_wrapper.SyncKVMap[string, chan defines.Values]](),
		ApiProvider:           sync_wrapper.NewSyncKVMap[string, string](),
		values:                sync_wrapper.NewSyncKVMap[string, defines.Values](),
		CanCloseWithError:     can_close.NewClose(server.Close),
	}
	go func() {
		master.CloseWithError(<-server.WaitClosed())
	}()
	server.ExposeAPI("/ping", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		return defines.Values{[]byte("pong")}
	}, false)
	server.ExposeAPI("/new_client", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		nodeInfo := master.onNewNode(string(caller))
		go func() {
			for {
				if !server.CallWithResponse(caller, "/ping", defines.Empty).
					SetTimeout(time.Second).
					BlockGetResponse().
					EqualString("pong") {
					master.onNodeOffline(string(caller), nodeInfo)
					return
				}
				time.Sleep(time.Second)
			}
		}()
		go func() {
			for {
				select {
				case <-nodeInfo.Ctx.Done():
					return
				case msg := <-nodeInfo.MsgToPub:
					server.CallOmitResponse(caller, "/on_new_msg", msg)
				}
			}
		}()
		return defines.FromString("ok")
	}, false)
	server.ExposeAPI("/subscribe", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return defines.Empty
		}
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty
		}
		slaveInfo.SubScribedTopics.Set(topic, struct{}{})
		master.slaveSubscribedTopics.UnsafeGetAndUpdate(topic, func(val *sync_wrapper.SyncKVMap[string, chan defines.Values]) *sync_wrapper.SyncKVMap[string, chan defines.Values] {
			if val == nil {
				val = sync_wrapper.NewSyncKVMap[string, chan defines.Values]()
			}
			val.Set(string(caller), slaveInfo.MsgToPub)
			return val
		})
		return defines.Empty
	}, false)
	server.ExposeAPI("/publish", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty
		}
		msg := args.ConsumeHead()
		master.publishMessage(string(caller), topic, msg)
		return defines.Empty
	}, false)
	server.ExposeAPI("/reg_api", func(provider defines.ZMQCaller, args defines.Values) defines.Values {
		apiName, err := args.ToString()
		if err != nil {
			return defines.Empty
		}
		slaveInfo, ok := master.slaves.Get(string(provider))
		if !ok {
			return defines.FromString(ErrNotReg.Error())
		}
		// reject if api exist
		if master.LocalAPINode.HasAPI(apiName) {
			return defines.FromString(ErrAPIExist.Error())
		}
		slaveInfo.ExposedApis.Set(apiName, struct{}{})
		master.ApiProvider.Set(apiName, string(provider))
		// master to slave call
		master.LocalAPINode.ExposeAPI(apiName, func(args defines.Values) (result defines.Values, err error) {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return defines.Empty, fmt.Errorf("not found")
			} else {
				return defines.UnwrapOutput(server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResponse())
			}
		}, false)
		// slave to salve call
		server.ExposeAPI(apiName, func(caller defines.ZMQCaller, args defines.Values) defines.Values {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return defines.Empty
			}
			return server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResponse()
		}, false)
		return defines.FromString("ok")
	}, false)
	master.ExposeAPI("/set-value", func(args defines.Values) (result defines.Values, err error) {
		key, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		master.values.Set(key, args.ConsumeHead())
		return defines.Empty, nil
	}, false)
	master.ExposeAPI("/get-value", func(args defines.Values) (result defines.Values, err error) {
		key, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		head, ok := master.values.Get(key)
		if ok {
			return head, nil
		} else {
			return defines.Empty, nil
		}
	}, false)
	server.ExposeAPI("/set-tags", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		tags := args.ToStrings()
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return defines.FromString("need reg first")
		}
		for _, tag := range tags {
			slaveInfo.Tags.Set(tag, struct{}{})
		}
		return defines.Empty
	}, false)
	server.ExposeAPI("/check-tag", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		tag, err := args.ToString()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		has := master.CheckNetTag(tag)
		return defines.WrapOutput(defines.FromBool(has), nil)
	}, false)
	server.ExposeAPI("/try-lock", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		lockName, err := args.ToString()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.tryLock(lockName, lockTime, string(caller))
		return defines.WrapOutput(defines.FromBool(has), nil)
	}, false)
	server.ExposeAPI("/reset-lock-time", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		lockName, err := args.ToString()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.resetLockTime(lockName, lockTime, string(caller))
		return defines.WrapOutput(defines.FromBool(has), nil)
	}, false)
	server.ExposeAPI("/unlock", func(caller defines.ZMQCaller, args defines.Values) defines.Values {
		lockName, err := args.ToString()
		if err != nil {
			return defines.WrapOutput(defines.Empty, err)
		}
		master.unlock(lockName, string(caller))
		return defines.Empty
	}, false)
	return master
}
