package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewLogsCmd returns a command to manage logs
func NewLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "logs",
	}

	cmd.AddCommand(NewLogsProcessCmd())
	return cmd
}

// NewLogsProcessCmd returns a command to process the assets
func NewLogsProcessCmd() *cobra.Command {
	var logsDir string
	var outDir string
	cmd := &cobra.Command{
		Use: "process",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(false); err != nil {
					return err
				}
				defer helpers.DeferIgnoreError(app.Shutdown)

				logVersion()

				if logsDir == "" {
					logsDir = app.Config.GetRawLogDir()
				}

				a, err := analyze.NewAnalyzer()
				if err != nil {
					return err
				}

				log := zapr.NewLogger(zap.L())
				log.Info("Processing logs", "logs", logsDir)

				if err := a.Analyze(context.Background(), logsDir, app.Config.GetTracesDBDir(), app.Config.GetBlocksDBDir()); err != nil {
					return err
				}
				log.Info("Processed logs", "logs", logsDir, "traces", app.Config.GetTracesDBDir(), "blocks", app.Config.GetBlocksDBDir())
				return nil
			}()

			if err != nil {
				fmt.Printf("Error processing logs;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&logsDir, "logs", "", "", "(Optional) Directory containing logs to process")
	cmd.Flags().StringVarP(&outDir, "out", "", "", "(Optional) Directory to write the output to")
	return cmd
}
