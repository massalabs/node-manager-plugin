package nodeManager

import (
	"errors"
	"runtime"
	"syscall"

	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
	"golang.org/x/sys/unix"
)

func IsRunning(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return nodeStatus != nodeStatusPkg.NodeStatusOff && nodeStatus != nodeStatusPkg.NodeStatusCrashed
}

func IsClosedOrClosing(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return !IsRunning(nodeStatus) || nodeStatus == nodeStatusPkg.NodeStatusStopping
}

func connRefused(err error) bool {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		var errno syscall.Errno
		if errors.As(err, &errno) && errno == unix.ECONNREFUSED {
			return true
		}
	}
	return false
}
