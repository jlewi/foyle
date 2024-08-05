package eval

import (
	"context"
	"fmt"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

const (
	CodeAfterMarkdownName = "AssertCodeAfterMarkdown"
	OneCodeCellName       = "AssertOneCodeCell"
	EndsWithCodeCellName  = "AssertEndsWithCodeCell"
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

// AssertOneCodeCell is an assertion that checks the response has one code cell.
// We don't want to suggest multiple code cells because that can be confusing. If one command depends on the
// output of another we should generate one at a time.
type AssertOneCodeCell struct {
}

func (a *AssertOneCodeCell) Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (*v1alpha1.Assertion, error) {
	assertion := &v1alpha1.Assertion{
		Name: a.Name(),
	}

	if len(doc.Blocks) == 0 {
		assertion.Result = v1alpha1.AssertResult_FAILED
		return assertion, nil
	}

	numCode := 0
	for _, b := range answer {
		if b.GetKind() == v1alpha1.BlockKind_CODE {
			numCode++
		}
	}

	if numCode != 1 {
		assertion.Result = v1alpha1.AssertResult_FAILED
		assertion.Detail = fmt.Sprintf("Answer doesn't contain exactly one code cell; it has %d code cells ", numCode)
		return assertion, nil
	}

	assertion.Result = v1alpha1.AssertResult_PASSED
	return assertion, nil
}

func (a *AssertOneCodeCell) Name() string {
	return OneCodeCellName
}

// AssertEndsWithCodeCell is an assertion that checks that if the response has a code cell then it ends with the
// code cell. If we need the user to run a command we should just suggest the command and not additional output
// after that.
type AssertEndsWithCodeCell struct {
}

func (a *AssertEndsWithCodeCell) Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (*v1alpha1.Assertion, error) {
	assertion := &v1alpha1.Assertion{
		Name: a.Name(),
	}

	if len(doc.Blocks) == 0 {
		assertion.Result = v1alpha1.AssertResult_SKIPPED
		assertion.Detail = "Doc is empty"
		return assertion, nil
	}

	hasCodeCell := false
	for _, b := range answer {
		if b.GetKind() == v1alpha1.BlockKind_CODE {
			hasCodeCell = true
			break
		}
	}

	if !hasCodeCell {
		assertion.Result = v1alpha1.AssertResult_SKIPPED
		assertion.Detail = "Answer doesn't contain a code cell"
		return assertion, nil
	}

	last := answer[len(answer)-1]

	if last.GetKind() != v1alpha1.BlockKind_CODE {
		assertion.Result = v1alpha1.AssertResult_FAILED
		return assertion, nil
	}

	assertion.Result = v1alpha1.AssertResult_PASSED
	return assertion, nil
}

func (a *AssertEndsWithCodeCell) Name() string {
	return EndsWithCodeCellName
}
