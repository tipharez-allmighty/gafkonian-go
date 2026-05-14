// Package error provides specific error code and messages for the project
package error

import "fmt"

const (
	SystemError                int16 = 0
	UnkownProtocolError        int16 = -1
	UnsupportedAPIVersionError int16 = 35
	UnsupportedAPIKeyError     int16 = 36
	InsufficientHeaderError    int16 = 37
	InsufficientBodyError      int16 = 38
	EmptyTopicsBodyError       int16 = 39
	TopicNotFoundError         int16 = 40
	PartitionNotFoundError     int16 = 41
)

var ErrorMessages = map[int16]string{
	SystemError:                "system error: %v",
	UnkownProtocolError:        "unknown protocol error",
	UnsupportedAPIVersionError: "unsupported API version %v",
	UnsupportedAPIKeyError:     "unsupported API key %v",
	InsufficientHeaderError:    "insufficient data for header: %v < 8",
	InsufficientBodyError:      "insufficient data for body: %v < 4",
	EmptyTopicsBodyError:       "there is no topics to describe",
	TopicNotFoundError:         "unknown topic",
	PartitionNotFoundError:     "unknown partition",
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
