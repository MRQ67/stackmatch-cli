package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/scanner"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/spf13/cobra"
)

// scanEnvironment scans the current development environment
func scanEnvironment() *types.EnvironmentData {
	envData := &types.EnvironmentData{
		Tools:              make(map[string]string),
		PackageManagers:    make(map[string]string),
		CodeEditors:        make(map[string]string),
		ConfiguredLanguages: make(map[string]string),
		ConfigFiles:        []string{},
		System:             types.SystemInfo{},
	}

	// Set scan timestamp
	envData.ScanDate = time.Now()

	// Detect system information
	scanner.DetectSystemInfo(&envData.System)

	// Detect package managers
	scanner.DetectPackageManagers(envData)

	// Detect programming languages
	scanner.DetectProgrammingLanguages(envData)

	// Detect development tools
	scanner.DetectTools(envData)
	// Detect code editors and IDEs
	scanner.DetectEditors(envData)
	// Scan for configuration files
	scanner.DetectConfigFiles(envData)

	return envData
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Scan & push environment data to Supabase",
	Long: `Scans the current development environment and uploads the configuration to Supabase.
This requires authentication and Supabase URL/API key to be set.`,
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Use the global authenticated Supabase client
		if supabaseClient == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Scan the environment
		envData := scanEnvironment()

		// Get the current user from the session
		user := auth.GetCurrentUser()
		if user == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Upload to Supabase
		ctx := context.Background()
		id, err := supabaseClient.SaveEnvironment(ctx, envData)
		if err != nil {
			log.Fatalf("Failed to save environment: %v", err)
		}

		fmt.Printf("Successfully saved environment with ID: %s\n", id)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
