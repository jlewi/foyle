package learn

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
	"google.golang.org/protobuf/encoding/protojson"
)

func Test_SortIndexes(t *testing.T) {
	type testCase struct {
		name   string
		input  []float64
		maxDim int
	}

	cases := []testCase{
		{
			name:   "simpl1",
			input:  []float64{3, 1, 2},
			maxDim: 3,
		},
		{
			name:   "simple2",
			input:  []float64{3, 2, 1},
			maxDim: 3,
		},
		{
			name:   "duplicates",
			input:  []float64{3, 2, 1, 2, 3, 7},
			maxDim: 6,
		},
		{
			// We want to make sure we only return the maxDim elements
			// In this case maxDim is 3 so the values input[4:6] should be ignored.
			// TO validate this we make the second half of the list be less than the first half.
			// If the returned indexes have a value less than max dim then sortIndexes considered values from the
			// second half.
			name:   "maxDim < length",
			input:  []float64{3, 2, 1, -1, -2, -3},
			maxDim: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := mat.NewVecDense(len(c.input), c.input)
			actualIndexes := sortIndexes(v, c.maxDim)

			if len(actualIndexes) != c.maxDim {
				t.Errorf("Expected %v indexes but got %v", c.maxDim, len(actualIndexes))
			}
			// Verify its sorted
			actual := make([]float64, len(actualIndexes))
			for i, idx := range actualIndexes {
				actual[i] = c.input[idx]

				// Make sure the index is less than maxDim
				if idx >= c.maxDim {
					t.Errorf("Expected index < %v but got %v", c.maxDim, idx)
				}
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

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing Viper; %v", err)
	}

	cfg := config.GetConfig()

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}

	vectorizer := oai.NewVectorizer(client)
	db, err := NewInMemoryExampleDB(*cfg, vectorizer)
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

func Test_UpdateExample(t *testing.T) {
	// Create some initial examples
	examples := []*v1alpha1.Example{
		{
			Id:        "1",
			Embedding: []float32{1, 2},
		},
		{
			Id:        "2",
			Embedding: []float32{3, 4},
		},
		{
			Id:        "3",
			Embedding: []float32{5, 6},
		},
		{
			Id:        "4",
			Embedding: []float32{7, 8},
		},
	}

	type testCase struct {
		name               string
		examples           []*v1alpha1.Example
		expectedMat        mat.Matrix
		expectedExampleIds []string
		// TODO(jeremy): We should check that the examples are properly associated with the indexes
	}

	cases := []testCase{
		{
			name:               "insert",
			examples:           examples,
			expectedExampleIds: []string{"1", "2", "3", "4"},
			expectedMat: mat.NewDense(4, 2, []float64{
				1, 2,
				3, 4,
				5, 6,
				7, 8,
			}),
		},
		{

			name: "overwrite",
			examples: []*v1alpha1.Example{
				examples[0],
				examples[1],
				// Now overwrite the examples
				{
					Id:        examples[1].Id,
					Embedding: []float32{9, 10},
				},
				{
					Id:        examples[0].Id,
					Embedding: []float32{100, 102},
				},
			},
			expectedExampleIds: []string{"1", "2"},
			expectedMat: mat.NewDense(2, 2, []float64{
				100, 102,
				9, 10,
			}),
		},
	}

	vLen := 2
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := &InMemoryExampleDB{
				examples:   make([]*v1alpha1.Example, 0),
				idToRow:    make(map[string]int),
				embeddings: mat.NewDense(int(float32(len(c.examples))/1.5), vLen, nil),
			}

			for _, e := range c.examples {
				if err := db.updateExample(e); err != nil {
					t.Fatalf("Error updating example; %v", err)
				}
			}

			actualIds := make([]string, len(db.examples))
			for i, e := range db.examples {
				actualIds[i] = e.Id
			}

			// Make sure each column is associated with the correct column
			if d := cmp.Diff(c.expectedExampleIds, actualIds); d != "" {
				t.Fatalf("Unexpected example ids; diff %v", d)
			}

			// Check # of columns because this shouldn't be changed during grow events.
			// We don't check that numExpectedRows db.embeddings.Dims()[0] because the matrix will likely have
			// extra rows as a result of preallocating rows for furture examples.
			_, actualColumns := db.embeddings.Dims()

			if actualColumns != vLen {
				t.Errorf("Expected %v rows but got %v", vLen, actualColumns)
			}

			for row := 0; row < len(c.expectedExampleIds); row++ {
				for col := 0; col < vLen; col++ {
					expected := c.expectedMat.At(row, col)
					actual := db.embeddings.At(row, col)
					if expected != actual {
						t.Errorf("embeddings[%d, %d]:expected %v but got %v", row, col, expected, actual)
					}
				}
			}

			// Ensure the reverse index is correct
			for id, row := range db.idToRow {
				actualId := db.examples[row].Id
				if id != actualId {
					t.Errorf("idToRow[%v]: expected %v but got %v", id, id, actualId)
				}
			}
		})
	}
}

func Test_initialNumberOfRows(t *testing.T) {
	type testCase struct {
		name        string
		numExamples int
		expected    int
	}

	cases := []testCase{
		{
			name:        "edge-case-1",
			numExamples: 1,
			expected:    1,
		},
		{
			name:        "small",
			numExamples: 10,
			expected:    6,
		},
		{
			name:        "large",
			numExamples: 100,
			expected:    66,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := initialNumberOfRows(c.numExamples)
			if actual != c.expected {
				t.Errorf("Expected %v but got %v", c.expected, actual)
			}
		})
	}
}
