package docs

import (
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"testing"
)

func Test_CreateQuery(t *testing.T) {
	doc1 := &v1alpha1.Doc{
		Blocks: []*v1alpha1.Block{
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 0",
			},
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 1",
			},
			{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: "cell 2",
			},
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 3",
			},
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 4",
			},
		},
	}

	type testCase struct {
		name     string
		input    *v1alpha1.GenerateRequest
		expected []*v1alpha1.Block
	}

	cases := []testCase{
		{
			name: "stop-at-start",
			input: &v1alpha1.GenerateRequest{
				Doc:           doc1,
				SelectedIndex: 1,
			},
			expected: doc1.Blocks[0:2],
		},
		{
			name: "start-on-codeblock",
			input: &v1alpha1.GenerateRequest{
				Doc:           doc1,
				SelectedIndex: 2,
			},
			expected: doc1.Blocks[0:3],
		},
		{
			name: "stop-on-code",
			input: &v1alpha1.GenerateRequest{
				Doc:           doc1,
				SelectedIndex: 4,
			},
			expected: doc1.Blocks[3:5],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			blocks, err := CreateQuery(nil, tc.input)
			if err != nil {
				t.Fatalf("CreateQuery failed: %v", err)
			}
			if len(blocks) != len(tc.expected) {
				t.Errorf("CreateQuery returned %d blocks; want %d", len(blocks), len(tc.expected))
			}

			if d := cmp.Diff(tc.expected, blocks, testutil.DocComparer, testutil.BlockComparer); d != "" {
				t.Errorf("CreateQuery returned unexpected blocks:\n%v", d)
			}

		})
	}
}
