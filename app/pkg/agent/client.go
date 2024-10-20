package agent

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
)

// RetryInterceptor defines a retry interceptor
type RetryInterceptor struct {
	MaxRetries int
	Backoff    time.Duration
}

func (r *RetryInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	// We return a function that will wrap the next function call in a try loop.
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		var resp connect.AnyResponse
		var err error

		for i := 0; i <= r.MaxRetries; i++ {
			resp, err = next(ctx, req)

			// If no error, return the response
			if err == nil {
				return resp, nil
			}

			// Check if the error is a DeadlineExceeded or Cancelled
			// Check if the error is a DeadlineExceeded or Canceled
			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				code := connectErr.Code()
				if code == connect.CodeDeadlineExceeded || code == connect.CodeCanceled {
					// Delay before retrying
					time.Sleep(r.Backoff)
					continue
				}
			}

			// For other errors, return immediately
			return nil, err
		}

		// After max retries, return the last error
		return nil, err
	}
}

// WrapStreamingClient implements [Interceptor] with a no-op.
func (r *RetryInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	// TODO(jeremy): We should implement this
	return next
}

// WrapStreamingHandler implements [Interceptor] with a no-op.
func (r *RetryInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	// TODO(jeremy): We should implement this
	return next
}
