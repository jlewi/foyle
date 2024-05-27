package analyze

import (
	"encoding/json"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/config"
)

func TestGetBlockLog(t *testing.T) {
	t.Errorf("Test needs to be updated after we switched to using pebble to stor the block logs")
	// Create a new CrudHandler with a mock configuration
	handler, err := NewCrudHandler(config.Config{})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Add a block log to the handler's cache for testing
	handler.blockLogs = map[string]logspb.BlockLog{
		"test-id": {
			Id:       "test-id",
			ExitCode: 7,
		},
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
	if d := cmp.Diff(handler.blockLogs["test-id"], actual); d != "" {
		t.Errorf("Unexpected diff body (-want +got):\n%s", d)
	}
}
