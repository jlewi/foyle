package agent

import (
	"context"
	"io"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/jlewi/foyle/app/pkg/llms"

	"github.com/jlewi/foyle/app/pkg/learn"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
)

const (
	maxTries = 3
	// MaxDocChars is an upper limit for the number of characters to include in prompts to avoid hitting
	// OpenAI's context length limits. This can be an upper bound because if we get a context length exceeded
	// error the code will automatically try to shrink the document even further.
	// We use the heuristic 1 token ~ 2 characters
	// We are currently using GPT3.5 which has a context window of 16385 tokens.
	// (https://platform.openai.com/docs/models/gpt-3-5-turbo)
	// If we use 50% of that's 16000 characters.
	MaxDocChars = 16000
	temperature = 0.9
)

// Agent is the agent.
type Agent struct {
	v1alpha1.UnimplementedGenerateServiceServer
	v1alpha1connect.UnimplementedAIServiceHandler
	completer llms.Completer
	config    config.Config
	db        *learn.InMemoryExampleDB
}

func NewAgent(cfg config.Config, completer llms.Completer, inMemoryExampleDB *learn.InMemoryExampleDB) (*Agent, error) {
	if cfg.Agent == nil {
		return nil, errors.New("Configuration is missing AgentConfig; configuration must define the agent field.")
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Creating agent", "config", cfg.Agent)

	if completer == nil {
		return nil, errors.New("Completer is required")
	}
	if cfg.Agent.RAG != nil && cfg.Agent.RAG.Enabled {
		if inMemoryExampleDB == nil {
			return nil, errors.New("RAG is enabled but learn is nil; learn must be set to use RAG")
		}
		log.Info("RAG is enabled")
	} else {
		inMemoryExampleDB = nil
	}

	return &Agent{
		completer: completer,
		config:    cfg,
		db:        inMemoryExampleDB,
	}, nil
}

func (a *Agent) Generate(ctx context.Context, req *v1alpha1.GenerateRequest) (*v1alpha1.GenerateResponse, error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	traceId := span.SpanContext().TraceID()
	log = log.WithValues("traceId", traceId, "evalMode", a.config.EvalMode())
	ctx = logr.NewContext(ctx, log)

	var examples []*v1alpha1.Example
	if a.config.UseRAG() {
		var err error
		examples, err = a.db.GetExamples(ctx, req.Doc, a.config.RagMaxResults())
		if err != nil {
			// Fail gracefully; keep going without examples
			log.Error(err, "Failed to get examples")
			examples = nil
		}
	}

	log.Info("Agent.Generate", zap.Object("request", req))
	blocks, err := a.completeWithRetries(ctx, req, examples)
	if err != nil {
		// TODO(jeremy): Should we set a status code?
		log.Error(err, "Agent.Generate failed to generate completions")
		return nil, err
	}

	// Attach block ids to any blocks generated.
	// N.B. This is kind of a last resort to make sure all blocks have an ID set. In general, we want to set blockIds
	// earlier in the processing pipeline so that any log messages involving blocks has block ids set. BlockIDs
	// should get set in parseResponse. When block Ids are first set.
	blockIds, err := docs.SetBlockIds(blocks)
	if err != nil {
		log.Error(err, "Agent.Generate, failed to set block ids", "blocks", blocks, "blockIds", blockIds)
	} else {
		log.Info("Agent.Generate returning blocks", "blockIds", blockIds)
	}

	resp := &v1alpha1.GenerateResponse{
		Blocks:  blocks,
		TraceId: traceId.String(),
	}

	log.Info("Agent.Generate returning response", zap.Object("response", resp))
	return resp, nil
}

func (a *Agent) completeWithRetries(ctx context.Context, req *v1alpha1.GenerateRequest, examples []*v1alpha1.Example) ([]*v1alpha1.Block, error) {
	log := logs.FromContext(ctx)

	t := docs.NewTailer(req.Doc.GetBlocks(), MaxDocChars)

	exampleArgs := make([]Example, 0, len(examples))
	for _, example := range examples {
		exampleArgs = append(exampleArgs, Example{
			Input:  docs.DocToMarkdown(example.Query),
			Output: docs.BlocksToMarkdown(example.Answer),
		})
	}
	for try := 0; try < maxTries; try++ {
		args := promptArgs{
			Document: t.Text(),
			Examples: exampleArgs,
		}

		var sb strings.Builder
		if err := promptTemplate.Execute(&sb, args); err != nil {
			return nil, errors.Wrapf(err, "Failed to execute prompt template")
		}

		blocks, err := a.completer.Complete(ctx, systemPrompt, sb.String())

		if err != nil {
			if oai.ErrorIs(err, oai.ContextLengthExceededCode) {
				log.Info("OpenAI:ContextLengthExceeded", "err", err)
				if !t.Shorten() {
					return nil, errors.Wrapf(err, "the document can't be shortened any further to fit within the context window")
				}
				continue
			}
			// TODO(jeremy): Should we surface the error to the user as blocks in the notebook
			return nil, errors.Wrapf(err, "CreateChatCompletion failed")
		}

		return blocks, nil
	}
	err := errors.Errorf("Failed to generate a chat completion after %d tries", maxTries)
	log.Error(err, "Failed to generate a chat completion", "maxTries", maxTries)
	return nil, err
}

func (a *Agent) StreamGenerate(ctx context.Context, stream *connect.BidiStream[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]) error {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	traceId := span.SpanContext().TraceID()
	log = log.WithValues("traceId", traceId, "evalMode", a.config.EvalMode())
	log.Info("Agent.StreamGenerate")
	notebookUri := ""
	var selectedCell int32
	reqCount := 0

	// statusChannel is used by the two go routines (i.e the request handler and response handler) to signal
	// to main routine that they have exited and therefore the request should be terminated.
	statusChan := make(chan *status.Status, 2)

	// Create a channel that will be used to signal to the receiver to generate a completion
	// The maximum number of events inflight should be 1 but we use 20 just to get some buffer
	trigger := make(chan bool, 20)

	// lastDoc is the serialized version of the most recent document. It will be non empty if there is a version
	// of the document awaiting processing.
	var pendingDoc *v1alpha1.Doc
	mu := &sync.Mutex{}

	// Start a thread to asynchronously generate completions.
	// We will generate one completion at a time. pendingDoc is used to enqueue a document to be processed.
	// if pendingDoc is nil then there is no updated document waiting to be processed.
	generateCtx, generateCancelFunc := context.WithCancel(ctx)
	// Ensure cancelFunc is called when this function returns; ensures we terminate the go routine
	defer generateCancelFunc()
	go func(ctx context.Context) {
		for {
			select {
			case <-trigger:
				log.Info("Received trigger signal")
				generateRequest := func() *v1alpha1.GenerateRequest {
					mu.Lock()
					defer mu.Unlock()
					if pendingDoc == nil {
						return nil
					}
					// This should be safe because each time we update pendingDoc we update it to point to
					// a new doc object. So the other thread won't be modifying the doc pendingDoc points to
					r := &v1alpha1.GenerateRequest{
						Doc: pendingDoc,
					}
					pendingDoc = nil
					return r
				}()
				if generateRequest == nil {
					// There is no pending document to process
					continue
				}

				response, err := a.createCompletion(ctx, generateRequest, notebookUri, selectedCell)

				if err != nil {
					log.Error(err, "createCompletion failed")
					// TODO(jeremy): Instead of terminating the request should we just try to recover on
					// The next request?
					statusChan <- status.Newf(codes.Internal, err.Error())
					return
				}

				log.V(logs.Debug).Info("Sending response", zap.Object("response", response))
				if err := stream.Send(response); err != nil {
					log.Error(err, "Failed to send response")
					// TODO(jeremy): Should we be using connect codes and routines? e.g.
					// connect.NewError(
					statusChan <- status.Newf(codes.Internal, "failed to send response; %v", err)
					return
				}

			case <-ctx.Done():
				log.Info("Context cancelled; stopping completion generation")
				statusChan <- status.New(codes.Canceled, "Stream context canceled")
				return
			}
		}
	}(generateCtx)

	readeCtx, readCancelFunc := context.WithCancel(ctx)
	// Ensure cancelFunc is called when this function returns; ensures we terminate the go routine
	defer readCancelFunc()
	go func(ctx context.Context) {
		// Start  a thread to receive requests from the client
		// Keep track of the doc
		var doc *v1alpha1.Doc

		resultChan := make(chan *v1alpha1.StreamGenerateRequest, 2)
		errChan := make(chan error, 2)

		for {

			// N.B. This go function reads a single request and adds it to the channel.
			// This way we can have a select statement to detect if the context gets cancelled before the read occurs.
			go func() {
				result, err := stream.Receive()
				if err != nil {
					errChan <- err
				} else {
					resultChan <- result
				}
			}()

			var err error
			var req *v1alpha1.StreamGenerateRequest
			select {
			case <-ctx.Done():
				log.Info("Context cancelled; stop listening for requests")
				return
			case err = <-errChan:
			case req = <-resultChan:
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					// The client has closed the stream
					log.Info("Client closed the stream")
					statusChan <- status.New(codes.OK, "Client closed the stream")
					return
				}
				if errors.Is(err, context.Canceled) {
					// The context was cancelled (e.g., client disconnected)
					log.Info("Stream context cancelled")
					statusChan <- status.New(codes.Canceled, "Stream context canceled")
					return
				}
				// Some other error occurred
				connectErr, ok := err.(*connect.Error)
				if ok && connectErr.Code() == connect.CodeDeadlineExceeded {
					// Streaming connections are expected to timeout because of the http timeout
					log.V(logs.Debug).Info("Streaming connection closed, deadline exceeded")
				} else {
					log.Error(err, "Error receiving from stream")
				}
				statusChan <- status.New(codes.Canceled, "Client closed the stream")
				return
			}
			reqCount++

			isValidErr := func() error {
				if reqCount == 1 {
					if req.GetFullContext() == nil {
						return status.Errorf(codes.InvalidArgument, "First request must have a full context")
					}
					log.Info("Received full context", "context", req.GetFullContext())
					if req.GetFullContext().GetNotebookUri() == "" {
						return status.Errorf(codes.InvalidArgument, "First request must have a notebookUri")
					}
					if req.GetFullContext().GetSelected() < 0 {
						return status.Errorf(codes.InvalidArgument, "First request must have a selected cell")
					}

					doc, err = converters.NotebookToDoc(req.GetFullContext().GetNotebook())
					if err != nil {
						log.Error(err, "Failed to convert notebook to doc")
						return status.Errorf(codes.InvalidArgument, "Failed to convert notebook to doc")
					}
					notebookUri = req.GetFullContext().GetNotebookUri()
					selectedCell = req.GetFullContext().GetSelected()

					if int(selectedCell) >= len(doc.Blocks) {
						log.Error(errors.New("Invalid request"), "Selected cell is out of bounds", "selectedCell", selectedCell, "numCells", len(doc.Blocks))
						return status.Errorf(codes.InvalidArgument, "Selected cell is out of bounds: index %d; number of cells %d", selectedCell, len(doc.Blocks))
					}
				} else {
					if req.GetUpdate() == nil {
						return status.Errorf(codes.InvalidArgument, "Every request except the first one must have an update")
					}

					block, err := converters.CellToBlock(req.GetUpdate().GetCell())
					if err != nil {
						log.Error(err, "Failed to convert cell to block")
						return status.Errorf(codes.InvalidArgument, "Failed to convert cell to block")
					}
					doc.Blocks[selectedCell] = block
				}
				return nil
			}()

			if isValidErr != nil {
				log.Info("Request is invalid", "err", isValidErr)
				statusChan <- status.Convert(isValidErr)
				return
			}

			log.Info("Received request", zap.Object("request", req))
			// Serialize the doc and make it available for processing
			func() {
				mu.Lock()
				defer mu.Unlock()

				// We only need to send a trigger if pendingDoc was nil.
				// If its nonNil then we've already sent a trigger that hasn't been processed yet
				sendTrigger := pendingDoc == nil

				var ok bool
				pendingDoc, ok = proto.Clone(doc).(*v1alpha1.Doc)
				if !ok {
					log.Info("Failed to clone doc")
					statusChan <- status.New(codes.Internal, "Failed to clone doc")
					return
				}

				// Signal the completion generator to generate a completion
				if sendTrigger {
					log.Info("Sending trigger signal")
					trigger <- true
				}
			}()
		}
	}(readeCtx)

	select {
	// Terminate because the request got cancelled
	case <-ctx.Done():
		log.Info("Context cancelled; stopping streaming request", "err", ctx.Err())
		// Cancel functions will be called when this function returns
		return ctx.Err()
	case s := <-statusChan:
		return s.Err()

	}
}

// createCompletion is a helper function to create a single completion as part of a stream.
func (a *Agent) createCompletion(ctx context.Context, generateRequest *v1alpha1.GenerateRequest, notebookUri string, selectedCell int32) (*v1alpha1.StreamGenerateResponse, error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	traceId := span.SpanContext().TraceID()
	tp := tracer()
	// We need to generate a new ctx with a new trace ID because we want one trace per completion
	// We need to use withNewRoot because we want to make it a new trace and not rooted at the current one
	generateCtx, generateSpan := tp.Start(ctx, "StreamAgentGenerate", trace.WithNewRoot(), trace.WithAttributes(attribute.String("streamTraceID", traceId.String())))
	generateTraceId := generateSpan.SpanContext().TraceID()
	log = log.WithValues("traceId", generateTraceId, "streamTraceId", traceId.String())
	generateCtx = logr.NewContext(generateCtx, log)
	defer generateSpan.End()

	generateResponse, err := a.Generate(generateCtx, generateRequest)
	if err != nil {
		return nil, err
	}

	cells, err := converters.BlocksToCells(generateResponse.GetBlocks())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to convert blocks to cells")
	}

	response := &v1alpha1.StreamGenerateResponse{
		Cells:       cells,
		NotebookUri: notebookUri,
		InsertAt:    selectedCell + 1,
	}

	return response, nil
}

func (a *Agent) Status(ctx context.Context, req *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	traceId := span.SpanContext().TraceID()
	log = log.WithValues("traceId", traceId, "evalMode", a.config.EvalMode())
	log.Info("Agent.Simple")

	response := &v1alpha1.StatusResponse{
		Status: v1alpha1.AIServiceStatus_OK,
	}
	res := connect.NewResponse(response)
	res.Header().Set("AIService-Version", "v1alpha1")
	return res, nil
}
