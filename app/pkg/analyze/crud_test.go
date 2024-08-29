package analyze

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"github.com/cockroachdb/pebble"
	"github.com/gin-gonic/gin"
	"github.com/jlewi/foyle/app/pkg/config"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
)

func populateDB(db *pebble.DB) error {
	block := &logspb.BlockLog{
		Id:         "test-id",
		GenTraceId: "sometrace",
	}

	return dbutil.SetProto(db, block.GetId(), block)
}

func populateTraceDB(db *pebble.DB) error {
	trace := &logspb.Trace{
		Id: "test-trace",
		Data: &logspb.Trace_Execute{
			Execute: &logspb.ExecuteTrace{
				Request: &v1alpha1.ExecuteRequest{
					Block: &v1alpha1.Block{
						Id:       "test-block",
						Contents: "echo hello",
					},
				},
			},
		},
	}
	return dbutil.SetProto(db, trace.GetId(), trace)
}

func createCrudHandler() (*CrudHandler, error) {
	logsDir, err := os.MkdirTemp("", "crudTest")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create temp dir")
	}
	cfg := config.Config{
		Logging: config.Logging{
			LogDir: logsDir,
		},
	}

	db, err := pebble.Open(cfg.GetBlocksDBDir(), &pebble.Options{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open DB")
	}

	if err := populateDB(db); err != nil {
		return nil, errors.Wrapf(err, "Failed to populate DB")
	}

	tracesDB, err := pebble.Open(cfg.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open DB")
	}

	if err := populateTraceDB(tracesDB); err != nil {
		return nil, errors.Wrapf(err, "Failed to populate DB")
	}

	// Create a new CrudHandler with a mock configuration
	return NewCrudHandler(cfg, db, tracesDB, nil)
}

func tearDown(handler *CrudHandler) {
	handler.blocksDB.Close()
	handler.tracesDB.Close()
}

func TestGetBlockLog(t *testing.T) {
	handler, err := createCrudHandler()
	defer tearDown(handler)

	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	// Create a new Gin engine
	router := gin.Default()

	// Register the GetBlockLog handler
	router.GET("/block/:id", handler.GetBlockLog)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "/block/test-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Record the response
	w := httptest.NewRecorder()

	// Serve the HTTP request
	router.ServeHTTP(w, req)

	// Check the HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	// Check the response body
	actualBody := w.Body.String()
	actual := &logspb.BlockLog{}
	if err := json.Unmarshal([]byte(actualBody), &actual); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
}

func TestGetTrace(t *testing.T) {
	handler, err := createCrudHandler()
	defer tearDown(handler)

	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	type testCase struct {
		name         string
		id           string
		expectedCode *connect.Code
	}

	notFound := connect.CodeNotFound
	cases := []testCase{
		{
			name: "basic",
			id:   "test-trace",
		},
		{
			name:         "notFound",
			id:           "someunknownid",
			expectedCode: &notFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &connect.Request[logspb.GetTraceRequest]{
				Msg: &logspb.GetTraceRequest{
					Id: c.id,
				},
			}
			resp, err := handler.GetTrace(context.Background(), req)

			if c.expectedCode == nil {
				if resp == nil {
					t.Fatalf("Expected response but got nil")
				}
			} else {
				if connect.CodeOf(err) != *c.expectedCode {
					t.Fatalf("Expected code %v, got %v", *c.expectedCode, connect.CodeOf(err))
				}
			}
		})
	}
}
