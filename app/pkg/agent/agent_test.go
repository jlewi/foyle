package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/runme/ulid"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"

	"github.com/sashabaranov/go-openai"

	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/jlewi/foyle/app/pkg/learn"

	"github.com/jlewi/foyle/app/api"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
)

func Test_Generate(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	// This test is useful for two things
	// 1. Its a smoke test for all the logic.
	// 2. It can be used to do some basic prompt testing.
	//    We could start to think about this as level 1 evals in the context of Hamel's blog.
	type testCase struct {
		name string
		doc  *v1alpha1.Doc
		// maxResults is the number of results to retrieve for Rag completions.
		// <=0 means no RAG
		maxResults int
	}

	cases := []testCase{
		{
			name: "basic",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use gcloud to list all the cloud build jobs in project foyle",
					},
				},
			},
			maxResults: 0,
		},
		{
			name: "test-gcloud-iam",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "How do I debug why workload identity isn't working for a deployment in GKE?",
					},
				},
			},
			maxResults: 0,
		},
		{
			name: "prdiff",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use git to compute a diff and then generate a PR description",
					},
				},
			},
			maxResults: 2,
		},
	}

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Failed to initialize viper: %v", err)
	}
	cfg := config.GetConfig()

	// Setup logs
	c := zap.NewDevelopmentConfig()
	log, err := c.Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(log)

	client, err := oai.NewClient(*cfg)
	if err != nil {
		t.Fatalf("Error creating OpenAI client; %v", err)
	}

	vectorizer := oai.NewVectorizer(client)

	cfg.Agent.RAG = &api.RAGConfig{
		MaxResults: 3,
	}
	cfg.Agent.RAG.Enabled = true

	inMemoryDB, err := learn.NewInMemoryExampleDB(*cfg, vectorizer)
	if err != nil {
		t.Fatalf("Error creating in memory DB; %v", err)
	}

	cfg.Agent.ModelProvider = api.ModelProviderOpenAI
	cfg.Agent.Model = openai.GPT4oMini

	completer, err := oai.NewCompleter(*cfg, client)
	if err != nil {
		t.Fatalf("Error creating completer; %v", err)
	}
	agentWithRag, err := NewAgent(*cfg, completer, inMemoryDB)

	if err != nil {
		t.Fatalf("Error creating agent; %v", err)
	}

	cfgNoRag := cfg.DeepCopy()
	cfgNoRag.Agent.RAG.Enabled = false
	agentNoRag, err := NewAgent(cfgNoRag, completer, nil)

	if err != nil {
		t.Fatalf("Error creating agent; %v", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &v1alpha1.GenerateRequest{
				Doc: c.doc,
			}
			var agent *Agent
			if c.maxResults > 0 {
				agent = agentWithRag
				agent.config.Agent.RAG.MaxResults = c.maxResults
			} else {
				agent = agentNoRag
			}
			resp, err := agent.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("Error generating; %v", err)
			}
			t.Logf("Response: %+v", resp)
		})
	}
}

func Test_StreamingClient(t *testing.T) {
	// Test the streaming client
	// Useful when testing with a service version
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	// Setup logs
	c := zap.NewDevelopmentConfig()
	newLog, err := c.Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(newLog)
	log := zapr.NewLogger(newLog)
	// This is code to help us test streaming with the connect protocol
	addr := "http://127.0.0.1:8877/api"
	//addr := "http://127.0.0.1:9977/api"

	log.Info("Server started")
	if err := runClient(addr); err != nil {
		t.Fatalf("Error running client for addres %v; %v", addr, err)
	}
}

func Test_Streaming(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	// Setup logs
	c := zap.NewDevelopmentConfig()
	newLog, err := c.Build()
	if err != nil {
		t.Fatalf("Error creating logger; %v", err)
	}
	zap.ReplaceGlobals(newLog)
	log := zapr.NewLogger(newLog)
	// This is code to help us test streaming with the connect protocol
	a := &Agent{}

	addr := "http://localhost:8088"
	go func() {
		if err := setupAndRunServer(addr, a); err != nil {
			log.Error(err, "Error running server")
		}
	}()

	// N.B. There's probably a race condition here because the client might start before the server is fully up.
	// Or maybe that's implicitly handled because the connection won't succeed until the server is up?
	if err := waitForServer(addr); err != nil {
		t.Fatalf("Error waiting for server; %v", err)
	}
	log.Info("Server started")
	if err := runClient(addr); err != nil {
		t.Fatalf("Error running client for addres %v; %v", addr, err)
	}
}

func waitForServer(addr string) error {
	endTime := time.Now().Add(30 * time.Second)
	wait := 2 * time.Second
	for time.Now().Before(endTime) {
		resp, err := http.Get(fmt.Sprintf("http://%s/healthz", addr))
		if err != nil {
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}
	return errors.Errorf("Server didn't start in time")
}

func runClient(baseURL string) error {
	log := zapr.NewLogger(zap.L())
	u, err := url.Parse(baseURL)
	if err != nil {
		log.Error(err, "Failed to parse URL")
		panic(err)
	}

	var client v1alpha1connect.AIServiceClient

	if u.Scheme == "https" {
		// Configure the TLS settings
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Set to true only for testing; otherwise validate the server's certificate
		}

		client = v1alpha1connect.NewAIServiceClient(
			&http.Client{
				Transport: &http2.Transport{
					TLSClientConfig: tlsConfig,
					DialTLSContext: func(ctx context.Context, network, addr string, config *tls.Config) (net.Conn, error) {
						// Create a secure connection with TLS
						return tls.Dial(network, addr, config)
					},
				},
			},
			baseURL,
		)
	} else {
		client = v1alpha1connect.NewAIServiceClient(
			&http.Client{
				Transport: &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						// Use the standard Dial function to create a plain TCP connection
						return net.Dial(network, u.Host)
					},
				},
			},
			baseURL,
		)
	}

	// Make sure we can send a status request
	statusReq := &v1alpha1.StatusRequest{}
	statusResp, err := client.Status(context.Background(), connect.NewRequest(statusReq))
	if err != nil {
		log.Error(err, "Failed to send status request")
		return errors.Wrapf(err, "Failed to send status request")
	}

	log.Info("Status response", "response", statusResp)

	ctx := context.Background()
	stream := client.StreamGenerate(ctx)

	// Send requests
	requests := []string{"Hello", "How are you?", "Goodbye"}

	contextId := ulid.GenerateID()
	for i, prompt := range requests {

		var req *v1alpha1.StreamGenerateRequest
		if i == 0 {
			req = &v1alpha1.StreamGenerateRequest{
				ContextId: contextId,
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
						NotebookUri: "/path/to/notebook",
					},
				},
			}
		} else {
			req = &v1alpha1.StreamGenerateRequest{
				ContextId: contextId,
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

	// Signal that we're done sending
	if err := stream.CloseRequest(); err != nil {
		log.Error(err, "Failed to close request stream: %v")
	}

	// Receive responses
	for {
		response, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream closed")
			break
		}
		if err != nil {
			log.Error(err, "Failed to receive response")
			return errors.Wrapf(err, "Failed to receive response")
		}
		log.Info("Received response", "response", response)
	}
	return nil
}

func setupAndRunServer(addr string, a *Agent) error {
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

	log.Info("Server starting on ", "address", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	log.Info("Server stopped")
	return nil
}

func Test_ShouldTrigger(t *testing.T) {
	type testCase struct {
		name          string
		doc           *v1alpha1.Doc
		selectedIndex int32
		expected      bool
	}

	cases := []testCase{
		{
			name: "markupcell",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use gcloud to list all the cloud build jobs in project foyle",
						Kind:     v1alpha1.BlockKind_MARKUP,
					},
				},
			},
			selectedIndex: 0,
			expected:      true,
		},
		{
			name: "codecell",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "Use gcloud to list all the cloud build jobs in project foyle",
						Kind:     v1alpha1.BlockKind_CODE,
					},
				},
			},
			selectedIndex: 0,
			expected:      true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := shouldTrigger(c.doc, c.selectedIndex)
			if actual != c.expected {
				t.Fatalf("Expected %v but got %v", c.expected, actual)
			}
		})
	}
}

func Test_PostProcessBlocks(t *testing.T) {
	type testCase struct {
		name     string
		blocks   []*v1alpha1.Block
		expected []*v1alpha1.Block
	}

	cases := []testCase{
		{
			name: "output-tag-in-codeblocks",
			blocks: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Contents: "</output>",
				},
			},
			expected: []*v1alpha1.Block{},
		},
		{
			name: "whitespace-only-block",
			blocks: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Contents: "   ",
				},
			},
			expected: []*v1alpha1.Block{},
		},
		{
			name: "merge-markup-blocks",
			blocks: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "first block",
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "second block",
				},
			},
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "first block\nsecond block",
				},
			},
		},
		{
			name: "stop-at-code-block",
			blocks: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "first block",
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Contents: "echo hello",
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "last block",
				},
			},
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "first block",
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Contents: "echo hello",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := postProcessBlocks(c.blocks)
			if err != nil {
				t.Fatalf("Error post processing blocks; %v", err)
			}
			if d := cmp.Diff(c.expected, actual, cmpopts.IgnoreUnexported(v1alpha1.Block{})); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}

func Test_dropResponse(t *testing.T) {
	type testCase struct {
		name     string
		response *v1alpha1.StreamGenerateResponse
		expected bool
	}

	cases := []testCase{
		{
			name: "basic",
			response: &v1alpha1.StreamGenerateResponse{
				Cells: []*parserv1.Cell{
					{
						Kind:  parserv1.CellKind_CELL_KIND_CODE,
						Value: "print('Hello, World!')",
					},
				},
			},
			expected: false,
		},
		{
			name: "empty",
			response: &v1alpha1.StreamGenerateResponse{
				Cells: []*parserv1.Cell{},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := dropResponse(c.response)
			if result != c.expected {
				t.Errorf("Expected %v; got %v", c.expected, result)
			}
		})
	}
}
