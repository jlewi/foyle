package analyze

import (
	"context"
	"encoding/json"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"io"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
)

// AnthropicLog represents a log entry for an Anthropic request
// N.B. we should consider making this a proto. The only reason its not is because our client library uses go lang
// structs so we'd have to define a proto with the JSON equivalent.
type AnthropicLog struct {
	TraceID  string
	LogFile  string
	Request  *anthropic.MessagesRequest
	Response *anthropic.MessagesResponse
}

// readLLMLog tries to fetch the raw LLM request/response from the log
func readLLMLog(ctx context.Context, traceId string, logFile string) (*logspb.GetLLMLogsResponse, error) {
	log := logs.FromContext(ctx)
	file, err := os.Open(logFile)

	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "Failed to open file %s", logFile))
	}
	d := json.NewDecoder(file)

	for {
		entry := &api.LogEntry{}
		if err := d.Decode(entry); err != nil {
			if err == io.EOF {
				return nil, nil
			}
			log.Error(err, "Failed to decode log entry")
		}
		if entry.TraceID() != traceId {
			continue
		}
		if strings.HasSuffix(entry.Function(), "anthropic.(*Completer).Complete") {
			aLog, err := entryToAnthropicLog(ctx, entry)
			if err != nil {
				return nil, err
			}
			if aLog == nil {
				return nil, errors.Wrapf(err, "Failed to parse logEntry for LLM call for trace %s", traceId)
			}

			reqJson, err := json.Marshal(aLog.Request)
			if err != nil {
				log.Error(err, "Failed to marshal request")
				reqJson = []byte("Failed to marshal request")
			}

			respJson, err := json.Marshal(aLog.Response)
			if err != nil {
				log.Error(err, "Failed to marshal response")
				respJson = []byte("Failed to marshal response")
			}

			resp := &logspb.GetLLMLogsResponse{
				RequestHtml:  renderAnthropicRequest(aLog.Request),
				ResponseHtml: renderAnthropicResponse(aLog.Response),
				RequestJson:  string(reqJson),
				ResponseJson: string(respJson),
			}
			return resp, nil
		}
	}
}

// entryToAnthropicLog parses an Anthropic request log message
//
// N.B. If there are multiple requests as part of the same trace then only the last request will be returned.
// TODO(jeremy): Ideally we'd join the request with its response and return the one that succeeded. The reason
// There might be multiple is because context exceeded length; in which case only one request which has been
// sufficiently shortened will have an actual response.
func entryToAnthropicLog(ctx context.Context, entry *api.LogEntry) (*AnthropicLog, error) {
	log := logs.FromContext(ctx)

	aLog := &AnthropicLog{
		Request:  &anthropic.MessagesRequest{},
		Response: &anthropic.MessagesResponse{},
	}
	reqBytes := entry.Request()
	if reqBytes != nil {
		if err := json.Unmarshal(reqBytes, aLog.Request); err != nil {
			// TODO(jeremy): Should we include the error in the response?
			log.Error(err, "Failed to unmarshal request")
		}
	}

	respBytes := entry.Response()
	if respBytes != nil {
		if err := json.Unmarshal(respBytes, aLog.Response); err != nil {
			// TODO(jeremy): Should we include the error in the response?
			log.Error(err, "Failed to unmarshal response")
		}
	}

	return aLog, nil
}
