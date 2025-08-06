package error

import (
	"fmt"
)

// NodeManagerErrorCode represents different types of errors in the node manager
type NodeManagerErrorCode string

const (
	// Generic errors
	ErrUnknown NodeManagerErrorCode = "UNKNOWN_ERROR"

	// DB
	ErrDBNotFoundItem NodeManagerErrorCode = "DB_NOT_FOUND_ITEM"

	// Staking Manager
	ErrStakingManagerPendingOperationNotCompleted NodeManagerErrorCode = "STAKING_MANAGER_PENDING_OPERATION_NOT_COMPLETED"
)

// NodeManagerError represents a structured error in the node manager
type NodeManagerError struct {
	Message    string                 `json:"message"`
	Code       NodeManagerErrorCode   `json:"code"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Error implements the error interface
func (e *NodeManagerError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code NodeManagerErrorCode, message string) *NodeManagerError {
	return &NodeManagerError{
		Code:       code,
		Message:    message,
		Parameters: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with a NodeManagerError
func Wrap(err error, code NodeManagerErrorCode, message string) *NodeManagerError {
	if err == nil {
		return &NodeManagerError{
			Code:       code,
			Message:    message,
			Parameters: make(map[string]interface{}),
		}
	}

	// If the error is already a NodeManagerError, just update the message
	if nmErr, ok := err.(*NodeManagerError); ok {
		return &NodeManagerError{
			Code:       nmErr.Code,
			Message:    fmt.Sprintf("%s: %s", message, nmErr.Message),
			Parameters: nmErr.Parameters,
		}
	}

	return &NodeManagerError{
		Code:       code,
		Message:    fmt.Sprintf("%s: %s", message, err.Error()),
		Parameters: make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with a formatted message
func Wrapf(err error, code NodeManagerErrorCode, format string, args ...interface{}) *NodeManagerError {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Is checks if the error is of a specific type
func Is(err error, code NodeManagerErrorCode) bool {
	if nmErr, ok := err.(*NodeManagerError); ok {
		return nmErr.Code == code
	}
	return false
}

// GetCode returns the error code from an error
func GetCode(err error) NodeManagerErrorCode {
	if nmErr, ok := err.(*NodeManagerError); ok {
		return nmErr.Code
	}
	return ErrUnknown
}

// GetMessage returns the error message from an error
func GetMessage(err error) string {
	if nmErr, ok := err.(*NodeManagerError); ok {
		return nmErr.Message
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetParameters returns the parameters from an error
func GetParameters(err error) map[string]interface{} {
	if nmErr, ok := err.(*NodeManagerError); ok {
		return nmErr.Parameters
	}
	return nil
}

// AddParameter adds a parameter to the error
func (e *NodeManagerError) AddParameter(key string, value interface{}) {
	if e.Parameters == nil {
		e.Parameters = make(map[string]interface{})
	}
	e.Parameters[key] = value
}

// GetParameter gets a parameter from the error
func (e *NodeManagerError) GetParameter(key string) (interface{}, bool) {
	if e.Parameters == nil {
		return nil, false
	}
	value, exists := e.Parameters[key]
	return value, exists
}
