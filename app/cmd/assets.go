package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlewi/foyle/app/pkg/application"
	"github.com/jlewi/foyle/app/pkg/assets"
	"github.com/jlewi/monogo/helpers"
	"github.com/spf13/cobra"
)

const (
	defaultTag = "latest"
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
	var tag string
	cmd := &cobra.Command{
		Use: "download",
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

				m, err := assets.NewManager(*app.Config)
				if err != nil {
					return err
				}

				if tag == "" {
					if commit == commitNotSet {
						// Since the commit isn't set we are using a development build so we use the latest tag
						tag = defaultTag
					} else {
						tag = commit
					}
				}
				log := zapr.NewLogger(zap.L())
				log.Info("Downloading assets", "tag", tag)
				if err := m.Download(context.Background(), tag); err != nil {
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

	cmd.Flags().StringVarP(&tag, "tag", "", "", "The tag for the assets to download. If empty downloads the assets matching the commit of the binary")
	return cmd
}
