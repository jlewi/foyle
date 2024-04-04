package cmd

import (
	"fmt"
	"os"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/spf13/cobra"
)

const (
	appName = "foyle"
)

func NewRootCmd() *cobra.Command {
	var cfgFile string
	var level string
	var jsonLog bool
	rootCmd := &cobra.Command{
		Short: appName,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, config.ConfigFlagName, "", fmt.Sprintf("config file (default is $HOME/.%s/config.yaml)", appName))
	rootCmd.PersistentFlags().StringVarP(&level, config.LevelFlagName, "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", false, "Enable json logging.")

	rootCmd.AddCommand(NewAssetsCmd())
	rootCmd.AddCommand(NewVersionCmd(appName, os.Stdout))
	rootCmd.AddCommand(NewServeCmd())
	return rootCmd
}
