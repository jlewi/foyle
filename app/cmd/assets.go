package cmd

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/jlewi/foyle/app/pkg/assets"
	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
	"os"
)

// NewAssetsCmd returns a command to download the assets
func NewAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "assets",
	}

	cmd.AddCommand(NewAssetsDownloadCmd())
	return cmd
}

// NewAssetsDownloadCmd returns a command to download the assets
func NewAssetsDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "download",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := application.NewApp()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(); err != nil {
					return err
				}
				defer helpers.DeferIgnoreError(app.Shutdown)

				logVersion()

				m, err := assets.NewManager(*app.Config)
				if err != nil {
					return err
				}

				if err := m.Download(context.Background()); err != nil {
					return err
				}
				return nil
			}()

			if err != nil {
				fmt.Printf("Error running request;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
