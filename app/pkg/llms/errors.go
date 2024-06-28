package llms

type ContextLengthExceededError struct {
	Cause error
}

func (e ContextLengthExceededError) Error() string {
	return e.Cause.Error()
}
