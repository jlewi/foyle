package logs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// This is a test to demonstrate how to log a protobuf message using zap and the generated protobuf code.
// https://stackoverflow.com/questions/68411821/correctly-log-protobuf-messages-as-unescaped-json-with-zap-logger
func Test_ZapPB(t *testing.T) {
	c := zap.NewProductionConfig()

	oFile, err := os.CreateTemp("", "logs.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	outputName := oFile.Name()
	if err := oFile.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	t.Logf("Output writing to: %s", outputName)
	c.OutputPaths = []string{"stdout", outputName}
	logger, err := c.Build()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// We need to AllowZapFields to ensure the protobuf message is logged as a JSON object.
	// N.B. This breaks the implementation agnosticism of logr.
	log := zapr.NewLoggerWithOptions(logger, zapr.AllowZapFields(true))
	request := &v1alpha1.ExecuteResponse{
		Outputs: []*v1alpha1.BlockOutput{
			{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: "hello",
					},
				},
			},
		},
	}
	log.Info("Test_ZapPB", zap.Object("req", request))
	//if err := log.Sync(); err != nil {
	//	t.Logf("Ignoring sync logger: %v", err)
	//}

	// Make sure we can decode it
	file, err := os.Open(outputName)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}

	type logLine struct {
		RawProto *json.RawMessage `json:"req"`
	}

	l := &logLine{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(l); err != nil {
		t.Fatalf("Failed to decode log entry: %v", err)
	}

	actual := &v1alpha1.ExecuteResponse{}
	if err := protojson.Unmarshal(*l.RawProto, actual); err != nil {
		t.Fatalf("Failed to decode block log: %v", err)
	}

	if d := cmp.Diff(request, actual, testutil.BlockComparer, cmpopts.IgnoreUnexported(v1alpha1.ExecuteResponse{})); d != "" {
		t.Errorf("Unexpected diff: %v", d)
	}
}
