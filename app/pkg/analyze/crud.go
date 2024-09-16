package analyze

import (
	"context"
	"sort"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/logs"

	"github.com/cockroachdb/pebble"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// CrudHandler is a handler for CRUD operations on log entries
type CrudHandler struct {
	cfg      config.Config
	blocksDB *pebble.DB
	tracesDB *pebble.DB
	analyzer *Analyzer
}

func NewCrudHandler(cfg config.Config, blocksDB *pebble.DB, tracesDB *pebble.DB, analyzer *Analyzer) (*CrudHandler, error) {
	return &CrudHandler{
		cfg:      cfg,
		blocksDB: blocksDB,
		tracesDB: tracesDB,
		analyzer: analyzer,
	}, nil
}

func (h *CrudHandler) GetTrace(ctx context.Context, request *connect.Request[logspb.GetTraceRequest]) (*connect.Response[logspb.GetTraceResponse], error) {
	getReq := request.Msg
	if getReq.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("No traceID provided"))
	}

	trace := &logspb.Trace{}
	err := dbutil.GetProto(h.tracesDB, getReq.GetId(), trace)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "Failed to get trace with id %s", getReq.GetId()))
		} else {
			log := logs.FromContext(ctx)
			log.Error(err, "Failed to read trace with id", "id", getReq.GetId())
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to get Trace with id %s", getReq.GetId()))
		}
	}

	return connect.NewResponse(&logspb.GetTraceResponse{Trace: trace}), nil
}

func (h *CrudHandler) GetLLMLogs(ctx context.Context, request *connect.Request[logspb.GetLLMLogsRequest]) (*connect.Response[logspb.GetLLMLogsResponse], error) {
	log := logs.FromContext(ctx)
	getReq := request.Msg
	if getReq.GetTraceId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("No traceID provided"))
	}

	logFiles, err := findLogFiles(ctx, h.cfg.GetLogDir())
	if err != nil {
		log.Error(err, "Failed to find log files")
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "Failed to find log files"))
	}

	// Sort the slice in descending order
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i] > logFiles[j]
	})

	// We loop over all the logFiles until we find it which is not efficient.
	for _, logFile := range logFiles {
		resp, err := readLLMLog(ctx, getReq.GetTraceId(), logFile)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to get LLM call log for trace id %s; logFile: %s", getReq.GetTraceId(), getReq.GetLogFile()))
		}
		if resp != nil {
			return connect.NewResponse(resp), nil
		}
	}

	return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("No log file found for traceID %v", getReq.GetTraceId()))
}

func (h *CrudHandler) GetBlockLog(ctx context.Context, request *connect.Request[logspb.GetBlockLogRequest]) (*connect.Response[logspb.GetBlockLogResponse], error) {
	log := zapr.NewLogger(zap.L())
	getReq := request.Msg

	if getReq.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("No blocklog id provided"))
	}

	bLog := &logspb.BlockLog{}
	if err := dbutil.GetProto(h.blocksDB, getReq.GetId(), bLog); err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "No BlockLog with id %s was found", getReq.GetId()))
		} else {
			log.Error(err, "Failed to read block with id", "id", getReq.GetId())
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to get BlockLog with id %s", getReq.GetId()))
		}
	}

	return connect.NewResponse(&logspb.GetBlockLogResponse{BlockLog: bLog}), nil
}

func (h *CrudHandler) Status(ctx context.Context, request *connect.Request[logspb.GetLogsStatusRequest]) (*connect.Response[logspb.GetLogsStatusResponse], error) {
	response := &logspb.GetLogsStatusResponse{
		Watermark: h.analyzer.GetWatermark(),
	}

	return connect.NewResponse(response), nil
}
