package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/jlewi/monogo/networking"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type FakeAgent struct {
	*v1alpha1connect.UnimplementedAIServiceHandler
	numTries int
}

func (f *FakeAgent) GenerateCells(ctx context.Context, req *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error) {
	f.numTries += 1
	if f.numTries < 2 {
		return nil, connect.NewError(connect.CodeDeadlineExceeded, errors.New("Deadline exceeded"))
	}

	resp := &v1alpha1.GenerateCellsResponse{
		Cells: []*parserv1.Cell{
			{
				Kind: parserv1.CellKind_CELL_KIND_MARKUP,
			},
		},
	}
	return connect.NewResponse(resp), nil
}

func setupAndRunFakeServer(addr string, a *FakeAgent) (*http.Server, error) {
	log := zapr.NewLogger(zap.L())
	mux := http.NewServeMux()
	path, handler := v1alpha1connect.NewAIServiceHandler(a)
	mux.Handle(path, handler)

	srv := &http.Server{
		Addr: addr,
		// NB that we are using h2c here to support HTTP/2 without TLS
		// bidirectional streaming requires HTTP/2
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Graceful shutdown setup
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Info("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	go func() {
		log.Info("Server starting on ", "address", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(err, "Server ListenAndServe error")
		}

		<-idleConnsClosed
		log.Info("Server stopped")
	}()
	return srv, nil
}

func Test_RetryInterceptor(t *testing.T) {
	port, err := networking.GetFreePort()
	if err != nil {
		t.Fatalf("Error getting free port: %v", err)
	}

	addr := fmt.Sprintf("localhost:%d", port)

	fake := &FakeAgent{}
	srv, err := setupAndRunFakeServer(addr, fake)
	if err != nil {
		t.Fatalf("Error starting server: %v", err)
	}
	baseURL := fmt.Sprintf("http://%s", addr)
	client := v1alpha1connect.NewAIServiceClient(
		&http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
					// Use the standard Dial function to create a plain TCP connection
					return net.Dial(network, addr)
				},
			},
		},
		baseURL,
		connect.WithInterceptors(&RetryInterceptor{
			MaxRetries: 3,
			Backoff:    10 * time.Millisecond,
		}),
	)

	// First call should fail but the interceptor should retry
	resp, err := client.GenerateCells(context.Background(), connect.NewRequest(&v1alpha1.GenerateCellsRequest{}))
	if err != nil {
		t.Fatalf("Error calling GenerateCells: %v", err)
	}

	if len(resp.Msg.Cells) != 1 {
		t.Fatalf("Expected 1 cell but got: %v", len(resp.Msg.Cells))
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Logf("Error shutting down server: %v", err)
	}
}
