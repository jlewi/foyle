package analyze

import (
	"github.com/jlewi/foyle/app/pkg/config"
	"k8s.io/utils/temp"
	"testing"
)

func Test_SessionsDB(t *testing.T) {
	dir, err := temp.CreateTempDir("sessionsDBTest")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}

	cfg := config.Config{}
	// TO control the location of the sqllite database we have to set LogDir
	cfg.Logging = config.Logging{
		LogDir: dir.Name,
	}

	_, err = NewSessionsDB(cfg)
	if err != nil {
		t.Fatalf("Error creating SessionsDB: %v", err)
	}
}
