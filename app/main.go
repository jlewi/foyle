package main

import (
	"fmt"
	"os"

	"github.com/jlewi/foyle/app/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}
