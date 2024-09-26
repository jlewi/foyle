package eval

import (
	"connectrpc.com/connect"
	"context"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"os"
	"path/filepath"
	"testing"

	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/config"
	"go.uber.org/zap"
)

func Test_Evaluator(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	// This test assumes you have already started an agent with the appropriate configuration that you
	// want to evaluate.

	log, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Error initializing Viper; %v", err)
	}
	cfg := config.GetConfig()

	e, err := NewEvaluator(*cfg)
	if err != nil {
		t.Fatalf("Error creating evaluator; %v", err)
	}

	experiment, err := experimentForTesting()
	if err != nil {
		t.Fatalf("Error creating experiment; %v", err)
	}

	if err := e.Reconcile(context.Background(), *experiment); err != nil {
		t.Fatalf("Error reconciling; %+v", err)
	}
}

func Test_Evaluator_RunGenerate(t *testing.T) {
	result := &v1alpha1.EvalResult{
		Example: &v1alpha1.EvalExample{
			FullContext: &v1alpha1.FullContext{
				Notebook: &parserv1.Notebook{
					Cells: []*parserv1.Cell{
						{
							Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
							Value: "RunSomeCode",
						},
					},
				},
			},
			ExpectedCells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_CODE,
					Value: "gcloud builds list",
				},
			},
		},
	}
	fake := &fakeClient{
		GenerateCellsResponse: &v1alpha1.GenerateCellsResponse{
			Cells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_CODE,
					Value: "some command",
				},
			},
		},

		generateTraceID: "someTrace",
	}
	if err := runGenerate(context.Background(), result, fake); err != nil {
		t.Fatalf("Error running execute; %v+", err)
	}

	if result.ActualCells[0].Value != "some command" {
		t.Errorf("Expected actual cell to be 'some command' but got %v", result.ActualCells[0].Value)
	}

	if result.GetGenTraceId() != "someTrace" {
		t.Errorf("Expected trace id to be 'some trace' but got %v", result.GetGenTraceId())
	}

	// Make sure the events are correct
	if fake.Events[0].Type != v1alpha1.LogEventType_SESSION_START {
		t.Errorf("Expected first event to be a session start but got %v", fake.Events[0].Type)
	}

	if fake.Events[1].Type != v1alpha1.LogEventType_SESSION_END {
		t.Errorf("Expected last event to be a session end but got %v", fake.Events[0].Type)
	}
}

func Test_Evaluator_RunExecute(t *testing.T) {
	result := &v1alpha1.EvalResult{
		Example: &v1alpha1.EvalExample{
			ExpectedCells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_CODE,
					Value: "gcloud executed command",
					Metadata: map[string]string{
						converters.IdField:      "idFieldShouldBeIgnored",
						converters.RunmeIdField: "runmeIdFieldShouldBeIgnored",
					},
				},
			},
		},
		ActualCells: []*parserv1.Cell{
			{
				Kind:  parserv1.CellKind_CELL_KIND_CODE,
				Value: "gcloud predicted command",
				Metadata: map[string]string{
					converters.IdField: "idOfActualCell",
				},
			},
		},
	}

	fake := &fakeClient{}
	if err := runExecute(context.Background(), result, fake); err != nil {
		t.Fatalf("Error running execute; %v+", err)
	}

	// Make sure the events are correct
	if fake.Events[0].Type != v1alpha1.LogEventType_SESSION_START {
		t.Errorf("Expected first event to be a session start but got %v", fake.Events[0].Type)
	}

	if fake.Events[1].Type != v1alpha1.LogEventType_EXECUTE {
		t.Errorf("Expected event to be a execution event  but got %v", fake.Events[0].Type)
	}

	if fake.Events[1].SelectedId != "idOfActualCell" {
		t.Errorf("SelectedID is not correct")
	}

	if converters.GetCellID(fake.Events[1].Cells[0]) != "idOfActualCell" {
		t.Errorf("ID of cell is not correct")
	}

	if fake.Events[1].Cells[0].Value != "gcloud executed command" {
		t.Errorf("Executed cell is not the expected cell ")
	}

	if fake.Events[2].Type != v1alpha1.LogEventType_SESSION_END {
		t.Errorf("Expected last event to be a session end but got %v", fake.Events[0].Type)
	}
}

type fakeClient struct {
	Events                []*v1alpha1.LogEvent
	GenerateCellsResponse *v1alpha1.GenerateCellsResponse
	generateTraceID       string
}

func (f *fakeClient) StreamGenerate(context.Context) *connect.BidiStreamForClient[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse] {
	//TODO implement me
	panic("implement me")
}

func (f *fakeClient) GenerateCells(ctx context.Context, req *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error) {
	if f.GenerateCellsResponse == nil {
		return connect.NewResponse(&v1alpha1.GenerateCellsResponse{}), nil
	}
	resp := connect.NewResponse(f.GenerateCellsResponse)
	resp.Header().Set(agent.TraceIDHeader, f.generateTraceID)

	return resp, nil
}

func (f *fakeClient) GetExample(context.Context, *connect.Request[v1alpha1.GetExampleRequest]) (*connect.Response[v1alpha1.GetExampleResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeClient) Status(context.Context, *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeClient) LogEvents(ctx context.Context, req *connect.Request[v1alpha1.LogEventsRequest]) (*connect.Response[v1alpha1.LogEventsResponse], error) {
	if f.Events == nil {
		f.Events = make([]*v1alpha1.LogEvent, 0, 100)
	}
	f.Events = append(f.Events, req.Msg.Events...)
	return connect.NewResponse(&v1alpha1.LogEventsResponse{}), nil
}

//func Test_Evaluator_Google_Sheets(t *testing.T) {
//	if os.Getenv("GITHUB_ACTIONS") != "" {
//		t.Skipf("Test is skipped in GitHub actions")
//	}
//
//	t.Fatalf("Evaluator test needs to be updated per https://github.com/jlewi/foyle/issues/140")
//
//	log, err := zap.NewDevelopmentConfig().Build()
//	if err != nil {
//		t.Fatalf("Error creating logger; %v", err)
//	}
//	zap.ReplaceGlobals(log)
//
//	if err := config.InitViper(nil); err != nil {
//		t.Fatalf("Error initializing Viper; %v", err)
//	}
//	cfg := config.GetConfig()
//
//	e, err := NewEvaluator(*cfg)
//	if err != nil {
//		t.Fatalf("Error creating evaluator; %v", err)
//	}
//
//	experiment, err := experimentForTesting()
//	if err != nil {
//		t.Fatalf("Error creating experiment; %v", err)
//	}
//
//	db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
//	if err != nil {
//		t.Fatalf("Error opening DB; %v", err)
//	}
//	defer helpers.DeferIgnoreError(db.Close)
//	//if err := e.updateGoogleSheet(context.Background(), *experiment, db); err != nil {
//	//	t.Fatalf("Error updating Google Sheet; %v", err)
//	//}
//}

func experimentForTesting() (*api.Experiment, error) {
	//cwd, err := os.Getwd()
	//if err != nil {
	//	return nil, errors.Wrapf(err, "Error getting working directory")
	//}
	//evalDir, err := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "data", "eval"))
	//if err != nil {
	//	return nil, errors.Wrapf(err, "Error getting eval directory")
	//}
	log := zapr.NewLogger(zap.L())
	oDir, err := os.MkdirTemp("", "testOutput")
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating temp directory")
	}

	dbFile := filepath.Join(oDir, "results.sqlite")
	log.Info("Output database", "database", dbFile)

	return &api.Experiment{
		Spec: api.ExperimentSpec{
			// EvalDir is the directory containing the eval example protos
			EvalDir:      "/Users/jlewi/tmp/examples-for-testing",
			AgentAddress: "http://localhost:10777/api",
			OutputDB:     dbFile,
		},
	}, nil
}

//func Test_updateEvalResultDistance(t *testing.T) {
//	type testCase struct {
//		name               string
//		result             *v1alpha1.EvalResult
//		expectedDistance   int32
//		expectedNormalized float32
//	}
//
//	cases := []testCase{
//		{
//			// Test the case where the actual answer contains no codeblocks
//			name: "nocodeblocks",
//			result: &v1alpha1.EvalResult{
//				Example: &v1alpha1.EvalExample{
//					Id: "1234",
//					Answer: []*v1alpha1.Block{
//						{
//							Kind:     v1alpha1.BlockKind_CODE,
//							Contents: "gcloud builds list",
//						},
//					},
//				},
//				ExampleFile: "",
//				Actual: []*v1alpha1.Block{
//					{
//						Kind:     v1alpha1.BlockKind_MARKUP,
//						Contents: "Not a code cell",
//					},
//				},
//			},
//			expectedDistance:   3,
//			expectedNormalized: 1.0,
//		},
//	}
//	parser, err := executor.NewBashishParser()
//	if err != nil {
//		t.Fatalf("Error creating parser; %v", err)
//	}
//
//	for _, c := range cases {
//		t.Run(c.name, func(t *testing.T) {
//			updateEvalResultDistance(context.Background(), parser, c.result)
//			if err != nil {
//				t.Fatalf("Unexpected error: %v", err)
//			}
//			if c.result.Distance != c.expectedDistance {
//				t.Errorf("Expected distance %d but got %d", c.expectedDistance, c.result.Distance)
//			}
//			if c.result.NormalizedDistance != c.expectedNormalized {
//				t.Errorf("Expected normalized distance %f but got %f", c.expectedNormalized, c.result.NormalizedDistance)
//			}
//		})
//	}
//}
