package eval

import (
	"context"
	"github.com/jlewi/foyle/app/api"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/monogo/helpers"
	"go.uber.org/zap"
)

func Test_Evaluator(t *testing.T) {
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

func Test_Evaluator_Google_Sheets(t *testing.T) {
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

	e, err := NewEvaluator(*cfg)
	if err != nil {
		t.Fatalf("Error creating evaluator; %v", err)
	}

	experiment, err := experimentForTesting()
	if err != nil {
		t.Fatalf("Error creating experiment; %v", err)
	}

	db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("Error opening DB; %v", err)
	}
	defer helpers.DeferIgnoreError(db.Close)
	if err := e.updateGoogleSheet(context.Background(), *experiment, db); err != nil {
		t.Fatalf("Error updating Google Sheet; %v", err)
	}
}

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
