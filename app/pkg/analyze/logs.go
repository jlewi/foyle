package analyze

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
	"io"
	"os"
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

// readAnthropicRequest reads an Anthropic request from a log file
//
// N.B. If there are multiple requests as part of the same trace then only the last request will be returned.
// TODO(jeremy): Ideally we'd join the request with its response and return the one that succeeded. The reason
// There might be multiple is because context exceeded length; in which case only one request which has been
// sufficiently shortened will have an actual response.
func readAnthropicLog(ctx context.Context, traceId string, logFile string) (*AnthropicLog, error) {
	log := logs.FromContext(ctx)
	file, err := os.Open(logFile)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "Failed to open file %s", logFile))
	}
	d := json.NewDecoder(file)

	aLog := &AnthropicLog{
		TraceID: traceId,
		LogFile: logFile,
	}
	req := &anthropic.MessagesRequest{}
	resp := &anthropic.MessagesResponse{}
	for {
		entry := &api.LogEntry{}
		if err := d.Decode(entry); err != nil {
			if err == io.EOF {
				return aLog, nil
			}
			log.Error(err, "Failed to decode log entry")
		}
		if entry.TraceID() != traceId {
			continue
		}

		if entry.Request() != nil {
			// TODO(jeremy): We should check what the model provider is using the function argument and parse
			if err := json.Unmarshal(entry.Request(), req); err != nil {
				// TODO(jeremy): Should we include the error in the response?
				log.Error(err, "Failed to unmarshal request")
				continue
			} else {
				aLog.Request = req
				req = &anthropic.MessagesRequest{}
			}
		}

		if entry.Response() != nil {
			if err := json.Unmarshal(entry.Response(), resp); err != nil {
				// TODO(jeremy): Should we include the error in the response?
				log.Error(err, "Failed to unmarshal response")
				continue
			} else {
				aLog.Response = resp
				resp = &anthropic.MessagesResponse{}
			}
		}
	}
}
