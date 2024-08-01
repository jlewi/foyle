package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jlewi/hydros/pkg/util"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

// Test_StreamGenerate is used during development to test that we can connect to the streaming service
// when its hosted in the server.
// It needs to live in the application directory because we use the application resource to setup the server.
func Test_StreamGenerate(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	app, err := initApp()
	if err != nil {
		t.Fatalf("Error initializing app; %v", err)
	}
	defer util.DeferIgnoreError(app.Shutdown)
	go func() {
		// N.B. We are running it in a subprocess so if you pause in a debugger the server won't be able to
		// respond to the request. Not sure if that's a good thing or bad thing.
		if err := app.Serve(); err != nil {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "Error running server")
		}
	}()

	address := fmt.Sprintf("http://localhost:%d", app.Config.Server.HttpPort)
	// N.B. There's probably a race condition here because the client might start before the server is fully up.
	// Or maybe that's implicitly handled because the connection won't succeed until the server is up?
	if err := waitForServer(address); err != nil {
		t.Fatalf("Error waiting for server; %v", err)
	}

	//select {}
	// Now have the client stream some requests
	if err := runClient(app.Config.APIBaseURL()); err != nil {
		t.Fatalf("Error running client; %v", err)
	}
	//select {}
}

// Test_RunServer runs the server for error.
// Use this if you want to run the server as separate process
func Test_RunServer(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	app, err := initApp()
	if err != nil {
		t.Fatalf("Error initializing app; %v", err)
	}
	if err := app.OpenDBs(); err != nil {
		t.Fatalf("Error opening DBs; %v", err)
	}
	defer util.DeferIgnoreError(app.Shutdown)
	if err := app.Serve(); err != nil {
		t.Fatalf("Error running server; %v", err)
	}
}

func Test_StreamGenerateClient(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	app, err := initApp()
	if err != nil {
		t.Fatalf("Error initializing app; %v", err)
	}
	defer util.DeferIgnoreError(app.Shutdown)

	// This test only runs the client. It assumes the server is started in a different process.
	// This is useful if you want to step through the client code and don't want to pause the server while
	// stepping. You need to start the server separately.

	if err := runClient("http://localhost:9050/api"); err != nil {
		t.Fatalf("Error running client; %v", err)
	}
}

func initApp() (*App, error) {
	app := NewApp()

	if err := app.LoadConfig(nil); err != nil {
		return nil, err
	}

	// Do some surgery on the config to make it suitable for testing
	// We want to not interfere with a "production" version of the server running on your machine
	app.Config.Server.HttpPort = 9050
	app.Config.Server.GRPCPort = 9070

	app.Config.Logging.LogDir = "/tmp/foyle/logs"
	app.Config.Logging.Level = "debug"
	app.Config.Learner.ExampleDirs = []string{"/tmp/foyle/examples"}

	if err := app.SetupLogging(true); err != nil {
		return nil, err
	}

	return app, nil
}

func waitForServer(addr string) error {
	endTime := time.Now().Add(30 * time.Second)
	wait := 2 * time.Second
	log := zapr.NewLogger(zap.L())
	for time.Now().Before(endTime) {
		resp, err := http.Get(fmt.Sprintf("http://%s/healthz", addr))
		if err != nil {
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		}
		payload, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Error reading response body")
		}
		log.Info("Response", "status", resp.StatusCode, "body", string(payload))
		time.Sleep(wait)
	}
	return errors.Errorf("Server didn't start in time")
}

func runClient(baseURL string) error {
	log := zapr.NewLogger(zap.L())
	client := v1alpha1connect.NewAIServiceClient(
		// N.B. We need to use HTTP2 if we want to support bidirectional streaming
		//http.DefaultClient,
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
	)

	ctx := context.Background()

	stream := client.StreamGenerate(ctx)

	// Send requests
	requests := []string{"Hello", "How are you?", "Goodbye"}
	for i, prompt := range requests {

		var req *v1alpha1.StreamGenerateRequest

		if i == 0 {
			req = &v1alpha1.StreamGenerateRequest{
				Request: &v1alpha1.StreamGenerateRequest_FullContext{
					FullContext: &v1alpha1.FullContext{
						Notebook: &parserv1.Notebook{
							Cells: []*parserv1.Cell{
								{
									Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
									Value: prompt,
								},
							},
						},
						NotebookUri: "file://foo.md",
						Selected:    0,
					},
				},
			}
		} else {
			req = &v1alpha1.StreamGenerateRequest{
				Request: &v1alpha1.StreamGenerateRequest_Update{
					Update: &v1alpha1.UpdateContext{
						Cell: &parserv1.Cell{
							Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
							Value: prompt,
						},
					},
				},
			}
		}

		err := stream.Send(req)

		if err != nil {
			log.Error(err, "Failed to send request")
		}
		fmt.Printf("Sent request: %s\n", prompt)
	}

	// Receive responses
	// We should get at least 2 responses.
	// The first response should be fore the initial request
	// And then we should do a second generation for all the updates
	for i := 0; i < 2; i++ {
		response, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream closed")
			break
		}
		if err != nil {
			log.Error(err, "Failed to receive response")
			return err
		}
		log.Info("Received response", "response", response)
	}

	// Signal that we're done sending.
	// This will cause the server to abort with EOF
	if err := stream.CloseRequest(); err != nil {
		log.Error(err, "Failed to close request stream: %v")
	}

	return nil
}
