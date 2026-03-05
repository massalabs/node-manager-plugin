package nodeManager

import (
	"errors"
	"runtime"
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
		// On Unix-like systems, ECONNREFUSED is the standard way to signal
		// "connection refused".
		if errno == syscall.ECONNREFUSED {
			return true
		}

		// On Windows, network errors come from Winsock (WSAECONNREFUSED = 10061),
		// and Go wraps them as syscall.Errno without mapping them to ECONNREFUSED.
		// Detect that numeric code explicitly so we correctly treat it as
		// "connection refused" while the node API is not yet ready.
		if runtime.GOOS == "windows" && errno == syscall.Errno(10061) {
			return true
		}
	}

	return false
}
