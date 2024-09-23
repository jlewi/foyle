package eval

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"os"
	"testing"
)

func Test_Judge(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping test; this test doesn't run in GitHub Actions")
	}

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Failed to initialize Viper: %v", err)
	}

	cfg := config.GetConfig()

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	judge, err := NewJudge(client)
	if err != nil {
		t.Fatalf("Failed to create Judge: %v", err)
	}

	result := &v1alpha1.EvalResult{
		Example: &v1alpha1.EvalExample{
			ExpectedCells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_CODE,
					Value: "kubectl get pods",
				},
			},
		},
		ActualCells: []*parserv1.Cell{
			{
				Kind:  parserv1.CellKind_CELL_KIND_CODE,
				Value: "kubectl -n foyle get pods",
			},
		},
	}

	if err := judge.Score(context.TODO(), result); err != nil {
		t.Fatalf("Failed to score: %+v", err)
	}

	t.Logf("Judge Equivalent: %v", result.CellsMatchResult.String())
	t.Logf("Judge Explanation:\n%v", result.JudgeExplanation)
}
