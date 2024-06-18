package cmd

import (
	"fmt"
	"os"

	"github.com/jlewi/monogo/helpers"

	"github.com/jlewi/foyle/app/pkg/application"

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
				logVersion()
				defer helpers.DeferIgnoreError(app.Shutdown)
				return app.Serve()
			}()

			if err != nil {
				fmt.Printf("Error running request;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
