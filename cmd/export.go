package cmd

import (
	"fmt"
	"time"

	"github.com/MRQ67/stackmatch-cli/internal/utils"
	"github.com/MRQ67/stackmatch-cli/pkg/exporter"
	"github.com/MRQ67/stackmatch-cli/pkg/scanner"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [filename]",
	Short: "Scan the environment and export it to a JSON file",
	Long:  `Scans the local development environment and saves the complete configuration to a specified JSON file.
This file can be used for sharing, analysis, or later with the 'import' command.`,
	Args:  cobra.ExactArgs(1), // Ensures exactly one argument (the filename) is provided
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := args[0]
		fmt.Printf("Scanning environment to export to %s...\n", outputFile)

		envData := types.EnvironmentData{
			StackmatchVersion: cliVersion,
			ScanDate:          time.Now().UTC(),
			Tools:             make(map[string]string),
			PackageManagers:   make(map[string]string),
			CodeEditors:       make(map[string]string),
			ConfiguredLanguages: make(map[string]string),
			ConfigFiles:       []string{},
		}

		// Run all our detection logic
		fmt.Println("• Detecting system info...")
		scanner.DetectSystemInfo(&envData.System)
		fmt.Println("• Detecting programming languages...")
		scanner.DetectProgrammingLanguages(&envData)
		fmt.Println("• Detecting development tools...")
		scanner.DetectTools(&envData)
		fmt.Println("• Detecting package managers...")
		scanner.DetectPackageManagers(&envData)
		fmt.Println("• Detecting code editors...")
		scanner.DetectEditors(&envData)
		fmt.Println("• Detecting config files...")
		scanner.DetectConfigFiles(&envData)

		fmt.Println("\nScan complete.")

		// Export the data
		err := exporter.WriteJSON(envData, outputFile)
		if err != nil {
			utils.ExitWithError(fmt.Errorf("could not export data: %w", err))
		}

		fmt.Printf("Environment successfully exported to %s\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
