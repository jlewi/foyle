package eval

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-logr/logr"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/app/pkg/runme/ulid"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/logs/logspbconnect"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/jlewi/foyle/app/api"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/executor"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type Evaluator struct {
	config config.Config
	parser *executor.BashishParser
}

// N.B. One issue with noise in the simulation is that the speed of log processing affects whether example
// has been learned from by the next time it is processed.

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

	if experiment.Spec.AgentAddress == "" {
		return errors.New("AgentAddress is required")
	}

	if experiment.Spec.OutputDB == "" {
		return errors.New("OutputDB is required")
	}

	if experiment.Spec.EvalDir == "" {
		return errors.New("EvalDir is required")
	}

	aiClient := newAIServiceClient(experiment.Spec.AgentAddress)

	logsClient := logspbconnect.NewLogsServiceClient(
		newHTTPClient(),
		experiment.Spec.AgentAddress,
	)

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

	// Now we need to get the id of the last processed example so we can skip over examples that have already been
	// processed
	lastProcessedTime, err := getLastProcessedTime(ctx, manager)
	if err != nil {
		return errors.Wrapf(err, "Failed to get last processed time")
	}

	// N.B. Since we set lastProcessedTime to the time of the last processed result we won't reprocess that result.
	// We might want to rethink that if we want to be able to reprocess an example that failed on an error and
	// we want to retry that

	log.Info("Processing eval examples", "lastProcessedTime", lastProcessedTime)

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
	if err := e.processExamples(ctx, examples, lastProcessedTime, aiClient, logsClient, manager); err != nil {
		return err
	}

	return nil
}

func (e *Evaluator) processExamples(ctx context.Context, examples []*v1alpha1.EvalExample, lastProcessedTime time.Time, client v1alpha1connect.AIServiceClient, logsClient logspbconnect.LogsServiceClient, manager *ResultsManager) error {
	oLog := logs.FromContext(ctx)

	oaiClient, err := oai.NewClient(e.config)
	if err != nil {
		return errors.Wrapf(err, "Failed to create OpenAI client")
	}

	judge, err := NewJudge(oaiClient)

	if err != nil {
		return errors.Wrapf(err, "Failed to create Judge")
	}

	// Now iterate over the examples and process them.
	for _, example := range examples {
		log := oLog.WithValues("exampleId", example.GetId())

		// TODO(jeremy): Should we just read the row from the database and check if it exists and has been completed?
		// Finding the lastProcessed time and then using that seems like maybe its premature optimization? But maybe
		// since I wrote it might as well keep it.
		exampleTime := example.GetTime().AsTime()
		if exampleTime.Before(lastProcessedTime) || exampleTime == lastProcessedTime {
			log.V(logs.Debug).Info("Skipping example; already processed")
			continue
		}
		log.Info("Processing example")

		var processErr error

		uErr := manager.Update(ctx, example.GetId(), func(result *v1alpha1.EvalResult) error {
			processErr = e.processResult(ctx, result, example, client, logsClient, judge)
			// We need to return for the transaction to be committed.
			return nil
		})

		if processErr != nil {
			log.Error(processErr, "Failed to process example")
			// For now we abort on error to see what's going on.
			return processErr
		}

		if uErr != nil {
			log.Error(uErr, "Failed to update result")
			// For now we abort on error to see what's going on.
			return uErr
		}

		result, err := manager.Get(ctx, example.GetId())
		if err != nil {
			return errors.Wrapf(err, "Failed to get latest result for example %s", example.GetId())
		}

		if result.Error != "" {
			// Generating a completion failed for this example so we should keep going.
			// There won't be a blocklog to wait for.
			continue
		}

		if err := e.waitForBlockLog(ctx, result, logsClient); err != nil {
			log.Error(err, "Failed to wait for block log")
			// For now we abort on error to see what's going on.
			return errors.Wrapf(err, "Failed to get block log for example %s", example.GetId())
		}

		var ragErr error
		// Getting the bestRAG result depends on the trace having been processed so we run after waiting for the BlockLog
		uErr = manager.Update(ctx, example.GetId(), func(result *v1alpha1.EvalResult) error {
			ragErr = e.reconcileBestRAGResult(ctx, result, logsClient)
			return nil
		})

		if ragErr != nil {
			log.Error(ragErr, "Failed to reconcile best RAG result")
			// For now we abort on error to see what's going on.
			return ragErr
		}

		if uErr != nil {
			log.Error(uErr, "Failed to update result")
			// For now we abort on error to see what's going on.
			return uErr
		}

	}
	return nil
}

// processResult process the result. It is updated in place
func (e *Evaluator) processResult(ctx context.Context, result *v1alpha1.EvalResult, example *v1alpha1.EvalExample, client v1alpha1connect.AIServiceClient, logsClient logspbconnect.LogsServiceClient, judge *Judge) error {
	result.Example = example
	log := logs.FromContext(ctx).WithValues("exampleId", example.GetId())
	ctx = logr.NewContext(ctx, log)

	if err := runGenerate(ctx, result, client); err != nil {
		return err
	}

	if result.Error != "" {
		// Since an error occurred generating a completion for this example we can't continue to
		// process this example
		// We return nil because we want the evaluator to continue with other examples
		return nil
	}

	if err := runExecute(ctx, result, client); err != nil {
		return err
	}

	if err := judge.Score(ctx, result); err != nil {
		err := errors.Wrapf(err, "Failed to judge example %s", example.GetId())
		result.Error = err.Error()
		return err
	}

	return nil
}

// runGenerate runs the generate step for the example.
//
// runGenerate returns an error if there is a problem that should cause evaluation to abort rather than processing
// other examples (e.g. unable to contact the agent). If there is a problem generating a completion for this specific
// example then the result will be nil but result.Error will be set
func runGenerate(ctx context.Context, result *v1alpha1.EvalResult, client v1alpha1connect.AIServiceClient) error {
	log := logs.FromContext(ctx)
	// ID for the generate session
	genSessionID := ulid.GenerateID()

	// We need to send a session event to the agent to simulate the session starting.
	// This is because SessionStart event will contain the full context used with the execution
	logEventReq := &v1alpha1.LogEventsRequest{}
	logEventReq.Events = append(logEventReq.Events, &v1alpha1.LogEvent{
		Type:          v1alpha1.LogEventType_SESSION_START,
		ContextId:     genSessionID,
		SelectedIndex: result.Example.GetFullContext().GetSelected(),
	})

	_, err := client.LogEvents(ctx, connect.NewRequest(logEventReq))
	if err != nil {
		log.Error(err, "Failed to log events")
		// For now abort on error to see what's going on.
		return errors.Wrapf(err, "Failed to log events")
	}

	request := &v1alpha1.GenerateCellsRequest{
		Notebook:      result.Example.GetFullContext().GetNotebook(),
		SelectedIndex: result.Example.GetFullContext().GetSelected(),
	}

	resp, err := client.GenerateCells(ctx, connect.NewRequest(request))
	if err != nil {
		log.Error(err, "Failed to generate cells")
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			// TODO(https://github.com/jlewi/foyle/issues/257)
			// Currently GenerateCells returns a connect.Error if the completer can't generate a completion
			// because of too many tokens.
			if connect.CodeOf(err) == connect.CodeUnknown {
				result.Error = err.Error()
				// We return nil because the problem is specific to this example so the evaluator should move on
				// to other examples
				return nil
			} else {
				result.Error = err.Error()
				// Assume its a problem that could affect other examples so abort it.
				return err
			}
		} else {
			result.Error = err.Error()
			return err
		}
	}

	result.ActualCells = resp.Msg.GetCells()

	traceParent := resp.Header().Get(agent.TraceIDHeader)
	if traceParent == "" {
		return errors.New("GenerateCells response didn't contain traceparent header")
	}
	result.GenTraceId = traceParent

	// We need to close the generate session session.
	endEventsReq := &v1alpha1.LogEventsRequest{
		Events: []*v1alpha1.LogEvent{
			{
				ContextId: genSessionID,
				Type:      v1alpha1.LogEventType_SESSION_END,
			},
		},
	}

	_, err = client.LogEvents(ctx, connect.NewRequest(endEventsReq))
	if err != nil {
		log.Error(err, "Failed to log events")
		// For now abort on error to see what's going on.
		return errors.Wrapf(err, "Failed to log events")
	}
	return nil
}

func runExecute(ctx context.Context, result *v1alpha1.EvalResult, client v1alpha1connect.AIServiceClient) error {
	log := logs.FromContext(ctx)
	// We need to send a LOG event to the agent to simulate the cells being executed.
	executeEventReq := &v1alpha1.LogEventsRequest{}

	// Start a session to execute the cell
	execSessionID := ulid.GenerateID()

	if len(result.Example.ExpectedCells) != 1 {
		return errors.New("Expected cells isn't 1; How did this make it into the evaluation dataset? Shouldn't all examples in the eval set have 1 expected cell")
	}

	if len(result.ActualCells) < 1 {
		// In this case the LLM failed to generate a cell. There's no point sending an execution event because
		// There's no cellId to link the executed cell to the generation event.
		// Currently, Foyle doesn't have a way of learning when the LLM fails to generate a cell. Learning
		// only occurs if 1) Foyle generates a cell, 2) user edits cell 3) user executes the cell
		return nil
	}

	cell := result.Example.ExpectedCells[0]

	if cell.Kind != parserv1.CellKind_CELL_KIND_CODE {
		return errors.New("The expected cell in the example isn't of type CELL_KIND_CODE. How did this make it into the evaluation dataset? Shouldn't all examples in the eval set have 1 expected cell of type CELL_KIND_CODE")
	}

	actualID := converters.GetCellID(result.ActualCells[0])
	if actualID == "" {
		return errors.New("Actual cell ID is empty")
	}

	converters.SetCellID(cell, actualID)

	executeEventReq.Events = append(executeEventReq.Events, &v1alpha1.LogEvent{
		Type:      v1alpha1.LogEventType_SESSION_START,
		ContextId: execSessionID,
	})

	executeEventReq.Events = append(executeEventReq.Events, &v1alpha1.LogEvent{
		ContextId: execSessionID,
		Type:      v1alpha1.LogEventType_EXECUTE,
		Cells: []*parserv1.Cell{
			cell,
		},
		SelectedIndex: 0,
		SelectedId:    converters.GetCellID(cell),
	})

	executeEventReq.Events = append(executeEventReq.Events, &v1alpha1.LogEvent{
		Type:      v1alpha1.LogEventType_SESSION_END,
		ContextId: execSessionID,
	})

	if _, err := client.LogEvents(ctx, connect.NewRequest(executeEventReq)); err != nil {
		log.Error(err, "Failed to log events")
		result.Error = errors.Wrapf(err, "Failed to log events").Error()
		return errors.Wrapf(err, "Failed to log events")
	}
	return nil
}

func (e *Evaluator) waitForBlockLog(ctx context.Context, result *v1alpha1.EvalResult, client logspbconnect.LogsServiceClient) error {
	// We need to wait for the block log to be processed.
	// This is done to
	// 1. Increase the likelihood we have learned from the block
	// 2. To verify that the evaluator properly sends the data needed for the agent to learn from the block.
	log := logs.FromContext(ctx)
	if len(result.GetActualCells()) == 0 {
		// Since no cells were actually generated there won't be any blocklog to wait for.
		return nil
	}

	// TODO(jeremy): What should we do if there's more than 1 code cell?
	var codeCell *parserv1.Cell
	for _, cell := range result.GetActualCells() {
		if cell.Kind == parserv1.CellKind_CELL_KIND_CODE {
			codeCell = cell
			break
		}
	}

	if codeCell == nil {
		// Since there is no code cell there's no blockLog to fetch
		return nil
	}

	cellID := converters.GetCellID(codeCell)
	if cellID == "" {
		return errors.New("Cell ID is empty")
	}

	timeOut := time.Now().Add(3 * time.Minute)

	var blockLog *logspb.BlockLog
	for time.Now().Before(timeOut) {

		resp, err := client.GetBlockLog(ctx, connect.NewRequest(&logspb.GetBlockLogRequest{
			Id: cellID,
		}))

		if err != nil {
			log.Info("Failed to get block log", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		blockLog = resp.Msg.GetBlockLog()
		if blockLog.ExecutedBlock == nil || blockLog.GeneratedBlock == nil {
			log.Info("Block log isn't ready yet")
			time.Sleep(5 * time.Second)
			continue
		}

		if blockLog.GeneratedBlock.GetContents() != result.ActualCells[0].Value {
			return errors.Errorf("BlockLog generated block doesn't match actual cell. This means the result of GenerateCells returned to the evaluator doesn't match the result that the Agent read from the BlockLogs and stored in its BlockLog; want: %s; got %s", result.ActualCells[0].Value, blockLog.GeneratedBlock.GetContents())
		}

		return nil
	}

	return errors.New("Timed out waiting for block log. This could indicate we aren't properly sending the events needed to generate a BlockLog suitable for learning.")
}

func (e *Evaluator) reconcileBestRAGResult(ctx context.Context, evalResult *v1alpha1.EvalResult, client logspbconnect.LogsServiceClient) error {
	if evalResult.GenTraceId == "" {
		return errors.WithStack(errors.New("GenTraceId is empty"))
	}

	timeOut := time.Now().Add(3 * time.Minute)
	var genTrace *logspb.Trace
	for {
		if time.Now().After(timeOut) {
			return errors.Errorf("Timed out waiting for traceId to be ready; traceId: %s", evalResult.GenTraceId)
		}

		resp, err := client.GetTrace(ctx, connect.NewRequest(&logspb.GetTraceRequest{
			Id: evalResult.GenTraceId,
		}))

		if err == nil {
			genTrace = resp.Msg.GetTrace()
			break
		}

		// Check if the error is a "not found" error
		// We want to retry if the trace isn't found because it might not have been processed yet
		if connect.CodeOf(err) != connect.CodeNotFound {
			// If it's any other error, consider it a permanent error
			return errors.Wrapf(err, "Failed to get trace %s", evalResult.GenTraceId)
		}

		time.Sleep(5 * time.Second)
	}

	// TODO(jeremy): Should we update EvalResult to indicate the failure
	// What should we do if the experiment doesn't involve learning

	if genTrace == nil {
		return errors.WithStack(errors.Errorf("Trace %s is nil", evalResult.GenTraceId))
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
			if evalResult.BestRagResult == nil {
				evalResult.BestRagResult = ragResult
				continue
			}

			if evalResult.BestRagResult.Score < ragResult.Score {
				evalResult.BestRagResult = ragResult
			}
		}
	}

	return nil
}

// isSortedByTimeDescending checks if the slice is sorted by Time in descending order
func isSortedByTimeDescending(slice []*v1alpha1.EvalResult) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1].Example.GetTime().AsTime().Before(slice[i].Example.GetTime().AsTime()) {
			return false
		}
	}
	return true
}

// listEvalFiles returns a list of the all the binary protobuf files in the directory evalDir.
func listEvalFiles(ctx context.Context, evalDir string) ([]string, error) {
	examples := make([]string, 0, 100)
	if evalDir == "" {
		return examples, errors.Wrapf(errors.New("evalDir is empty"), "evalDir is empty")
	}
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

// getLastProcessedTime returns the time of the last processed example
func getLastProcessedTime(ctx context.Context, manager *ResultsManager) (time.Time, error) {
	// Default the time of the lastProcessedEval example to some time in the past.
	// This way all examples should be after it and get reprocessed
	lastProcessedTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	alreadyProcessed, _, err := manager.ListResults(ctx, nil, 10)

	if err != nil {
		return lastProcessedTime, errors.Wrapf(err, "Failed to list already processed results")
	}

	if len(alreadyProcessed) == 0 {
		return lastProcessedTime, nil
	}

	if !isSortedByTimeDescending(alreadyProcessed) {
		return lastProcessedTime, errors.New("Results aren't sorted by time in descending order")
	}

	return alreadyProcessed[0].Example.GetTime().AsTime(), nil
}

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
