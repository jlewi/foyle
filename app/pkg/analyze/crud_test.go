package analyze

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cockroachdb/pebble"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/gin-gonic/gin"
	"github.com/jlewi/foyle/app/pkg/config"
)

func populateDB(blocksDB string) error {
	db, err := pebble.Open(blocksDB, &pebble.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	bLog := &logspb.BlockLog{
		Id:         "test-id",
		GenTraceId: "sometrace",
	}

	data, err := proto.Marshal(bLog)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal block log")
	}
	return db.Set([]byte("test-id"), data, pebble.Sync)
}

func TestGetBlockLog(t *testing.T) {
	logsDir, err := os.MkdirTemp("", "crudTest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cfg := config.Config{
		Logging: config.Logging{
			LogDir: logsDir,
		},
	}
	if err := populateDB(cfg.GetBlocksDBDir()); err != nil {
		t.Fatalf("Failed to populate DB: %v", err)
	}

	// Create a new CrudHandler with a mock configuration
	handler, err := NewCrudHandler(cfg)
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
