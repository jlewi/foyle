package eval

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

const (
	CodeAfterMarkdownName = "AssertCodeAfterMarkdown"
)

// AssertCodeAfterMarkdown is an assertion that checks that if the prompt ends in a markdown cell then the response
// starts with a code cell.
type AssertCodeAfterMarkdown struct {
}

func (a *AssertCodeAfterMarkdown) Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (*v1alpha1.Assertion, error) {
	assertion := &v1alpha1.Assertion{
		Name: a.Name(),
	}

	if len(doc.Blocks) == 0 {
		assertion.Result = v1alpha1.AssertResult_SKIPPED
		return assertion, nil
	}

	last := doc.Blocks[len(doc.Blocks)-1]
	if last.GetKind() != v1alpha1.BlockKind_MARKUP {
		assertion.Result = v1alpha1.AssertResult_SKIPPED
		return assertion, nil
	}

	if len(answer) == 0 {
		assertion.Result = v1alpha1.AssertResult_FAILED
		assertion.Detail = "Answer is empty"
		return assertion, nil
	}

	if answer[0].GetKind() != v1alpha1.BlockKind_CODE {
		assertion.Result = v1alpha1.AssertResult_FAILED
		assertion.Detail = "Answer doesn't start with a code cell"
		return assertion, nil
	}

	assertion.Result = v1alpha1.AssertResult_PASSED
	return assertion, nil
}

func (a *AssertCodeAfterMarkdown) Name() string {
	return CodeAfterMarkdownName
}
