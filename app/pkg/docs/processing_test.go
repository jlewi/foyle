package docs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_peprocessDoc(t *testing.T) {
	type testCase struct {
		name     string
		input    *v1alpha1.GenerateRequest
		expected []*v1alpha1.Block
	}

	doc := &v1alpha1.Doc{
		Blocks: []*v1alpha1.Block{
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 0",
			},
			{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: "cell 1",
			},
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 2",
			},
		},
	}

	docWithGhost := &v1alpha1.Doc{
		Blocks: []*v1alpha1.Block{
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 0",
			},
			{
				Kind: v1alpha1.BlockKind_CODE,
				Metadata: map[string]string{
					converters.GhostKeyField: "true",
				},
				Contents: "cell 1",
			},
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "cell 2",
			},
		},
	}

	cases := []testCase{
		{
			name: "basic",
			input: &v1alpha1.GenerateRequest{
				Doc:           doc,
				SelectedIndex: 0,
			},
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "cell 0",
				},
			},
		},
		{
			name: "middle",
			input: &v1alpha1.GenerateRequest{
				Doc:           doc,
				SelectedIndex: 1,
			},
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "cell 0",
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Contents: "cell 1",
				},
			},
		},
		{
			name: "ghost",
			input: &v1alpha1.GenerateRequest{
				Doc:           docWithGhost,
				SelectedIndex: 2,
			},
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "cell 0",
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "cell 2",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := PreprocessDoc(c.input)

			opts := cmpopts.IgnoreUnexported(v1alpha1.Block{})
			if d := cmp.Diff(c.expected, actual, opts); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
