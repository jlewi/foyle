package config

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_ConfigDefaultConfig(t *testing.T) {
	// Create an empty configuration file and run various assertions on it
	v := viper.New()
	v.SetConfigFile("/tmp/doesnnotexist.yaml")

	if err := InitViperInstance(v, nil); err != nil {
		t.Fatalf("Failed to initialize the configuration.")
	}

	cfg, err := getConfigFromViper(v)

	if err != nil {
		t.Fatalf("Failed to get config; %+v", err)
	}

	if cfg.UseRAG() != defaultRagEnabled {
		t.Errorf("UseRAG want %v; got %v", defaultRagEnabled, cfg.UseRAG())
	}

	if len(cfg.GetTrainingDirs()) == 0 {
		t.Errorf("GetTrainingDirs shouldn't return an empty list")
	}
}
