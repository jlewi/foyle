package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func Test_ConfigDefaultConfig(t *testing.T) {
	type testCase struct {
		name                    string
		configFile              string
		expectedRAG             bool
		expectedHTTPPort        int
		expectedMaxDelaySeconds int
	}

	cases := []testCase{
		{
			name:                    "config-file-does-not-exist",
			configFile:              "doesnotexist.yaml",
			expectedRAG:             defaultRagEnabled,
			expectedHTTPPort:        defaultHTTPPort,
			expectedMaxDelaySeconds: 30,
		},
		{
			name:                    "empty-file",
			configFile:              "empty.yaml",
			expectedRAG:             defaultRagEnabled,
			expectedHTTPPort:        defaultHTTPPort,
			expectedMaxDelaySeconds: 30,
		},
		{
			name:                    "partial",
			configFile:              "partial.yaml",
			expectedRAG:             defaultRagEnabled,
			expectedHTTPPort:        defaultHTTPPort,
			expectedMaxDelaySeconds: 30,
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory")
	}
	tDir := filepath.Join(cwd, "test_data")

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create an empty configuration file and run various assertions on it
			v := viper.New()
			v.SetConfigFile(filepath.Join(tDir, c.configFile))

			if err := InitViperInstance(v, nil); err != nil {
				t.Fatalf("Failed to initialize the configuration.")
			}

			cfg, err := getConfigFromViper(v)

			if err != nil {
				t.Fatalf("Failed to get config; %+v", err)
			}

			if cfg.UseRAG() != c.expectedRAG {
				t.Errorf("UseRAG want %v; got %v", c.expectedRAG, cfg.UseRAG())
			}

			if len(cfg.GetTrainingDirs()) == 0 {
				t.Errorf("GetTrainingDirs shouldn't return an empty list")
			}

			if cfg.Server.HttpPort != c.expectedHTTPPort {
				t.Errorf("HttpPort want %d; got %d", c.expectedHTTPPort, cfg.Server.HttpPort)
			}
		})
	}
}
