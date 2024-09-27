package eval

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"google.golang.org/protobuf/encoding/protojson"
)

func Test_protoToRowUpdate(t *testing.T) {

	type testCase struct {
		name   string
		result *v1alpha1.EvalResult
	}

	cases := []testCase{
		{
			name: "Basic",
			result: &v1alpha1.EvalResult{
				Example: &v1alpha1.EvalExample{
					Id: "1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := protoToRowUpdate(c.result)
			if err != nil {
				t.Fatalf("Error converting EvalResult to row update: %v", err)
			}

			if actual.ID != c.result.Example.Id {
				t.Fatalf("Expected ID to be %v but got %v", c.result.Example.Id, actual.ID)
			}

			if actual.Time != c.result.Example.Time.AsTime() {
				t.Fatalf("Expected Time to be %v but got %v", c.result.Example.Time, actual.Time)
			}

			// We can't compare the serialized protos because the JSON serialization is non-deterministic

			actualPB := &v1alpha1.EvalResult{}
			if err := protojson.Unmarshal([]byte(actual.ProtoJson), actualPB); err != nil {
				t.Fatalf("Error deserializing actual result: %v", err)
			}

			comparer := cmpopts.IgnoreUnexported(v1alpha1.EvalResult{}, v1alpha1.EvalExample{}, time.Time{})
			if d := cmp.Diff(c.result, actualPB, comparer); d != "" {
				t.Fatalf("Unexpected diff between expected and actual EvalResults:\n%v", d)
			}
		})
	}
}
