//go:build windows

package nodeManager

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func isUserInterrupted(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus() == windows.CTRL_C_EVENT ||
				status.ExitStatus() == windows.CTRL_BREAK_EVENT
		}
	}
	return false
}
