package eval

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/app/pkg/runme/ulid"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jlewi/foyle/app/api"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/executor"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	uninitializedDistance = -1
)

type Evaluator struct {
	config config.Config
	parser *executor.BashishParser
}

// NewEvaluator creates a new Evaluator
// The evaluator assumes that the analyzer is already running in the background and processing logs.
// TODO(https://github.com/jlewi/foyle/issues/140): The evaluator may need to be updated now that we continuously
// process logs in the background.
//
// TODO(jeremy): We should probably redo the Evaluator so that instead of setting up the Agent we just
// communicate with the Agent via RPC.
func NewEvaluator(cfg config.Config) (*Evaluator, error) {
	parser, err := executor.NewBashishParser()

	if err != nil {
		return nil, err
	}

	return &Evaluator{
		config: cfg,
		parser: parser,
	}, nil
}

func (e *Evaluator) ReconcileNode(ctx context.Context, node *yaml.RNode) error {
	experiment := &api.Experiment{}
	if err := node.YNode().Decode(experiment); err != nil {
		return errors.Wrapf(err, "Failed to decode experiment")
	}

	return e.Reconcile(ctx, *experiment)
}

func (e *Evaluator) Reconcile(ctx context.Context, experiment api.Experiment) error {
	log := logs.FromContext(ctx).WithValues("experiment", experiment.Metadata.Name)
	//log.Info("Opening database", "database", experiment.Spec.DBDir)
	//db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
	//if err != nil {
	//	return err
	//}
	//defer helpers.DeferIgnoreError(db.Close)
	//
	//if experiment.Spec.Agent == nil {
	//	return errors.New("Agent is required")
	//}
	aiClient := newAIServiceClient(experiment.Spec.AgentAddress)

	manager, err := openResultsManager(experiment.Spec.OutputDB)
	if err != nil {
		return errors.Wrapf(err, "Failed to open results manager from file %s", experiment.Spec.OutputDB)
	}

	// Find all the binary protobuf files in the eval directory.
	// This should contain EvalExample protos.
	files, err := listEvalFiles(ctx, experiment.Spec.EvalDir)
	if err != nil {
		return err
	}

	log.Info("Found eval files", "numFiles", len(files))

	// 1. Get the last eval example id that was processed.
	// lastExample, err := getIDOfLastExample(edb)

	// Default the time of the lastProcessedEval example to some time in the past.
	// This way all examples should be after it and get reprocessed
	lastProcessedTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Make sure the files are in sorted order because the filename should contain the ULID.
	// The files should be named ${SESSION_ID}.evalexample.binpb

	// Loop over the eval examples and load them
	examples := make([]*v1alpha1.EvalExample, 0, len(files))
	for _, exampleFile := range files {
		b, err := os.ReadFile(exampleFile)
		if err != nil {
			// TODO(jeremy): We should probably store the error in the DB.
			log.Error(err, "Failed to read file", "file", exampleFile)
			continue
		}

		example := &v1alpha1.EvalExample{}
		if err := proto.Unmarshal(b, example); err != nil {
			log.Error(err, "Failed to unmarshal example", "file", exampleFile)
			continue
		}
		examples = append(examples, example)
	}

	// Now sort the examples in time order so we can process them in the same order they actually occurred
	sortEvalExamplesInTime(examples)

	// Now generate predictions for any results that are missing them.
	if err := e.processExamples(ctx, examples, lastProcessedTime, aiClient, manager); err != nil {
		return err
	}

	// TODO(jeremy): We should get the traces via API because only one process can access the pebble DB at a time.
	// And the agent needs access to the pebble DB traces.
	tracesDB, err := pebble.Open(e.config.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(tracesDB.Close)

	if err := e.reconcileBestRAGResult(ctx, nil, tracesDB); err != nil {
		return err
	}

	return nil
}

func (e *Evaluator) processExamples(ctx context.Context, examples []*v1alpha1.EvalExample, lastProcessedTime time.Time, client v1alpha1connect.AIServiceClient, manager *ResultsManager) error {
	oLog := logs.FromContext(ctx)

	// Now iterate over the examples and process them.
	for _, example := range examples {
		log := oLog.WithValues("exampleId", example.GetId())

		if example.Time.AsTime().Before(lastProcessedTime) {
			log.V(logs.Debug).Info("Skipping example; already processed")
			continue
		}

		sessionID := ulid.GenerateID()

		//selectedCell := example.GetFullContext().GetNotebook().GetCells()[example.GetFullContext().GetSelected()]
		//// We need to send a LOG event to the agent to simulate the cells being executed.
		//
		logEventReq := &v1alpha1.LogEventsRequest{}
		logEventReq.Events = append(logEventReq.Events, &v1alpha1.LogEvent{
			Type:          v1alpha1.LogEventType_SESSION_START,
			ContextId:     sessionID,
			SelectedIndex: example.GetFullContext().GetSelected(),
		})

		_, err := client.LogEvents(ctx, connect.NewRequest(logEventReq))
		if err != nil {
			log.Error(err, "Failed to log events")
			// For now abort on error to see what's going on.
			return errors.Wrapf(err, "Failed to log events")
		}

		request := &v1alpha1.GenerateCellsRequest{
			Notebook: example.GetFullContext().GetNotebook(),
		}

		resp, err := client.GenerateCells(ctx, connect.NewRequest(request))
		if err != nil {
			log.Error(err, "Failed to generate cells")
			uErr := manager.Update(ctx, example.GetId(), func(result *v1alpha1.EvalResult) error {
				result.Error = err.Error()
				return nil
			})
			if uErr != nil {
				log.Error(uErr, "Failed to update result")
			}
			continue
		}

		uErr := manager.Update(ctx, example.GetId(), func(result *v1alpha1.EvalResult) error {
			result.ActualCells = resp.Msg.GetCells()
			return nil
		})
		if uErr != nil {
			log.Error(uErr, "Failed to update result")
		}

		// TODO(jeremy): We should set the traceId based on OTEL.
		// There's a couple of ways we could do this.
		// 1. We could have the client set the traceId but then we'd have to configure the server to trust the client
		//    trace per https://github.com/connectrpc/otelconnect-go?tab=readme-ov-file#configuration-for-internal-services
		// 2. The server could set the trace id and I believe it should be in the response? and then the client can
		// get it?
		// result.GenTraceId = resp.Msg.GetTraceId()

		// We need to send a LOG event to the agent to simulate the cells being executed.
		executeEventReq := &v1alpha1.LogEventsRequest{}

		for _, cell := range resp.Msg.GetCells() {

			if cell.Kind != parserv1.CellKind_CELL_KIND_CODE {
				continue
			}

			executeEventReq.Events = append(executeEventReq.Events, &v1alpha1.LogEvent{
				Type: v1alpha1.LogEventType_SESSION_START,
			})

			executeEventReq.Events = append(executeEventReq.Events, &v1alpha1.LogEvent{
				Type: v1alpha1.LogEventType_EXECUTE,
				Cells: []*parserv1.Cell{
					cell,
				},
				SelectedIndex: 0,
				SelectedId:    converters.GetCellID(cell),
			})
		}

		if _, err := client.LogEvents(ctx, connect.NewRequest(executeEventReq)); err != nil {
			log.Error(err, "Failed to log events")
			// For now abort on error to see what's going on.
			return errors.Wrapf(err, "Failed to log events")
		}
	}
	return nil
}

//func updateResult(ctx context.Context, id string, result *v1alpha1.EvalResult, db *pebble.DB) error {
//	b, err := proto.Marshal(result)
//	if err != nil {
//		return errors.Wrapf(err, "Failed to marshal result")
//	}
//	if err := db.Set([]byte(id), b, nil); err != nil {
//		return errors.Wrapf(err, "Failed to write result to DB")
//	}
//	return nil
//}

//func (e *Evaluator) reconcileDistance(ctx context.Context, db *pebble.DB) error {
//	olog := logs.FromContext(ctx)
//	iter, err := db.NewIterWithContext(ctx, nil)
//	if err != nil {
//		return err
//	}
//	defer iter.Close()
//
//	for iter.First(); iter.Valid(); iter.Next() {
//		key := iter.Key()
//		if key == nil {
//			break
//		}
//
//		log := olog.WithValues("id", string(key))
//		value, err := iter.ValueAndErr()
//		if err != nil {
//			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
//		}
//
//		result := &v1alpha1.EvalResult{}
//		if err := proto.Unmarshal(value, result); err != nil {
//			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
//		}
//
//		if result.Distance >= 0 && result.Status != v1alpha1.EvalResultStatus_UNKNOWN_EVAL_RESULT_STATUS {
//			log.Info("Skipping; distance already computed")
//			continue
//		}
//
//		updateEvalResultDistance(ctx, e.parser, result)
//		log.Info("Updating distance", "distance", result.Distance)
//		if err := updateResult(ctx, string(key), result, db); err != nil {
//			log.Error(err, "Failed to update result")
//		}
//	}
//	return nil
//}

func (e *Evaluator) reconcileBestRAGResult(ctx context.Context, db *pebble.DB, traces *pebble.DB) error {
	log := logs.FromContext(ctx)
	log.Info("Code to identify best RAG result needs to be updated")
	return nil
	//olog := logs.FromContext(ctx)
	//iter, err := db.NewIterWithContext(ctx, nil)
	//if err != nil {
	//	return err
	//}
	//defer iter.Close()
	//
	//for iter.First(); iter.Valid(); iter.Next() {
	//	key := iter.Key()
	//	if key == nil {
	//		break
	//	}
	//
	//	log := olog.WithValues("id", string(key))
	//	value, err := iter.ValueAndErr()
	//	if err != nil {
	//		return errors.Wrapf(err, "Failed to read value for key %s", string(key))
	//	}
	//
	//	result := &v1alpha1.EvalResult{}
	//	if err := proto.Unmarshal(value, result); err != nil {
	//		return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
	//	}
	//
	//	// TODO(jeremy): How do we skip this step in the case where the experiment didn't involve RAG
	//	if result.BestRagResult != nil {
	//		log.Info("Skipping; best RAG result already computed")
	//		continue
	//	}
	//
	//	genTrace := &logspb.Trace{}
	//	if err := dbutil.GetProto(traces, result.GenTraceId, genTrace); err != nil {
	//		log.Error(err, "Failed to read gen trace", "id", result.GenTraceId)
	//		continue
	//	}
	//
	//	for _, span := range genTrace.Spans {
	//		if span.GetRag() == nil {
	//			continue
	//		}
	//		rag := span.GetRag()
	//		if rag.Results == nil {
	//			continue
	//		}
	//
	//		for _, ragResult := range rag.Results {
	//			if ragResult.Example == nil {
	//				continue
	//			}
	//			if result.BestRagResult == nil {
	//				result.BestRagResult = ragResult
	//				continue
	//			}
	//
	//			if result.BestRagResult.Score < ragResult.Score {
	//				result.BestRagResult = ragResult
	//			}
	//		}
	//	}
	//
	//	if result.BestRagResult == nil {
	//		continue
	//	}
	//	if err := updateResult(ctx, string(key), result, db); err != nil {
	//		log.Error(err, "Failed to update result")
	//	}
	//}
	//return nil
}

//func (e *Evaluator) updateGoogleSheet(ctx context.Context, experiment api.Experiment, db *pebble.DB) error {
//	log := logs.FromContext(ctx)
//	if e.config.Eval == nil || e.config.Eval.GCPServiceAccount == "" {
//		return errors.New("GCPServiceAccount is required to update Google Sheet")
//	}
//
//	sheetName := experiment.Spec.SheetName
//	sheetID := experiment.Spec.SheetID
//
//	if sheetID == "" {
//		return errors.New("SheetID is required to update Google Sheet")
//	}
//
//	if sheetName == "" {
//		return errors.New("SheetName is required to update Google Sheet")
//	}
//
//	log = log.WithValues("spreadsheetID", sheetID, "sheetName", sheetName)
//	log.Info("Updating Google Sheet")
//	credentialsConfig := &impersonate.CredentialsConfig{
//		TargetPrincipal: e.config.Eval.GCPServiceAccount,
//		Scopes:          []string{"https://www.googleapis.com/auth/spreadsheets", "https://www.googleapis.com/auth/drive"},
//	}
//
//	credentials, err := impersonate.CredentialsTokenSource(ctx, *credentialsConfig)
//	if err != nil {
//		log.Error(err, "Unable to create impersonated credentials")
//		return err
//	}
//
//	srv, err := sheets.NewService(ctx, option.WithTokenSource(credentials))
//	if err != nil {
//		log.Error(err, "Unable to retrieve Sheets client")
//		return err
//	}
//
//	// Create the sheet if it doesn't exist
//	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
//		Requests: []*sheets.Request{
//			{
//				AddSheet: &sheets.AddSheetRequest{
//					Properties: &sheets.SheetProperties{
//						Title: experiment.Spec.SheetName,
//					},
//				},
//			},
//		},
//	}
//
//	_, err = srv.Spreadsheets.BatchUpdate(experiment.Spec.SheetID, batchUpdateRequest).Context(ctx).Do()
//	if err != nil {
//		apiErr, ok := err.(*googleapi.Error)
//		if ok {
//			if apiErr.Code == 400 {
//				log.V(1).Info("Sheet already exists")
//			} else {
//				log.Error(err, "Unable to create new sheet ")
//				return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
//			}
//		} else {
//			return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
//		}
//	}
//
//	// Prepare the value range to write
//	writeRange := sheetName
//	values := [][]interface{}{{"id", "file", "prompt", "actual", "expected", "distance", "normalized_distance", "best_rag"}}
//
//	iter, err := db.NewIterWithContext(ctx, nil)
//	if err != nil {
//		return err
//	}
//	defer iter.Close()
//
//	for iter.First(); iter.Valid(); iter.Next() {
//		key := iter.Key()
//		if key == nil {
//			break
//		}
//
//		value, err := iter.ValueAndErr()
//		if err != nil {
//			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
//		}
//
//		result := &v1alpha1.EvalResult{}
//		if err := proto.Unmarshal(value, result); err != nil {
//			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
//		}
//
//		prompt := docs.DocToMarkdown(result.Example.Query)
//		row := []interface{}{result.Example.Id, result.ExampleFile, prompt, docs.BlocksToMarkdown(result.Actual), docs.BlocksToMarkdown(result.Example.Answer), result.Distance, result.NormalizedDistance}
//
//		bestRAG := ""
//		if result.BestRagResult != nil {
//			if result.BestRagResult.Example.Query != nil {
//				bestRAG = docs.DocToMarkdown(result.BestRagResult.Example.Query)
//			}
//		}
//		row = append(row, bestRAG)
//		values = append(values, row)
//	}
//	valueRange := &sheets.ValueRange{
//		Values: values,
//	}
//
//	// Write the value range to the sheet
//	_, err = srv.Spreadsheets.Values.Update(sheetID, writeRange, valueRange).
//		ValueInputOption("USER_ENTERED").
//		Context(ctx).
//		Do()
//	if err != nil {
//		log.Error(err, "Unable to write data to sheet")
//		return errors.Wrapf(err, "Unable to write data to sheet")
//	}
//
//	return nil
//}

// TODO(jeremy): We should get rid of this function and one that calls it
func findUnloadedFiles(ctx context.Context, db *pebble.DB, files []string) ([]string, error) {
	return nil, errors.New("findUnloadedFiles needs to be updated to work with new DB and protos")
	//unprocessed := map[string]bool{}
	//
	//iter, err := db.NewIterWithContext(ctx, nil)
	//if err != nil {
	//	return nil, err
	//}
	//defer iter.Close()
	//
	//for _, file := range files {
	//	unprocessed[file] = true
	//}
	//
	//// Iterate over the files in the DB and remove them from the list of files to load.
	//for iter.First(); iter.Valid(); iter.Next() {
	//	key := iter.Key()
	//	if key == nil {
	//		break
	//	}
	//
	//	value, err := iter.ValueAndErr()
	//	if err != nil {
	//		// Should we ignore the error?
	//		return nil, errors.Wrapf(err, "Failed to read value for key %s", string(key))
	//	}
	//
	//	result := &v1alpha1.EvalResult{}
	//	if err := proto.Unmarshal(value, result); err != nil {
	//		return nil, errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
	//	}
	//
	//	delete(unprocessed, result.ExampleFile)
	//
	//}
	//
	//toProcess := make([]string, 0, len(unprocessed))
	//for file := range unprocessed {
	//	toProcess = append(toProcess, file)
	//}
	//
	//return toProcess, nil
}

// listEvalFiles returns a list of the all the binary protobuf files in the directory evalDir.
func listEvalFiles(ctx context.Context, evalDir string) ([]string, error) {
	examples := make([]string, 0, 100)
	err := filepath.Walk(evalDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".binpb" {
			return nil
		}

		examples = append(examples, path)
		return nil
	})

	return examples, err
}

// loadMarkdownFiles loads a bunch of markdown files representing evaluation data and converts them into example
// protos. The final block in the markdown file is treated as the answer.
//func loadMarkdownAnswerFiles(ctx context.Context, db *pebble.DB, files []string) error {
//	oLog := logs.FromContext(ctx)
//
//	allErrors := &helpers.ListOfErrors{}
//	for _, path := range files {
//		log := oLog.WithValues("path", path)
//		log.Info("Processing file")
//
//		contents, err := os.ReadFile(path)
//		if err != nil {
//			log.Error(err, "Failed to read file")
//			allErrors.AddCause(err)
//			// Keep going
//			continue
//		}
//
//		doc := &v1alpha1.Doc{}
//
//		blocks, err := docs.MarkdownToBlocks(string(contents))
//		if err != nil {
//			log.Error(err, "Failed to convert markdown to blocks")
//			allErrors.AddCause(err)
//			// Keep going
//			continue
//		}
//
//		doc.Blocks = blocks
//
//		if len(doc.GetBlocks()) < 2 {
//			log.Info("Skipping doc; too few blocks; at least two are required")
//			continue
//		}
//
//		answer := doc.GetBlocks()[len(doc.GetBlocks())-1]
//		doc.Blocks = doc.Blocks[:len(doc.GetBlocks())-1]
//		if answer.Kind != v1alpha1.BlockKind_CODE {
//			log.Info("Skipping doc; last block must be code")
//			continue
//		}
//
//		// We generate a stable ID for the example by hashing the contents of the document.
//		example := &v1alpha1.Example{
//			Query:  doc,
//			Answer: []*v1alpha1.Block{answer},
//		}
//		example.Id = HashExample(example)
//
//		result := &v1alpha1.EvalResult{
//			Example:     example,
//			ExampleFile: path,
//			// initialize distance to a negative value so we can tell when it hasn't been computed
//			Distance: uninitializedDistance,
//		}
//
//		if err := dbutil.SetProto(db, example.GetId(), result); err != nil {
//			log.Error(err, "Failed to write result to DB")
//			allErrors.AddCause(err)
//			// Keep going
//			continue
//		}
//	}
//
//	if len(allErrors.Causes) > 0 {
//		return allErrors
//	}
//
//	return nil
//}

func sortEvalExamplesInTime(examples []*v1alpha1.EvalExample) {
	sort.Slice(examples, func(i, j int) bool {
		// Convert the Time field to time.Time objects
		timeI := examples[i].Time.AsTime()
		timeJ := examples[j].Time.AsTime()

		// Compare the times
		return timeI.Before(timeJ)
	})
}

func newAIServiceClient(baseURL string) v1alpha1connect.AIServiceClient {
	// Create a new client
	client := v1alpha1connect.NewAIServiceClient(
		newHTTPClient(),
		baseURL,
	)
	return client
}
