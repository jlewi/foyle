package eval

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jlewi/foyle/app/pkg/analyze"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func Test_ListResults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Test_ListResults")
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}

	dbFile := filepath.Join(tempDir, "results.db")

	db, err := sql.Open(analyze.SQLLiteDriver, dbFile)
	if err != nil {
		t.Fatalf("Error creating database: %v", err)
	}

	m, err := NewResultsManager(db)
	if err != nil {
		t.Fatalf("Error creating ResultsManager: %v", err)
	}

	// Try listing the results when there are no results
	rows, _, err := m.ListResults(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("Error listing results: %v", err)
	}

	if len(rows) != 0 {
		t.Fatalf("Expected no results but got %v", len(rows))
	}

	// Now insert some rows.
	baseTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	results := []*v1alpha1.EvalResult{
		{
			Example: &v1alpha1.EvalExample{
				Id:   "1",
				Time: timestamppb.New(baseTime),
			},
		},
		{
			Example: &v1alpha1.EvalExample{
				Id:   "2",
				Time: timestamppb.New(baseTime.Add(time.Hour)),
			},
		},
		{
			Example: &v1alpha1.EvalExample{
				Id:   "3",
				Time: timestamppb.New(baseTime.Add(-1 * time.Hour)),
			},
		},
	}

	for _, r := range results {
		uErr := m.Update(context.Background(), r.Example.Id, func(result *v1alpha1.EvalResult) error {
			proto.Merge(result, r)
			return nil
		})
		if uErr != nil {
			t.Fatalf("Error inserting result: %v", err)
		}
	}

	// List the results
	rows, cursor, err := m.ListResults(context.Background(), nil, 10)

	if err != nil {
		t.Fatalf("Error listing results: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("Expected 3 results but got %v", len(rows))
	}

	if !isSortedByTimeDescending(rows) {
		t.Fatalf("Results are not sorted by time")
	}

	expected := baseTime.Add(-1 * time.Hour)
	if *cursor != baseTime.Add(-1*time.Hour) {
		t.Fatalf("Cursor is invalid; Got %v; Want %v", *cursor, expected)
	}
}
