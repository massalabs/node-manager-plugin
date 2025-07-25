//go:build !windows

package nodeManager

import (
	"os/exec"
	"syscall"
)

func isUserInterrupted(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return ws.Signaled() &&
				(ws.Signal() == syscall.SIGTERM || ws.Signal() == syscall.SIGINT || ws.Signal() == syscall.SIGQUIT)
		}
	}
	return false
}
