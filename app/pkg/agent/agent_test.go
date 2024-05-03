package agent

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/api"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
)

func Test_Generate(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	// This test is useful for two things
	// 1. Its a smoke test for all the logic.
	// 2. It can be used to do some basic prompt testing.
	//    We could start to think about this as level 1 evals in the context of Hamel's blog.
	type testCase struct {
		name string
		doc  *v1alpha1.Doc
		// maxResults is the number of results to retrieve for Rag completions.
		// <=0 means no RAG
		maxResults int
	}

	cases := []testCase{
		{
			name: "basic",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use gcloud to list all the cloud build jobs in project foyle",
					},
				},
			},
			maxResults: 0,
		},
		{
			name: "prdiff",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use git to compute a diff and then generate a PR description",
					},
				},
			},
			maxResults: 2,
		},
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

	cfg.Agent.RAG = &api.RAGConfig{
		MaxResults: 3,
	}
	cfg.Agent.RAG.Enabled = true
	agentWithRag, err := NewAgent(*cfg, client)

	if err != nil {
		t.Fatalf("Error creating agent; %v", err)
	}

	cfgNoRag := cfg.DeepCopy()
	cfgNoRag.Agent.RAG.Enabled = false
	agentNoRag, err := NewAgent(cfgNoRag, client)

	if err != nil {
		t.Fatalf("Error creating agent; %v", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &v1alpha1.GenerateRequest{
				Doc: c.doc,
			}
			var agent *Agent
			if c.maxResults > 0 {
				agent = agentWithRag
				agent.config.Agent.RAG.MaxResults = c.maxResults
			} else {
				agent = agentNoRag
			}
			resp, err := agent.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("Error generating; %v", err)
			}
			t.Logf("Response: %+v", resp)
		})
	}
}
