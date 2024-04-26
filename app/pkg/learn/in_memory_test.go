package learn

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"sort"
	"testing"
)

func Test_SortIndexes(t *testing.T) {
	type testCase struct {
		name  string
		input []float64
	}

	cases := []testCase{
		{
			name:  "simpl1",
			input: []float64{3, 1, 2},
		},
		{
			name:  "simple2",
			input: []float64{3, 2, 1},
		},
		{
			name:  "duplicates",
			input: []float64{3, 2, 1, 2, 3, 7},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := mat.NewVecDense(len(c.input), c.input)
			actualIndexes := sortIndexes(v)

			// Verify its sorted
			actual := make([]float64, len(c.input))
			for i, idx := range actualIndexes {
				actual[i] = c.input[idx]
			}

			if !sort.Float64sAreSorted(actual) {
				t.Errorf("Expected sorted but got %v", actual)
			}
		})
	}
}

func Test_InMemoryDB(t *testing.T) {
	// This isn't really a unittest; it depends on your configuration and examples.
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	log, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)

	config.InitViper(nil)
	cfg := config.GetConfig()

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}

	db, err := NewInMemoryExampleDB(*cfg, client)
	if err != nil {
		t.Fatalf("Error creating learner; %v", err)
	}

	doc := &v1alpha1.Doc{
		Blocks: []*v1alpha1.Block{
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "Compute a diff with branch main and then summarize the changes to be a PR description",
			},
		},
	}
	examples, err := db.GetExamples(context.Background(), doc, 1)
	if err != nil {
		t.Fatalf("Error getting examples; %v", err)
	}

	if len(examples) == 0 {
		t.Fatalf("Expected at least one example")
	}

	for _, e := range examples {
		m := protojson.MarshalOptions{
			Indent: "  ",
		}
		result, err := m.Marshal(e)
		if err != nil {
			t.Fatalf("Failed to marshal example; %v", err)
		}
		t.Logf("Example:\n%v", string(result))
	}
}
