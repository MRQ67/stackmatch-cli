package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stackmatch",
	Short: "StackMatch: Clone environments, not just code.",
	Long:  `StackMatch is a CLI tool that helps developers scan, export, and import their development environment configurations.
It aims to eliminate "works on my machine" problems by providing a consistent way to manage development setups.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra prints its own errors for command-line parsing errors.
		// We exit here to ensure a non-zero exit code.
		os.Exit(1)
	}
}
