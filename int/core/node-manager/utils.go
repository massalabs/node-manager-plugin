package nodeManager

import (
	"errors"
	"syscall"

	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
)

func IsRunning(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return nodeStatus != nodeStatusPkg.NodeStatusOff && nodeStatus != nodeStatusPkg.NodeStatusCrashed
}

func IsClosedOrClosing(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return !IsRunning(nodeStatus) || nodeStatus == nodeStatusPkg.NodeStatusStopping
}

func connRefused(err error) bool {
	var errno syscall.Errno

	if errors.As(err, &errno) {
		return errno == syscall.ECONNREFUSED
	}

	return false
}
