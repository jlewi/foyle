package logsviewer

import (
	"context"
	"io"
	"net/http"
	"strings"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/go-logr/zapr"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	defaultClient *LogsClient
)

// LogsClient client for the logs service.
// TODO(jeremy): If we used protos and buf to generate the client we could avoid handrolling the clients?
type LogsClient struct {
	Endpoint string
}

func (c *LogsClient) GetBlockLog(ctx context.Context, blockID string) (*logspb.BlockLog, error) {
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
	block := &logspb.BlockLog{}
	if err := protojson.Unmarshal(b, block); err != nil {
		return nil, err
	}
	return block, nil
}

func GetClient() *LogsClient {
	if defaultClient == nil {
		log := zapr.NewLogger(zap.L())
		// N.B. I think using EndpointEnvVar is better using Window().Location().Href because of how we would deal with
		// Reverse proxies. If we're behind some sort of reverse proxy we'd probably want the server to set the
		// appropriate baseURL
		endpoint := app.Getenv(EndpointEnvVar)
		if endpoint == "" {
			log.Error(errors.New("EndpointEnvVar is not set"), "Failed to create logsclient")
		}
		if !strings.HasSuffix(endpoint, "/") {
			endpoint = endpoint + "/"
		}
		log.Info("Creating logs client", "endpoint", endpoint)
		defaultClient = &LogsClient{
			Endpoint: endpoint,
		}
	}
	return defaultClient
}
