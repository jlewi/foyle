package docs

import (
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

// GetExitCode returns the exit code from a block output if the block represents the exit code
// The function returns the exit code and a boolean indicating if the block represents the exit code.
// N.B. Keep this in sync with resultsToProto in executor.go
func GetExitCode(b v1alpha1.BlockOutput) (int, bool) {
	log := zapr.NewLogger(zap.L())
	for _, oi := range b.Items {
		if oi.GetTextData() == "" {
			continue
		}

		if strings.HasPrefix(oi.GetTextData(), "exitCode:") {
			parts := strings.Split(oi.GetTextData(), ":")
			if len(parts) != 2 {
				continue
			}
			code, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				log.Error(err, "Failed to parse exit code", "output", oi.GetTextData())
				continue
			}
			return code, true
		}
	}
	return -1, false
}
