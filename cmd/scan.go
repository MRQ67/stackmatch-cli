package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/MRQ67/stackmatch-cli/internal/utils"
	"github.com/MRQ67/stackmatch-cli/pkg/scanner"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the current system and detect development tools and configurations",
	Long: `The scan command inspects your local development environment to identify:
    - Operating System & Architecture
    - Installed development tools (e.g., Node.js, Python, Git, Docker)
    - Package managers (e.g., Homebrew, apt, Chocolatey)
    - Code editors (e.g., VS Code)
    - Versions of programming languages
    - Common configuration files (e.g., .gitconfig, .zshrc)

The output is a JSON representation of your environment that can be exported or used for comparison.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scanning environment...")

		envData := types.EnvironmentData{
			StackmatchVersion: cliVersion, // from version.go
			ScanDate:          time.Now().UTC(),
			Tools:             make(map[string]string),
			PackageManagers:   make(map[string]string),
			CodeEditors:       make(map[string]string),
			ConfiguredLanguages: make(map[string]string),
			ConfigFiles:       []string{},
		}

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

		fmt.Println("\nScan complete. Generating report...")

		jsonData, err := json.MarshalIndent(envData, "", "  ")
		if err != nil {
			utils.ExitWithError(fmt.Errorf("failed to generate JSON output: %w", err))
		}

		fmt.Println(string(jsonData))
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
