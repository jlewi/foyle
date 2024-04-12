package cmd

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/hydros/pkg/app"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Reconcile releases",
		Run: func(cmd *cobra.Command, args []string) {

			err := func() error {
				return Reconcile()
			}()
			if err != nil {
				fmt.Printf("generate failed; error %+v\n", err)
			}
		},
	}
	return cmd
}

func Reconcile() error {
	hApp := app.NewApp()

	if err := hApp.LoadConfig(nil); err != nil {
		return err
	}

	if err := hApp.SetupLogging(); err != nil {
		return err
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Reconciling releases")
	if err := hApp.ApplyPaths(context.Background(), []string{"/Users/jlewi/git_foyle/releasing.yaml"}, 0, false); err != nil {
		return err
	}

	log.Info("Reconciling image tags")
	if err := hApp.ApplyPaths(context.Background(), []string{"/Users/jlewi/git_foyle/frontend/foyle/image_replicator.yaml"}, 0, false); err != nil {
		return err
	}
	return nil
}
