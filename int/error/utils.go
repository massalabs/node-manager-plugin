package error

import (
	"errors"
	"syscall"
)

// closing zap logger (used by station logger) can return an "invalid argument" error that should be ignored
// See: https://github.com/uber-go/zap/issues/772, https://github.com/uber-go/zap/issues/1093
func IsZapLoggerInvalidArgumentError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) && errno == syscall.EINVAL {
		return true
	}
	return false
}
