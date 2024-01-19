package nodes

import "time"

func WaitUntilConflictNodeOffline(node Node, tag string, maxWaitTime time.Duration) bool {
	deadlineTime := time.Now().Add(maxWaitTime)
	for node.CheckNetTag(tag) {
		time.Sleep(time.Second / 2)
		if time.Now().After(deadlineTime) {
			return false
		}
	}
	return true
}
