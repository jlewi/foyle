package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"

	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/jlewi/foyle/app/pkg/learn"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
)

// NewLearnCmd returns a command to learn from past mistakes.
func NewLearnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "learn",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(false); err != nil {
					return err
				}
				if err := app.OpenDBs(); err != nil {
					return err
				}
				defer helpers.DeferIgnoreError(app.Shutdown)

				logVersion()

				client, err := oai.NewClient(*app.Config)
				if err != nil {
					return err
				}

				l, err := learn.NewLearner(*app.Config, client, app.LockingBlocksDB)
				if err != nil {
					return err
				}

				if err := l.Reconcile(context.Background(), "someid"); err != nil {
					return err
				}
				return errors.New("Not implemented; code needs to be updated for latest Reconcile implementation")
			}()

			if err != nil {
				fmt.Printf("Error learning;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
