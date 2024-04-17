package docs

import (
	"fmt"
	"testing"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_GetExitCode(t *testing.T) {
	type testCase struct {
		input        v1alpha1.BlockOutput
		expectedCode int
		expectedOk   bool
	}

	cases := []testCase{
		{
			input: v1alpha1.BlockOutput{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: "exitCode: 5",
					},
				},
			},
			expectedCode: 5,
			expectedOk:   true,
		},
		{
			input: v1alpha1.BlockOutput{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: "exitCode:7 ",
					},
				},
			},
			expectedCode: 7,
			expectedOk:   true,
		},
		{
			input: v1alpha1.BlockOutput{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: "other stuff",
					},
				},
			},
			expectedCode: 1,
			expectedOk:   false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			code, ok := GetExitCode(c.input)
			if code != c.expectedCode {
				t.Errorf("Expected code %d but got %d", c.expectedCode, code)
			}
			if ok != c.expectedOk {
				t.Errorf("Expected ok %t but got %t", c.expectedOk, ok)
			}
		})
	}
}
