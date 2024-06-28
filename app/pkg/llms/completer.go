package llms

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// Completer is an interface for generating completions
type Completer interface {
	Complete(ctx context.Context, systemPrompt string, message string) ([]*v1alpha1.Block, error)
}
