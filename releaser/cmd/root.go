package cmd

import (
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var level string
	var jsonLog bool
	rootCmd := &cobra.Command{
		Short: "foyle releaser",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			_, err := logging.InitLogger(level, !jsonLog)
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&level, "level", "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", false, "Enable json logging.")

	rootCmd.AddCommand(NewRunCmd())
	return rootCmd
}
