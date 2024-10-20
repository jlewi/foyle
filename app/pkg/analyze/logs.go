package analyze

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

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

	resp := &logspb.GetLLMLogsResponse{}

	provider := api.ModelProviderUnknown
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
		isMatch := false

		if matchers.IsOAIComplete(entry.Function()) {
			provider = api.ModelProviderOpenAI
			isMatch = true
		}

		if strings.HasSuffix(entry.Function(), "anthropic.(*Completer).Complete") {
			provider = api.ModelProviderAnthropic
			isMatch = true
		}

		// If tis not a matching request ignore it.
		if !isMatch {
			continue
		}
		if reqBytes := entry.Request(); reqBytes != nil {
			resp.RequestJson = string(reqBytes)
		}

		if resBytes := entry.Response(); resBytes != nil {
			resp.ResponseJson = string(resBytes)
		}

		// Since we have read the request and response less
		// This isn't a great implementation because we will end up reading all the logs if for some reason
		// The logs don't have the entries.
		if resp.RequestJson != "" && resp.ResponseJson != "" {
			break
		}
	}

	if err := renderHTML(resp, provider); err != nil {
		log.Error(err, "Failed to render HTML")
	}
	return resp, nil
}
