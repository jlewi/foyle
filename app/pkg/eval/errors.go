package eval

import "strings"

func isContextLengthExceeded(err error) bool {
	// TODO(jeremy): This is a hacky way of deciding when completion fails because the context length is exceeded.
	// We should have the server propogate better status codes.
	// This is for open ai
	if strings.Contains(err.Error(), "Please reduce the length of the messages") {
		return true
	}
	return false
}
