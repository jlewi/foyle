package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jlewi/foyle/app/api"
	"github.com/pkg/errors"

	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
)

func Test_Evaluator(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	t.Fatalf("Evaluator test needs to be updated per https://github.com/jlewi/foyle/issues/140")

	log, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing Viper; %v", err)
	}
	cfg := config.GetConfig()

	e, err := NewEvaluator(*cfg)
	if err != nil {
		t.Fatalf("Error creating evaluator; %v", err)
	}

	experiment, err := experimentForTesting()
	if err != nil {
		t.Fatalf("Error creating experiment; %v", err)
	}

	if err := e.Reconcile(context.Background(), *experiment); err != nil {
		t.Fatalf("Error reconciling; %v", err)
	}
}

//func Test_Evaluator_Google_Sheets(t *testing.T) {
//	if os.Getenv("GITHUB_ACTIONS") != "" {
//		t.Skipf("Test is skipped in GitHub actions")
//	}
//
//	t.Fatalf("Evaluator test needs to be updated per https://github.com/jlewi/foyle/issues/140")
//
//	log, err := zap.NewDevelopmentConfig().Build()
//	if err != nil {
//		t.Fatalf("Error creating logger; %v", err)
//	}
//	zap.ReplaceGlobals(log)
//
//	if err := config.InitViper(nil); err != nil {
//		t.Fatalf("Error initializing Viper; %v", err)
//	}
//	cfg := config.GetConfig()
//
//	e, err := NewEvaluator(*cfg)
//	if err != nil {
//		t.Fatalf("Error creating evaluator; %v", err)
//	}
//
//	experiment, err := experimentForTesting()
//	if err != nil {
//		t.Fatalf("Error creating experiment; %v", err)
//	}
//
//	db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
//	if err != nil {
//		t.Fatalf("Error opening DB; %v", err)
//	}
//	defer helpers.DeferIgnoreError(db.Close)
//	//if err := e.updateGoogleSheet(context.Background(), *experiment, db); err != nil {
//	//	t.Fatalf("Error updating Google Sheet; %v", err)
//	//}
//}

func experimentForTesting() (*api.Experiment, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting working directory")
	}
	evalDir, err := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "data", "eval"))
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting eval directory")
	}

	return &api.Experiment{
		Spec: api.ExperimentSpec{
			EvalDir:   evalDir,
			DBDir:     "/tmp/foyle/eval",
			SheetID:   "1O0thD-p9DBF4G_shGMniivBB3pdaYifgSzWXBxELKqE",
			SheetName: "Results",
			Agent: &api.AgentConfig{
				Model: config.DefaultModel,
				// No need to test RAG as part of testing evaluation.
				RAG: &api.RAGConfig{
					Enabled: false,
				},
			},
		},
	}, nil
}

//func Test_updateEvalResultDistance(t *testing.T) {
//	type testCase struct {
//		name               string
//		result             *v1alpha1.EvalResult
//		expectedDistance   int32
//		expectedNormalized float32
//	}
//
//	cases := []testCase{
//		{
//			// Test the case where the actual answer contains no codeblocks
//			name: "nocodeblocks",
//			result: &v1alpha1.EvalResult{
//				Example: &v1alpha1.EvalExample{
//					Id: "1234",
//					Answer: []*v1alpha1.Block{
//						{
//							Kind:     v1alpha1.BlockKind_CODE,
//							Contents: "gcloud builds list",
//						},
//					},
//				},
//				ExampleFile: "",
//				Actual: []*v1alpha1.Block{
//					{
//						Kind:     v1alpha1.BlockKind_MARKUP,
//						Contents: "Not a code cell",
//					},
//				},
//			},
//			expectedDistance:   3,
//			expectedNormalized: 1.0,
//		},
//	}
//	parser, err := executor.NewBashishParser()
//	if err != nil {
//		t.Fatalf("Error creating parser; %v", err)
//	}
//
//	for _, c := range cases {
//		t.Run(c.name, func(t *testing.T) {
//			updateEvalResultDistance(context.Background(), parser, c.result)
//			if err != nil {
//				t.Fatalf("Unexpected error: %v", err)
//			}
//			if c.result.Distance != c.expectedDistance {
//				t.Errorf("Expected distance %d but got %d", c.expectedDistance, c.result.Distance)
//			}
//			if c.result.NormalizedDistance != c.expectedNormalized {
//				t.Errorf("Expected normalized distance %f but got %f", c.expectedNormalized, c.result.NormalizedDistance)
//			}
//		})
//	}
//}
