package nodeManager

import (
	"strings"

	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
)

func IsRunning(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return nodeStatus != nodeStatusPkg.NodeStatusOff && nodeStatus != nodeStatusPkg.NodeStatusCrashed
}

func IsClosedOrClosing(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return !IsRunning(nodeStatus) || nodeStatus == nodeStatusPkg.NodeStatusStopping
}

func connRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connect: connection refused")
}
