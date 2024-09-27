package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NewProtoToJsonCmd creates a command for converting a proto to json
func NewProtoToJsonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prototojson <file>",
		Short: "Dump the binary proto file to json",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				log := zapr.NewLogger(zap.L())
				if len(args) == 0 {
					log.Info("prototojson takes at least one argument which should be the path of the proto to dump.")
				}

				file := args[0]

				var message proto.Message
				var typeName string
				if strings.HasSuffix(file, ".evalexample.binpb") {
					message = &v1alpha1.EvalExample{}
					typeName = "EvalExample"
				}

				if strings.HasSuffix(file, ".example.binpb") {
					message = &v1alpha1.Example{}
					typeName = "Example"
				}

				if message == nil {
					return errors.Errorf("The type of proto could not be determined from the path suffix for file: %s", file)
				}
				data, err := os.ReadFile(file)
				if err != nil {
					return errors.Wrapf(err, "Error reading file %s", file)
				}

				if err := proto.Unmarshal(data, message); err != nil {
					return errors.Wrapf(err, "Error unmarshalling proto of type %s from file %s", typeName, file)
				}

				jsonP := protojson.Format(message)
				fmt.Fprintf(os.Stdout, "%s\n", jsonP)
				return nil
			}()
			if err != nil {
				fmt.Printf("Error running convert;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
