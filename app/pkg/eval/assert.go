package eval

import (
	"context"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// Assertion is an interface for evaluating AI generations.
type Assertion interface {
	Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (AssertResult, error)
	// Name returns the name of the assertion.
	Name() string
}

type AssertResult string

const AssertPassed AssertResult = "passed"
const AssertFailed AssertResult = "failed"
const AssertSkipped AssertResult = "skipped"
