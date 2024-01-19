package nodes

import (
	"context"
	"errors"
	"fmt"
	"neo-omega-kernel/utils/sync_wrapper"
	"time"
)

type SlaveNodeInfo struct {
	Ctx              context.Context
	cancelFn         func()
	MsgToPub         chan Values
	SubScribedTopics *sync_wrapper.SyncKVMap[string, struct{}]
	ExposedApis      *sync_wrapper.SyncKVMap[string, struct{}]
	AcquiredLocks    *sync_wrapper.SyncKVMap[string, struct{}]
	Tags             *sync_wrapper.SyncKVMap[string, struct{}]
}

var ErrNotReg = errors.New("need reg node first")

type AsyncAPI func(args Values, setResult func(rets Values, err error))

type ZMQMasterNode struct {
	server ZMQAPIServer
	*LocalAPINode
	*LocalLock
	*LocalTags
	*LocalTopicNet
	slaves                *sync_wrapper.SyncKVMap[string, *SlaveNodeInfo]
	slaveSubscribedTopics *sync_wrapper.SyncKVMap[string, *sync_wrapper.SyncKVMap[string, chan Values]]
	ApiProvider           *sync_wrapper.SyncKVMap[string, string]
	values                *sync_wrapper.SyncKVMap[string, Values]
	runErr                chan error
}

func (n *ZMQMasterNode) IsMaster() bool {
	return true
}

func (n *ZMQMasterNode) Dead() chan error {
	return n.runErr
}

func (n *ZMQMasterNode) onNewNode(id string) *SlaveNodeInfo {
	// fmt.Println("new client: ", id)
	ctx, cancelFn := context.WithCancel(context.Background())
	nodeInfo := &SlaveNodeInfo{
		Ctx:              ctx,
		cancelFn:         cancelFn,
		MsgToPub:         make(chan Values, 128),
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

func (n *ZMQMasterNode) publishMessage(source string, topic string, msg Values) {
	msgWithTopic := n.LocalTopicNet.publishMessage(topic, msg)
	subScribers, ok := n.slaveSubscribedTopics.Get(topic)
	if ok {
		subScribers.Iter(func(receiver string, msgC chan Values) (continueInter bool) {
			if receiver == source {
				return true
			}
			select {
			case msgC <- msgWithTopic:
			default:
				go func() {
					msgC <- msgWithTopic
				}()
			}
			return true
		})
	}
}

func (n *ZMQMasterNode) PublishMessage(topic string, msg Values) {
	n.publishMessage("", topic, msg)
}

func (n *ZMQMasterNode) ExposeAPI(apiName string, api API) error {
	// master to master call
	n.LocalAPINode.ExposeAPI(apiName, api)
	// slave to master call
	n.server.ExposeAPI(apiName, func(caller ZMQCaller, args Values) Values {
		result, err := api(args)
		return wrapOutput(result, err)
	})
	return nil
}

func (c *ZMQMasterNode) GetValue(key string) (val Values, found bool) {
	v, ok := c.values.Get(key)
	return v, ok
}

func (c *ZMQMasterNode) SetValue(key string, val Values) {
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

func NewZMQMasterNode(server ZMQAPIServer) Node {
	master := &ZMQMasterNode{
		server:                server,
		LocalAPINode:          NewLocalAPINode(),
		LocalLock:             NewLocalLock(),
		LocalTags:             NewLocalTags(),
		LocalTopicNet:         NewLocalTopicNet(),
		slaves:                sync_wrapper.NewSyncKVMap[string, *SlaveNodeInfo](),
		slaveSubscribedTopics: sync_wrapper.NewSyncKVMap[string, *sync_wrapper.SyncKVMap[string, chan Values]](),
		runErr:                make(chan error, 2),
		ApiProvider:           sync_wrapper.NewSyncKVMap[string, string](),
		values:                sync_wrapper.NewSyncKVMap[string, Values](),
	}
	go func() {
		master.runErr <- master.server.Serve()
	}()
	server.ExposeAPI("/ping", func(caller ZMQCaller, args Values) Values {
		return Values{[]byte("pong")}
	})
	server.ExposeAPI("/new_client", func(caller ZMQCaller, args Values) Values {
		nodeInfo := master.onNewNode(string(caller))
		go func() {
			for {
				if !server.CallWithResponse(caller, "/ping", Empty).
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
		return FromString("ok")
	})
	server.ExposeAPI("/subscribe", func(caller ZMQCaller, args Values) Values {
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return Empty
		}
		topic, err := args.ToString()
		if err != nil {
			return Empty
		}
		slaveInfo.SubScribedTopics.Set(topic, struct{}{})
		master.slaveSubscribedTopics.UnsafeGetAndUpdate(topic, func(val *sync_wrapper.SyncKVMap[string, chan Values]) *sync_wrapper.SyncKVMap[string, chan Values] {
			if val == nil {
				val = sync_wrapper.NewSyncKVMap[string, chan Values]()
			}
			val.Set(string(caller), slaveInfo.MsgToPub)
			return val
		})
		return Empty
	})
	server.ExposeAPI("/publish", func(caller ZMQCaller, args Values) Values {
		topic, err := args.ToString()
		if err != nil {
			return Empty
		}
		msg := args.ConsumeHead()
		master.publishMessage(string(caller), topic, msg)
		return Empty
	})
	server.ExposeAPI("/reg_api", func(provider ZMQCaller, args Values) Values {
		apiName, err := args.ToString()
		if err != nil {
			return Empty
		}
		slaveInfo, ok := master.slaves.Get(string(provider))
		if !ok {
			return FromString(ErrNotReg.Error())
		}
		// reject if api exist
		if master.LocalAPINode.HasAPI(apiName) {
			return FromString(ErrAPIExist.Error())
		}
		slaveInfo.ExposedApis.Set(apiName, struct{}{})
		master.ApiProvider.Set(apiName, string(provider))
		// master to slave call
		master.LocalAPINode.ExposeAPI(apiName, func(args Values) (result Values, err error) {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return Empty, fmt.Errorf("not found")
			} else {
				return unwrapOutput(server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResponse())
			}
		})
		// slave to salve call
		server.ExposeAPI(apiName, func(caller ZMQCaller, args Values) Values {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return Empty
			}
			return server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResponse()
		})
		return FromString("ok")
	})
	master.ExposeAPI("/set-value", func(args Values) (result Values, err error) {
		key, err := args.ToString()
		if err != nil {
			return Empty, err
		}
		master.values.Set(key, args.ConsumeHead())
		return Empty, nil
	})
	master.ExposeAPI("/get-value", func(args Values) (result Values, err error) {
		key, err := args.ToString()
		if err != nil {
			return Empty, err
		}
		head, ok := master.values.Get(key)
		if ok {
			return head, nil
		} else {
			return Empty, nil
		}
	})
	server.ExposeAPI("/set-tags", func(caller ZMQCaller, args Values) Values {
		tags := args.ToStrings()
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return FromString("need reg first")
		}
		for _, tag := range tags {
			slaveInfo.Tags.Set(tag, struct{}{})
		}
		return Empty
	})
	server.ExposeAPI("/check-tag", func(caller ZMQCaller, args Values) Values {
		tag, err := args.ToString()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		has := master.CheckNetTag(tag)
		return wrapOutput(FromBool(has), nil)
	})
	server.ExposeAPI("/try-lock", func(caller ZMQCaller, args Values) Values {
		lockName, err := args.ToString()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.tryLock(lockName, lockTime, string(caller))
		return wrapOutput(FromBool(has), nil)
	})
	server.ExposeAPI("/reset-lock-time", func(caller ZMQCaller, args Values) Values {
		lockName, err := args.ToString()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.resetLockTime(lockName, lockTime, string(caller))
		return wrapOutput(FromBool(has), nil)
	})
	server.ExposeAPI("/unlock", func(caller ZMQCaller, args Values) Values {
		lockName, err := args.ToString()
		if err != nil {
			return wrapOutput(Empty, err)
		}
		master.unlock(lockName, string(caller))
		return Empty
	})
	return master
}
