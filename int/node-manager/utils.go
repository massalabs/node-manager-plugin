package nodeManager

import "strings"

func IsRunning(nodeStatus NodeStatus) bool {
	return nodeStatus != NodeStatusOff && nodeStatus != NodeStatusCrashed
}

func IsClosedOrClosing(nodeStatus NodeStatus) bool {
	return !IsRunning(nodeStatus) || nodeStatus == NodeStatusStopping
}

func connRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connect: connection refused")
}

func isMainnetFromVersion(version string) bool {
	return strings.Contains(version, "MAIN")
}
