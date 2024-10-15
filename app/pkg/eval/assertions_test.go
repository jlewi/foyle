package eval

import (
	"context"
	"testing"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

type testCase struct {
	name     string
	doc      *v1alpha1.Doc
	examples []*v1alpha1.Example
	answer   []*v1alpha1.Block
	expected *v1alpha1.Assertion
}

func TestAssertCodeAfterMarkdown(t *testing.T) {
	cases := []testCase{
		{
			name:     "Empty",
			doc:      &v1alpha1.Doc{},
			examples: []*v1alpha1.Example{},
			answer:   []*v1alpha1.Block{},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_CODE_AFTER_MARKDOWN,
				Result: v1alpha1.AssertResult_SKIPPED,
			},
		},
		{
			name: "Passed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_CODE,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_CODE_AFTER_MARKDOWN,
				Result: v1alpha1.AssertResult_PASSED,
			},
		},
		{
			name: "Passed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_MARKUP,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_CODE_AFTER_MARKDOWN,
				Result: v1alpha1.AssertResult_FAILED,
			},
		},
	}

	for _, c := range cases {
		a := &AssertCodeAfterMarkdown{}
		t.Run(c.name, func(t *testing.T) {
			got, err := a.Assert(context.Background(), c.doc, c.examples, c.answer)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if got.Result != c.expected.Result {
				t.Fatalf("Expected %v but got %v", c.expected.Result, got.Result)
			}
		})
	}
}

func TestAssertOneCodeCell(t *testing.T) {
	cases := []testCase{
		{
			name:     "Empty",
			doc:      &v1alpha1.Doc{},
			examples: []*v1alpha1.Example{},
			answer:   []*v1alpha1.Block{},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_CODE_AFTER_MARKDOWN,
				Result: v1alpha1.AssertResult_SKIPPED,
			},
		},
		{
			name: "Passed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_CODE,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_CODE_AFTER_MARKDOWN,
				Result: v1alpha1.AssertResult_PASSED,
			},
		},
		{
			name: "Failed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_MARKUP,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_ONE_CODE_CELL,
				Result: v1alpha1.AssertResult_FAILED,
			},
		},
	}

	for _, c := range cases {
		a := &AssertCodeAfterMarkdown{}
		t.Run(c.name, func(t *testing.T) {
			got, err := a.Assert(context.Background(), c.doc, c.examples, c.answer)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if got.Result != c.expected.Result {
				t.Fatalf("Expected %v but got %v", c.expected.Result, got.Result)
			}
		})
	}
}

func TestAssertEndsWithCodeCell(t *testing.T) {
	cases := []testCase{
		{
			name:     "Empty",
			doc:      &v1alpha1.Doc{},
			examples: []*v1alpha1.Example{},
			answer:   []*v1alpha1.Block{},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_ENDS_WITH_CODE_CELL,
				Result: v1alpha1.AssertResult_SKIPPED,
			},
		},
		{
			name: "Passed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_CODE,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_ENDS_WITH_CODE_CELL,
				Result: v1alpha1.AssertResult_PASSED,
			},
		},
		{
			name: "Failed",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind: v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			examples: []*v1alpha1.Example{},
			answer: []*v1alpha1.Block{
				{
					Kind: v1alpha1.BlockKind_CODE,
				},
				{
					Kind: v1alpha1.BlockKind_MARKUP,
				},
			},
			expected: &v1alpha1.Assertion{
				Name:   v1alpha1.Assertion_ENDS_WITH_CODE_CELL,
				Result: v1alpha1.AssertResult_FAILED,
			},
		},
	}

	for _, c := range cases {
		a := &AssertEndsWithCodeCell{}
		t.Run(c.name, func(t *testing.T) {
			got, err := a.Assert(context.Background(), c.doc, c.examples, c.answer)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if got.Result != c.expected.Result {
				t.Fatalf("Expected %v but got %v", c.expected.Result, got.Result)
			}
		})
	}
}
