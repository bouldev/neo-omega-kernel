package nodes

import (
	"time"

	"github.com/OmineDev/neomega-core/nodes/defines"
)

func WaitUntilConflictNodeOffline(node defines.Node, tag string, maxWaitTime time.Duration) bool {
	deadlineTime := time.Now().Add(maxWaitTime)
	for node.CheckNetTag(tag) {
		time.Sleep(time.Second / 2)
		if time.Now().After(deadlineTime) {
			return false
		}
	}
	return true
}
