package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jlewi/foyle/app/pkg/application"

	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
)

// NewServeCmd returns a command to run the server
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(true); err != nil {
					return err
				}
				if err := app.SetupOTEL(); err != nil {
					return err
				}
				if err := app.OpenDBs(); err != nil {
					return err
				}

				logDirs := make([]string, 0, 2)
				logDirs = append(logDirs, app.Config.GetRawLogDir())

				if app.Config.Learner != nil {
					logDirs = append(logDirs, app.Config.Learner.LogDirs...)
				}

				analyzer, err := app.SetupAnalyzer()
				if err != nil {
					return err
				}

				analyzer.Run(context.Background(), logDirs)
				s, err := app.SetupServer()
				if err != nil {
					return err
				}
				defer helpers.DeferIgnoreError(app.Shutdown)

				// Analyzer needs to be shutdown before the app because the app will close the database
				defer helpers.DeferIgnoreError(func() error {
					return analyzer.Shutdown(context.Background())
				})

				logVersion()
				return s.Run()

			}()

			if err != nil {
				fmt.Printf("Error running request;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
