package nodeManager

import "strings"

func IsRunning(nodeStatus NodeStatus) bool {
	return nodeStatus == NodeStatusOn || nodeStatus == NodeStatusBootstrapping || nodeStatus == NodeStatusStopping
}

func connRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connect: connection refused")
}
