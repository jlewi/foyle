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

				learner, err := app.SetupLearner()

				if err != nil {
					return err
				}

				if err := analyzer.Run(context.Background(), logDirs, learner.Enqueue); err != nil {
					return err
				}

				if err := learner.Start(context.Background()); err != nil {
					return err
				}

				s, err := app.SetupServer()
				if err != nil {
					return err
				}
				defer helpers.DeferIgnoreError(app.Shutdown)

				// Analyzer should be shutdown before the learner because analyzer tries to enqueue learner items
				defer helpers.DeferIgnoreError(func() error {
					return learner.Shutdown(context.Background())
				})

				// Analyzer needs to be shutdown before the app because the app will close the database
				// TODO(jeremy): Should we move this into app.shutdown and make app.Shutdown responsible for
				// shutting down analyzer in the proper order if there is one?
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
