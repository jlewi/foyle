package learn

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
	"os"
	"testing"
)

func Test_Learner(t *testing.T) {
	// This isn't really a test because it depends on your configuration and logs.
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

	l, err := NewLearner(cfg)
	if err != nil {
		t.Fatalf("Error creating learner; %v", err)
	}

	if err := l.Reconcile(context.Background()); err != nil {
		t.Fatalf("Error reconciling; %v", err)
	}
}
