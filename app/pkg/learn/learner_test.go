package learn

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/testutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/jlewi/foyle/app/pkg/analyze"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"go.uber.org/zap"
)

func Test_Learner(t *testing.T) {
	// This isn't really a test because it depends on your configuration and logs.
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	log, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing Viper; %v", err)
	}
	cfg := config.GetConfig()

	// If the directory doesn't exit opening the SQLLite database will fail.
	sessionsDBFile := cfg.GetSessionsDB()
	db, err := sql.Open(analyze.SQLLiteDriver, sessionsDBFile)
	defer helpers.DeferIgnoreError(db.Close)
	if err != nil {
		t.Fatalf("Failed to open database: %v", cfg.GetSessionsDB())
	}
	sessions, err := analyze.NewSessionsManager(db)
	if err != nil {
		t.Fatalf("Error creating sessions manager; %v", err)
	}

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}

	l, err := NewLearner(*cfg, client, sessions)
	if err != nil {
		t.Fatalf("Error creating learner; %v", err)
	}

	id := "01JD88ZCBD72YPBKVH8CAMC3MT"
	if err := l.Reconcile(context.Background(), id); err != nil {
		t.Fatalf("Error reconciling; %v", err)
	}
}

func Test_sessionToQuery(t *testing.T) {
	type testCase struct {
		name     string
		session  *logspb.Session
		expected *v1alpha1.GenerateRequest
	}

	testCases := []testCase{
		{
			// We want to make sure the selected cell isn't in the query
			name: "simple",
			session: &logspb.Session{
				FullContext: &v1alpha1.FullContext{
					Notebook: &parserv1.Notebook{
						Cells: []*parserv1.Cell{
							{
								Value: "Cell 1",
								Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
							},
							{
								Value: "Cell 2",
								Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
							},
						},
					},
					Selected: 1,
				},
			},
			expected: &v1alpha1.GenerateRequest{
				Doc: &v1alpha1.Doc{
					Blocks: []*v1alpha1.Block{
						{
							Kind:     v1alpha1.BlockKind_MARKUP,
							Contents: "Cell 1",
							Outputs:  []*v1alpha1.BlockOutput{},
						},
					},
				},
				SelectedIndex: 0,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := sessionToQuery(c.session)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if d := cmp.Diff(c.expected, actual, testutil.DocComparer, cmpopts.IgnoreUnexported(v1alpha1.GenerateRequest{})); d != "" {
				t.Fatalf("Diff:\n%v", d)
			}
		})
	}
}
