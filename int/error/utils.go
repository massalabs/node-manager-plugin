package error

import "strings"

// closing zap logger (used by station logger) can return an "invalid argument" error that should be ignored
// See: https://github.com/uber-go/zap/issues/772, https://github.com/uber-go/zap/issues/1093
func IsZapLoggerInvalidArgumentError(err error) bool {
	return strings.Contains(err.Error(), "invalid argument")
}
