package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_Executor(t *testing.T) {
	type testCase struct {
		req      *v1alpha1.ExecuteRequest
		expected *v1alpha1.ExecuteResponse
	}

	cases := []testCase{
		{
			req: &v1alpha1.ExecuteRequest{Block: &v1alpha1.Block{Contents: "echo \"something something\""}},
			expected: &v1alpha1.ExecuteResponse{
				Outputs: []*v1alpha1.BlockOutput{
					{
						Items: []*v1alpha1.BlockOutputItem{
							{
								Mime:     MimePlainText,
								TextData: "exitCode: 0",
							},
						},
					},
					{
						Items: []*v1alpha1.BlockOutputItem{
							{
								Mime:     MimePlainText,
								TextData: "stdout:\nsomething something",
							},
						},
					},
				},
			},
		},
	}

	e, err := NewExecutor()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Case %v", i), func(t *testing.T) {

			resp, err := e.Execute(context.Background(), c.req)
			if err != nil {
				t.Fatalf("Failed to execute: %v", err)
			}
			if d := cmp.Diff(c.expected, resp, cmpopts.IgnoreUnexported(v1alpha1.ExecuteResponse{}), cmpopts.IgnoreUnexported(v1alpha1.BlockOutput{}), cmpopts.IgnoreUnexported(v1alpha1.BlockOutputItem{})); d != "" {
				t.Errorf("Unexpected response (-want +got):\n%v", d)
			}
		})
	}
}
