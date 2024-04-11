package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
	// Use hydros to build the image

	return nil
}
