package assets

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
	"os"
	"testing"
)

func Test_Download(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Failed to initialize viper: %v", err)
	}
	cfg := config.GetConfig()

	// Setup logs
	c := zap.NewDevelopmentConfig()
	log, err := c.Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)
	m, err := NewManager(*cfg)
	// Hardcoding the download directory means on each run we won't try to redownload the assets
	// Useful during development/debugging
	m.downloadDir = "/tmp/foyleAssetDownloads"

	if err != nil {
		t.Fatalf("Error creating manager; %v", err)
	}

	if err := m.Download(context.Background()); err != nil {
		t.Fatalf("Error downloading assets; %v", err)
	}
}
