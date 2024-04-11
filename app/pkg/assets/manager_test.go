package assets

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
)

func Test_ResolveImage(t *testing.T) {
	type testCase struct {
		image    string
		tag      string
		expected string
	}

	cases := []testCase{
		{
			image:    "ghcr.io/jlewi/vscode-web-assets",
			tag:      "1234",
			expected: "ghcr.io/jlewi/vscode-web-assets:1234",
		},
		{
			image:    "ghcr.io/jlewi/vscode-web-assets:abcd",
			tag:      "1234",
			expected: "ghcr.io/jlewi/vscode-web-assets:abcd",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			actual, err := resolveTag(c.image, c.tag)
			if err != nil {
				t.Fatalf("Error resolving tag; %v", err)
			}
			if actual != c.expected {
				t.Errorf("Expected %v; got %v", c.expected, actual)
			}
		})
	}
}

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

	if err := m.Download(context.Background(), "latest"); err != nil {
		t.Fatalf("Error downloading assets; %v", err)
	}
}
