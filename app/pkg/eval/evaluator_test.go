package eval

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
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

	outDir := "/tmp/foyle/eval"
	if err := e.Reconcile(context.Background(), evalDir, outDir); err != nil {
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

	if err := e.updateGoogleSheet(context.Background()); err != nil {
		t.Fatalf("Error updating Google Sheet; %v", err)
	}
}
