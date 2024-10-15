package eval

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// TODO(jeremy): A lot of this code is probably obsolete now that we are using protos.

// Assertion is an interface for evaluating AI generations.
type Assertion interface {
	Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (*v1alpha1.Assertion, error)
	// Name returns the name of the assertion.
	Name() string
}

type AssertResult string
