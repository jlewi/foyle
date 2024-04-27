package learn

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"go.uber.org/zap"
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

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing Viper; %v", err)
	}
	cfg := config.GetConfig()

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}

	l, err := NewLearner(*cfg, client)
	if err != nil {
		t.Fatalf("Error creating learner; %v", err)
	}

	if err := l.Reconcile(context.Background()); err != nil {
		t.Fatalf("Error reconciling; %v", err)
	}
}
