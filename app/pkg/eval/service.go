package eval

import (
	"context"
	"path/filepath"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"connectrpc.com/connect"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// EvalServer is the server that implements the Eval service interface.
// This is used to make results available to the frontend.
type EvalServer struct {
	manager  *ResultsManager
	tracesDB *pebble.DB
	config   config.Config
}

func NewEvalServer(cfg config.Config, tracesDB *pebble.DB) *EvalServer {
	return &EvalServer{
		config:   cfg,
		tracesDB: tracesDB,
	}
}

func (s *EvalServer) List(
	ctx context.Context,
	req *connect.Request[v1alpha1.EvalResultListRequest],
) (*connect.Response[v1alpha1.EvalResultListResponse], error) {
	log := logs.FromContext(ctx)

	if req.Msg.GetDatabase() == "" {
		err := connect.NewError(connect.CodeInvalidArgument, errors.New("Request is missing database"))
		log.Error(err, "Invalid EvalResultListRequest")
		return nil, err
	}

	db, err := pebble.Open(req.Msg.GetDatabase(), &pebble.Options{})
	if err != nil {
		log.Error(err, "Failed to open database")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer helpers.DeferIgnoreError(db.Close)

	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer iter.Close()

	results := &v1alpha1.EvalResultListResponse{
		Items: make([]*v1alpha1.EvalResult, 0, 100),
	}

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			log.Error(err, "Failed to read value for key", "key", string(key))
			continue
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			log.Error(err, "Failed to unmarshal value for", "key", string(key))
			continue
		}
		results.Items = append(results.Items, result)
	}

	res := connect.NewResponse(results)
	res.Header().Set("Eval-Version", "v1alpha1")
	return res, nil
}

func (s *EvalServer) GetEvalResult(
	ctx context.Context,
	req *connect.Request[v1alpha1.GetEvalResultRequest],
) (*connect.Response[v1alpha1.GetEvalResultResponse], error) {
	log := logs.FromContext(ctx)

	if s.manager == nil {
		if err := s.tryToLoadResultsManager(); err != nil {
			// TOOD(jeremy): We should probably allow the database to be specified in the request
			log.Error(err, "ResultsManager not initialized")
			err := connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to load ResultsManager"))
			return nil, err
		}
	}

	if s.tracesDB == nil {
		err := connect.NewError(connect.CodeInternal, errors.New("TracesDB not initialized"))
		log.Error(err, "TracesDB not initialized")
		return nil, err
	}

	if req.Msg.GetId() == "" {
		err := connect.NewError(connect.CodeInvalidArgument, errors.New("Request is missing id"))
		log.Error(err, "Invalid GetEvalResultRequest")
		return nil, err
	}

	result, err := s.manager.Get(ctx, req.Msg.GetId())

	if err != nil {
		log.Info("Failed to get result", "id", req.Msg.GetId(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if result.GetGenTraceId() == "" {
		// TODO(jeremy): We should probably generate a report and the report should report the lack of a trace.
		return nil, connect.NewError(connect.CodeNotFound, errors.New("No trace ID found"))
	}

	trace := &logspb.Trace{}
	if err := dbutil.GetProto(s.tracesDB, result.GetGenTraceId(), trace); err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "Failed to get trace with id %s", result.GetGenTraceId()))
		} else {
			log := logs.FromContext(ctx)
			log.Error(err, "Failed to read trace with id", "id", result.GetGenTraceId())
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to get Trace with id %s", result.GetGenTraceId()))
		}
	}

	reportHTML, err := buildEvalReport(ctx, result, trace)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "Failed to build report"))
	}

	response := &v1alpha1.GetEvalResultResponse{
		ReportHTML: reportHTML,
	}

	res := connect.NewResponse(response)
	res.Header().Set("Eval-Version", "v1alpha1")
	return res, nil
}

func (s *EvalServer) AssertionTable(
	ctx context.Context,
	req *connect.Request[v1alpha1.AssertionTableRequest],
) (*connect.Response[v1alpha1.AssertionTableResponse], error) {
	log := logs.FromContext(ctx)

	if req.Msg.GetDatabase() == "" {
		err := connect.NewError(connect.CodeInvalidArgument, errors.New("Request is missing database"))
		log.Error(err, "Invalid EvalResultListRequest")
		return nil, err
	}

	db, err := pebble.Open(req.Msg.GetDatabase(), &pebble.Options{})
	if err != nil {
		log.Error(err, "Failed to open database")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer helpers.DeferIgnoreError(db.Close)

	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer iter.Close()

	results := &v1alpha1.AssertionTableResponse{
		Rows: make([]*v1alpha1.AssertionRow, 0, 100),
	}

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			log.Error(err, "Failed to read value for key", "key", string(key))
			continue
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			log.Error(err, "Failed to unmarshal value for", "key", string(key))
			continue
		}

		row, err := toAssertionRow(result)
		if err != nil {
			// TODO(jeremy): Should we put this in the response
			log.Error(err, "Failed to convert to assertion row", "key", string(key))
			continue
		}
		results.Rows = append(results.Rows, row)
	}

	res := connect.NewResponse(results)
	res.Header().Set("Eval-Version", "v1alpha1")
	return res, nil
}

func (s *EvalServer) tryToLoadResultsManager() error {
	// This is a bit of a hack to try have the agent automatically load the results manager when the database
	// isn't specified in the request.
	if s.manager != nil {
		return nil
	}

	log := zapr.NewLogger(zap.L())
	dbFile := filepath.Join(s.config.GetConfigDir(), "results.sqlite")
	log.Info("Opening results manager", "dbFile", dbFile)
	manager, err := openResultsManager(dbFile)
	if err != nil {
		return err
	}
	s.manager = manager
	return nil
}

func toAssertionRow(result *v1alpha1.EvalResult) (*v1alpha1.AssertionRow, error) {
	log := zapr.NewLogger(zap.L())

	row := &v1alpha1.AssertionRow{
		Id:          result.Example.GetId(),
		ExampleFile: result.GetExample().FullContext.NotebookUri,
	}

	doc, err := converters.NotebookToDoc(result.GetExample().GetFullContext().GetNotebook())

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to convert notebook to doc")
	}

	actualDoc, err := converters.NotebookToDoc(&parserv1.Notebook{
		Cells: result.GetActualCells(),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to convert actual cells to doc")
	}

	row.DocMd = docs.DocToMarkdown(doc)
	row.AnswerMd = docs.DocToMarkdown(actualDoc)

	for _, a := range result.GetAssertions() {
		switch a.Name {
		case CodeAfterMarkdownName:
			row.CodeAfterMarkdown = a.GetResult()
		case OneCodeCellName:
			row.OneCodeCell = a.GetResult()
		case EndsWithCodeCellName:
			row.EndsWithCodeCell = a.GetResult()
		default:
			log.Info("Unknown assertion", "name", a.Name)
		}
	}

	return row, nil
}
