package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/logs"
	"net/http"

	"connectrpc.com/connect"

	"github.com/cockroachdb/pebble"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// CrudHandler is a handler for CRUD operations on log entries
type CrudHandler struct {
	cfg      config.Config
	blocksDB *pebble.DB
	tracesDB *pebble.DB
	logFiles []string
}

func NewCrudHandler(cfg config.Config, blocksDB *pebble.DB, tracesDB *pebble.DB) (*CrudHandler, error) {
	// Get a list of log files. We don't do log rotation so we only need to fetch this at start
	logFiles, err := getLogFilesSorted(cfg.GetRawLogDir())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get log files")
	}

	return &CrudHandler{
		cfg:      cfg,
		blocksDB: blocksDB,
		tracesDB: tracesDB,
		logFiles: logFiles,
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
		// Assume its a not found error.
		return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "Failed to get trace with id %s", getReq.GetId()))
	}

	return connect.NewResponse(&logspb.GetTraceResponse{Trace: trace}), nil
}

func (h *CrudHandler) GetLLMLogs(ctx context.Context, request *connect.Request[logspb.GetLLMLogsRequest]) (*connect.Response[logspb.GetLLMLogsResponse], error) {
	log := logs.FromContext(ctx)
	getReq := request.Msg
	if getReq.GetTraceId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("No traceID provided"))
	}

	logsToSearch := h.logFiles
	if getReq.GetLogFile() != "" {
		logsToSearch = []string{getReq.GetLogFile()}
	}

	var logEntry *AnthropicLog
	found := false
	// Search through the logs until we find the entry
	for _, logFile := range logsToSearch {
		var err error
		logEntry, err = readAnthropicLog(ctx, getReq.GetTraceId(), logFile)
		if logEntry == nil {
			continue
		}
		if logEntry.Request != nil || logEntry.Response != nil {
			found = true
			break
		}
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to search logFile: %s", logFile))
		}
	}

	if !found {
		return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("No log entry found for traceID: %s in logFiles: %s", getReq.GetTraceId(), logsToSearch))
	}

	resp := &logspb.GetLLMLogsResponse{}
	resp.RequestHtml = renderAnthropicRequest(logEntry.Request)
	resp.ResponseHtml = renderAnthropicResponse(logEntry.Response)

	reqB, err := json.Marshal(logEntry.Request)
	if err != nil {
		log.Error(err, "Failed to marshal request")
		resp.RequestJson = fmt.Sprintf("Failed to marshal request; error %+v", err)
	} else {
		resp.RequestJson = string(reqB)
	}

	resB, err := json.Marshal(logEntry.Response)
	if err != nil {
		log.Error(err, "Failed to marshal response")
		resp.ResponseJson = fmt.Sprintf("Failed to marshal response; error %+v", err)
	} else {
		resp.ResponseJson = string(resB)
	}

	return connect.NewResponse(resp), nil
}

func (h *CrudHandler) GetBlockLog(c *gin.Context) {
	log := zapr.NewLogger(zap.L())

	id := c.Param("id")

	log = log.WithValues("id", id)
	bLog := &logspb.BlockLog{}
	if err := dbutil.GetProto(h.blocksDB, id, bLog); err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No block with id %s", id)})
			return
		} else {
			log.Error(err, "Failed to read block with id", "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read block with id %s; error %+v", id, err)})
			return
		}
	}

	b, err := protojson.Marshal(bLog)
	if err != nil {
		log.Error(err, "Failed to marshal block log", "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to marshal block with id %s; error %+v", id, err)})
		return

	}

	// Use the id to fetch or manipulate the resource
	// For now, we'll just echo it back
	if _, err := c.Writer.Write(b); err != nil {
		log.Error(err, "Failed to write response", "id", id)
	}
}
