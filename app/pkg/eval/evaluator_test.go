package eval

import (
	"context"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/monogo/helpers"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
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

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory; %v", err)
	}
	evalDir, err := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "data", "eval"))
	if err != nil {
		t.Fatalf("Error getting eval directory; %v", err)
	}

	experiment := EvalExperiment{
		EvalDir:       evalDir,
		DBDir:         "/tmp/foyle/eval",
		GoogleSheetID: "1O0thD-p9DBF4G_shGMniivBB3pdaYifgSzWXBxELKqE",
		SheetName:     "Results",
	}
	if err := e.Reconcile(context.Background(), experiment); err != nil {
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

	experiment := EvalExperiment{
		EvalDir:       "",
		DBDir:         "/tmp/foyle/eval",
		GoogleSheetID: "1O0thD-p9DBF4G_shGMniivBB3pdaYifgSzWXBxELKqE",
		SheetName:     "Results",
	}
	db, err := pebble.Open(experiment.DBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("Error opening DB; %v", err)
	}
	defer helpers.DeferIgnoreError(db.Close)
	if err := e.updateGoogleSheet(context.Background(), experiment, db); err != nil {
		t.Fatalf("Error updating Google Sheet; %v", err)
	}
}
