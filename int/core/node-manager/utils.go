package nodeManager

import (
	"errors"
	"runtime"
	"syscall"

	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
	"github.com/massalabs/station/pkg/logger"
)

func IsRunning(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return nodeStatus != nodeStatusPkg.NodeStatusOff && nodeStatus != nodeStatusPkg.NodeStatusCrashed
}

func IsClosedOrClosing(nodeStatus nodeStatusPkg.NodeStatus) bool {
	return !IsRunning(nodeStatus) || nodeStatus == nodeStatusPkg.NodeStatusStopping
}

func connRefused(err error) bool {
	logger.Debug("DEBUG bootstrapping err: %v", err)
	var errno syscall.Errno

	if errors.As(err, &errno) {
		logger.Debug("DEBUG errno: %d", errno)
		switch runtime.GOOS {
		case "linux":
			return errno == 111
		case "darwin":
			return errno == 3260
		case "windows":
			return errno == 10061
		}
	}

	return false
}
