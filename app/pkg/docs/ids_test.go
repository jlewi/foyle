package docs

import (
	"testing"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_SetBlockIds(t *testing.T) {
	type testCase struct {
		name   string
		blocks []*v1alpha1.Block
	}

	testCases := []testCase{
		{
			name: "Empty",
			blocks: []*v1alpha1.Block{
				{
					Contents: "Hello",
				},
				{
					Contents: "",
				},
				{
					Contents: "",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blockids, err := SetBlockIds(tc.blocks)
			if err != nil {
				t.Fatalf("SetBlockIds failed; error: %v", err)
			}
			// Make sure we got back n Unique
			unique := make(map[string]bool)
			for _, id := range blockids {
				unique[id] = true
			}

			if len(unique) != len(tc.blocks) {
				t.Fatalf("Expected %d unique ids but got %d", len(tc.blocks), len(unique))
			}
		})
	}
}
