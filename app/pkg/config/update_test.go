package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/api"
	"github.com/spf13/viper"
)

func Test_UpdateViperConfig(t *testing.T) {
	type testCase struct {
		name       string
		configFile string
		expression string
		expected   *Config
	}

	cases := []testCase{
		{
			name:       "model",
			configFile: "partial.yaml",
			expression: "agent.model=some-other-model",
			expected: &Config{
				Logging: Logging{
					Level:           "info",
					Sinks:           []LogSink{{JSON: true, Path: "gcplogs:///projects/fred-dev/logs/foyle"}, {Path: "stderr"}},
					MaxDelaySeconds: 30,
				},
				Agent: &api.AgentConfig{
					Model:         "some-other-model",
					ModelProvider: "anthropic",
					RAG: &api.RAGConfig{
						Enabled:    true,
						MaxResults: 3,
					},
				},
				Server: ServerConfig{
					BindAddress:         "0.0.0.0",
					GRPCPort:            9080,
					HttpPort:            8877,
					HttpMaxReadTimeout:  time.Minute,
					HttpMaxWriteTimeout: time.Minute,
				},
				OpenAI: &OpenAIConfig{
					APIKeyFile: "/Users/red/secrets/openapi.api.key",
				},
				Telemetry: &TelemetryConfig{
					Honeycomb: &HoneycombConfig{
						APIKeyFile: "/Users/fred/secrets/honeycomb.api.key",
					},
				},
				Learner: &LearnerConfig{LogDirs: []string{}, ExampleDirs: []string{"/Users/fred/.foyle/training"}},
			},
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

			cfg, err := UpdateViperConfig(v, c.expression)
			if err != nil {
				t.Fatalf("Failed to update config; %+v", err)
			}

			opts := cmpopts.IgnoreUnexported(Config{})
			if d := cmp.Diff(c.expected, cfg, opts); d != "" {
				t.Fatalf("Unexpected diff:\n%+v", d)
			}
		})
	}
}
