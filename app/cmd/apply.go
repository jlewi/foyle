package cmd

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
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
				if err := app.SetupLogging(false); err != nil {
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
