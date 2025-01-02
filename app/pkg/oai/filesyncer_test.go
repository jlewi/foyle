package oai

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"os"
	"testing"
)

func Test_FileSyncer(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_FileSyncer is a manual test that is skipped in CICD")
	}
	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing viper: %v", err)
	}
	cfg := config.GetConfig()

	client, err := NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	ctx := context.Background()
	file, err := client.GetFile(ctx, "file-9ik4Eous1jaJ16QRkXwfMZ")
	if err != nil {
		t.Fatalf("Error getting file: %v", err)
	}

	t.Logf("FileName: %v", file.FileName)
}

func Test_convertFilePathToHugoURL(t *testing.T) {
	type testCase struct {
		name     string
		path     string
		expected string
	}

	cases := []testCase{
		{
			name:     "basic",
			path:     `content/docs/runbooks/api/Oncall Foo Issues Runbook.md`,
			expected: "content/docs/runbooks/api/oncall-foo-issues-runbook/",
		},
		{
			name:     "index",
			path:     `docs/content/_index.md`,
			expected: "docs/content/",
		},
		{
			name:     "index",
			path:     `_index.md`,
			expected: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := convertFilePathToHugoURL(c.path)
			if actual != c.expected {
				t.Fatalf("Expected %v, got %v", c.expected, actual)
			}
		})
	}
}
