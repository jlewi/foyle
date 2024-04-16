package logs

import (
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"os"
	"testing"
)

func Test_ZapPB(t *testing.T) {
	c := zap.NewDevelopmentConfig()
	zap.NewProductionConfig()
	oFile, err := os.CreateTemp("", "logs.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	outputName := oFile.Name()
	if err := oFile.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	c.OutputPaths = []string{"stdout", outputName}
	log, err := c.Build()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

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
	log.Info("Test_ZapPB", ZapPB("req", request))
}
