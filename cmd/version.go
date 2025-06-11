package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const cliVersion = "0.1.2" // Define the current version

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of StackMatch",
	Long:  `All software has versions. This is StackMatch's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("StackMatch CLI version %s\n", cliVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd) // Add versionCmd to rootCmd
}
