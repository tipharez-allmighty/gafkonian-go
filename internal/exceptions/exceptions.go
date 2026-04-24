// Package exceptions provides specific error code and messages for the project
package exceptions

import "fmt"

const (
	UnkownProtocolError        int16 = -1
	UnsupportedAPIVersionError int16 = 35
	UnsupportedAPIKeyError     int16 = 36
	InsufficientHeaderError    int16 = 37
)

var ErrorMessages = map[int16]string{
	UnkownProtocolError:        "unknown protocol error",
	UnsupportedAPIVersionError: "unsupported API version %v",
	UnsupportedAPIKeyError:     "unsupported API key %v",
	InsufficientHeaderError:    "insufficient data for header: %v < 8",
}

type ProtocolError struct {
	Code    int16
	Message string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("%v-%v", e.Code, e.Message)
}

func RaiseError(code int16, args ...any) *ProtocolError {
	errorMessage, ok := ErrorMessages[code]
	if !ok {
		code = -1
		errorMessage = ErrorMessages[code]
	}
	return &ProtocolError{
		Code:    code,
		Message: fmt.Sprintf(errorMessage, args...),
	}
}
