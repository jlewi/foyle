package eval

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"os"
	"path/filepath"
	"sort"
	"time"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/go-cmd/cmd"

	"github.com/jlewi/foyle/app/pkg/dbutil"

	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/oai"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/executor"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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
	log.Info("Opening database", "database", experiment.Spec.DBDir)
	db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(db.Close)

	if experiment.Spec.Agent == nil {
		return errors.New("Agent is required")
	}
	agent, err := e.setupAgent(ctx, *experiment.Spec.Agent)
	if err != nil {
		return err
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

	// Default the time of the lastProcessedEval example to some time in the future.
	// This way all examples should be before it and get reprocessed
	lastProcessedTime := time.Now().Add(24 * time.Hour * 365 * 10)

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
	if err := e.processExamples(ctx, db, agent); err != nil {
		return err
	}

	// TODO(jeremy): We should get the traces via API because only one process can access the pebble DB at a time.
	// And the agent needs access to the pebble DB traces.
	tracesDB, err := pebble.Open(e.config.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(tracesDB.Close)

	if err := e.reconcileBestRAGResult(ctx, db, tracesDB); err != nil {
		return err
	}

	// Compute the distance
	if err := e.reconcileDistance(ctx, db); err != nil {
		return err
	}

	// Update the Google Sheet
	if err := e.updateGoogleSheet(ctx, experiment, db); err != nil {
		return err
	}
	return nil
}

func (e *Evaluator) setupAgent(ctx context.Context, agentConfig api.AgentConfig) (*agent.Agent, error) {
	cfg := e.config.DeepCopy()

	// Swap out the AgentConfig
	cfg.Agent = &agentConfig

	// Ensure we are in evaluation mode.
	cfg.Agent.EvalMode = true

	client, err := oai.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// TODO(jeremy): This will need to be updated when we support other configurations.
	completer, err := oai.NewCompleter(cfg, client)
	if err != nil {
		return nil, err
	}

	log := logs.FromContext(ctx)
	log.Info("Creating agen without inMemoryExampleDB", "config", cfg.Agent)
	if cfg.Agent.RAG != nil && cfg.Agent.RAG.Enabled {
		return nil, errors.New("RAG is enabled but eval code needs to be updated to ddeal with streaming logs")
	}

	// TODO(jeremy): How should we construct inMemoryExampleDB? In the eval case?
	agent, err := agent.NewAgent(cfg, completer, nil)

	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (e *Evaluator) processExamples(ctx context.Context, examples []*v1alpha1.EvalExample, lastProcessedTime time.Time) error {
	oLog := logs.FromContext(ctx)

	// TODO(jeremy): where should this actually be created and set
	var client v1alpha1connect.AIServiceClient

	// Where do we set this
	var manager *ResultsManager

	// Now iterate over the examples and process them.
	for _, example := range examples {
		log := oLog.WithValues("exampleId", example.GetId())

		if example.Time.AsTime().Before(lastProcessedTime) {
			log.V(logs.Debug).Info("Skipping example; already processed")
			continue
		}

		request := &v1alpha1.GenerateCellsRequest{
			Notebook: example.GetFullContext().GetNotebook(),
		}

		result := &v1alpha1.EvalResult{}

		resp, err := client.GenerateCells(ctx, connect.NewRequest(request))
		if err != nil {
			log.Error(err, "Failed to generate cells")
			result.Error = err.Error()
			uErr := manager.Update(ctx, example.GetId(), func(result *v1alpha1.EvalResult) error {
				result.Error = err.Error()
				return nil
			})
			if uErr != nil {
				log.Error(uErr, "Failed to update result")
			}
			continue
		}

		// Left off editing here
		
		if len(result.GetActual()) > 0 {
			log.Info("Skipping; already have answer", "path", result.ExampleFile)
			// We have the answer so we don't need to generate it.
			continue
		}

		if len(result.Actual) == 0 {
			// Initialize a trace
			resp, err := func() (*v1alpha1.GenerateResponse, error) {
				newCtx, span := tracer().Start(ctx, "(*Evaluator).reconcilePredictions")
				defer span.End()

				// We need to generate the answer.
				return agent.Generate(newCtx, &v1alpha1.GenerateRequest{
					Doc: result.Example.Query,
				})
			}()
			if err != nil {
				result.Error = err.Error()
				result.Status = v1alpha1.EvalResultStatus_ERROR
				continue
			}

			result.Actual = resp.GetBlocks()
			result.GenTraceId = resp.GetTraceId()

			log.Info("Writing result to DB")
			if err := updateResult(ctx, string(key), result, db); err != nil {
				return errors.Wrapf(err, "Failed to write result to DB")
			}
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

func (e *Evaluator) reconcileDistance(ctx context.Context, db *pebble.DB) error {
	olog := logs.FromContext(ctx)
	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		log := olog.WithValues("id", string(key))
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		if result.Distance >= 0 && result.Status != v1alpha1.EvalResultStatus_UNKNOWN_EVAL_RESULT_STATUS {
			log.Info("Skipping; distance already computed")
			continue
		}

		updateEvalResultDistance(ctx, e.parser, result)
		log.Info("Updating distance", "distance", result.Distance)
		if err := updateResult(ctx, string(key), result, db); err != nil {
			log.Error(err, "Failed to update result")
		}
	}
	return nil
}

func (e *Evaluator) reconcileBestRAGResult(ctx context.Context, db *pebble.DB, traces *pebble.DB) error {
	olog := logs.FromContext(ctx)
	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		log := olog.WithValues("id", string(key))
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		// TODO(jeremy): How do we skip this step in the case where the experiment didn't involve RAG
		if result.BestRagResult != nil {
			log.Info("Skipping; best RAG result already computed")
			continue
		}

		genTrace := &logspb.Trace{}
		if err := dbutil.GetProto(traces, result.GenTraceId, genTrace); err != nil {
			log.Error(err, "Failed to read gen trace", "id", result.GenTraceId)
			continue
		}

		for _, span := range genTrace.Spans {
			if span.GetRag() == nil {
				continue
			}
			rag := span.GetRag()
			if rag.Results == nil {
				continue
			}

			for _, ragResult := range rag.Results {
				if ragResult.Example == nil {
					continue
				}
				if result.BestRagResult == nil {
					result.BestRagResult = ragResult
					continue
				}

				if result.BestRagResult.Score < ragResult.Score {
					result.BestRagResult = ragResult
				}
			}
		}

		if result.BestRagResult == nil {
			continue
		}
		if err := updateResult(ctx, string(key), result, db); err != nil {
			log.Error(err, "Failed to update result")
		}
	}
	return nil
}

func updateEvalResultDistance(ctx context.Context, parser *executor.BashishParser, result *v1alpha1.EvalResult) {
	log := logs.FromContext(ctx).WithValues("id", result.GetExample().GetId())
	var actualBlock *v1alpha1.Block

	for _, b := range result.Actual {
		if b.Kind == v1alpha1.BlockKind_CODE {
			actualBlock = b
			break
		}
	}

	if len(result.Example.GetAnswer()) > 1 {
		log.Info("Warning; expected answer more than one answer block. Only the first is used")
	}

	expected, err := parser.Parse(result.Example.Answer[0].GetContents())
	if err != nil {
		log.Error(err, "Failed to parse expected answer to command")
		result.Error = err.Error()
		result.Status = v1alpha1.EvalResultStatus_ERROR
		return
	}

	var actual []executor.Instruction
	if actualBlock != nil {
		parsed, err := parser.Parse(actualBlock.GetContents())
		if err != nil {
			log.Error(err, "Failed to parse actual answer to command")
			result.Error = err.Error()
			result.Status = v1alpha1.EvalResultStatus_ERROR
			return
		}
		actual = parsed
	} else {
		// Since there is no code block. Initialize actual to an empty command.
		// This will cause the distance computed to be the maximum possible distance which is what we want
		actual = []executor.Instruction{
			{
				Command: cmd.NewCmd(""),
			},
		}
	}

	distance, err := Distance(expected[0], actual[0])

	if err != nil {
		log.Error(err, "Failed to compute distance")
		result.Error = err.Error()
		result.Status = v1alpha1.EvalResultStatus_ERROR
		return
	}

	if distance.Max < distance.Distance {
		log.Error(errors.New("Distance is greater than max distance"), "Distance is greater than max distance", "distance", distance.Distance, "max", distance.Max)
	}

	result.Distance = int32(distance.Distance)
	result.NormalizedDistance = distance.Normalized
	result.Status = v1alpha1.EvalResultStatus_DONE
}

func (e *Evaluator) updateGoogleSheet(ctx context.Context, experiment api.Experiment, db *pebble.DB) error {
	log := logs.FromContext(ctx)
	if e.config.Eval == nil || e.config.Eval.GCPServiceAccount == "" {
		return errors.New("GCPServiceAccount is required to update Google Sheet")
	}

	sheetName := experiment.Spec.SheetName
	sheetID := experiment.Spec.SheetID

	if sheetID == "" {
		return errors.New("SheetID is required to update Google Sheet")
	}

	if sheetName == "" {
		return errors.New("SheetName is required to update Google Sheet")
	}

	log = log.WithValues("spreadsheetID", sheetID, "sheetName", sheetName)
	log.Info("Updating Google Sheet")
	credentialsConfig := &impersonate.CredentialsConfig{
		TargetPrincipal: e.config.Eval.GCPServiceAccount,
		Scopes:          []string{"https://www.googleapis.com/auth/spreadsheets", "https://www.googleapis.com/auth/drive"},
	}

	credentials, err := impersonate.CredentialsTokenSource(ctx, *credentialsConfig)
	if err != nil {
		log.Error(err, "Unable to create impersonated credentials")
		return err
	}

	srv, err := sheets.NewService(ctx, option.WithTokenSource(credentials))
	if err != nil {
		log.Error(err, "Unable to retrieve Sheets client")
		return err
	}

	// Create the sheet if it doesn't exist
	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: experiment.Spec.SheetName,
					},
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(experiment.Spec.SheetID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		apiErr, ok := err.(*googleapi.Error)
		if ok {
			if apiErr.Code == 400 {
				log.V(1).Info("Sheet already exists")
			} else {
				log.Error(err, "Unable to create new sheet ")
				return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
			}
		} else {
			return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
		}
	}

	// Prepare the value range to write
	writeRange := sheetName
	values := [][]interface{}{{"id", "file", "prompt", "actual", "expected", "distance", "normalized_distance", "best_rag"}}

	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		prompt := docs.DocToMarkdown(result.Example.Query)
		row := []interface{}{result.Example.Id, result.ExampleFile, prompt, docs.BlocksToMarkdown(result.Actual), docs.BlocksToMarkdown(result.Example.Answer), result.Distance, result.NormalizedDistance}

		bestRAG := ""
		if result.BestRagResult != nil {
			if result.BestRagResult.Example.Query != nil {
				bestRAG = docs.DocToMarkdown(result.BestRagResult.Example.Query)
			}
		}
		row = append(row, bestRAG)
		values = append(values, row)
	}
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Write the value range to the sheet
	_, err = srv.Spreadsheets.Values.Update(sheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Error(err, "Unable to write data to sheet")
		return errors.Wrapf(err, "Unable to write data to sheet")
	}

	return nil
}

//func findUnloadedFiles(ctx context.Context, db *pebble.DB, files []string) ([]string, error) {
//	unprocessed := map[string]bool{}
//
//	iter, err := db.NewIterWithContext(ctx, nil)
//	if err != nil {
//		return nil, err
//	}
//	defer iter.Close()
//
//	for _, file := range files {
//		unprocessed[file] = true
//	}
//
//	// Iterate over the files in the DB and remove them from the list of files to load.
//	for iter.First(); iter.Valid(); iter.Next() {
//		key := iter.Key()
//		if key == nil {
//			break
//		}
//
//		value, err := iter.ValueAndErr()
//		if err != nil {
//			// Should we ignore the error?
//			return nil, errors.Wrapf(err, "Failed to read value for key %s", string(key))
//		}
//
//		result := &v1alpha1.EvalResult{}
//		if err := proto.Unmarshal(value, result); err != nil {
//			return nil, errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
//		}
//
//		delete(unprocessed, result.ExampleFile)
//
//	}
//
//	toProcess := make([]string, 0, len(unprocessed))
//	for file := range unprocessed {
//		toProcess = append(toProcess, file)
//	}
//
//	return toProcess, nil
//}

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
