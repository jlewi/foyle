package logsviewer

import (
	"context"
	"encoding/json"
	"github.com/jlewi/foyle/app/api"
	"io"
	"net/http"
)

// LogsClient client for the logs service.
// TODO(jeremy): If we used protos and buf to generate the client we could avoid handrolling the clients?
type LogsClient struct {
	Endpoint string
}

func (c *LogsClient) GetBlockLog(ctx context.Context, blockID string) (*api.BlockLog, error) {
	// For now don't use the cache always dynamically generate the trace.
	r, err := http.Get(c.Endpoint + "api/blocklogs/" + blockID)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	block := &api.BlockLog{}
	if err := json.Unmarshal(b, block); err != nil {
		return nil, err
	}
	return block, nil
}
