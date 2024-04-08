package agent

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
)

func Test_Generate(t *testing.T) {
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

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}
	agent, err := NewAgent(*cfg, client)

	if err != nil {
		t.Fatalf("Error creating agent; %v", err)
	}

	req := &v1alpha1.GenerateRequest{
		Doc: &v1alpha1.Doc{
			Blocks: []*v1alpha1.Block{
				{
					Contents: "Use gcloud to list all the cloud build jobs in project foyle",
				},
			},
		},
	}
	resp, err := agent.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Error generating; %v", err)
	}

	t.Logf("Response: %+v", resp)
}
