package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewApplyCmd create an apply command
func NewApplyCmd() *cobra.Command {
	// TODO(jeremy): We should update apply to support the image resource.
	applyCmd := &cobra.Command{
		Use:   "apply <resource.yaml> <resourceDir> <resource.yaml> ...",
		Short: "Apply the specified resource.",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				log := zapr.NewLogger(zap.L())
				if len(args) == 0 {
					log.Info("apply takes at least one argument which should be the file or directory YAML to apply.")
					return errors.New("apply takes at least one argument which should be the file or directory YAML to apply.")
				}
				logVersion()

				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				// We need to log to a file because in the case of Eval we want to capture traces.
				if err := app.SetupLogging(true); err != nil {
					return err
				}
				// We need to setup OTEL because we rely on the trace provider.
				if err := app.SetupOTEL(); err != nil {
					return err
				}

				// DBs can only be opened in a single process.
				if err := app.OpenDBs(); err != nil {
					return err
				}

				if err := app.SetupRegistry(); err != nil {
					return err
				}

				return app.ApplyPaths(context.Background(), args)
			}()
			if err != nil {
				fmt.Printf("Error running apply;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return applyCmd
}
