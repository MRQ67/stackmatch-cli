package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/MRQ67/stackmatch-cli/pkg/scanner"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Scan the current environment and upload it to Supabase",
	Long: `Scans the local development environment and uploads the configuration to Supabase.
This requires authentication and Supabase URL/API key to be set.`,
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize Supabase client with config values
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Scan the environment
		envData := types.EnvironmentData{
			Tools:             make(map[string]string),
			PackageManagers:   make(map[string]string),
			CodeEditors:       make(map[string]string),
			ConfiguredLanguages: make(map[string]string),
			ConfigFiles:       []string{},
		}

		// Initialize system info
		scanner.DetectSystemInfo(&envData.System)

		// Run all detection functions
		scanner.DetectProgrammingLanguages(&envData)
		scanner.DetectTools(&envData)
		scanner.DetectPackageManagers(&envData)
		scanner.DetectEditors(&envData)
		scanner.DetectConfigFiles(&envData)

		// Upload to Supabase
		ctx := context.Background()
		envID, err := supabaseClient.SaveEnvironment(ctx, &envData)
		if err != nil {
			log.Fatalf("Failed to upload environment to Supabase: %v", err)
		}

		fmt.Printf("Successfully uploaded environment to Supabase with ID: %s\n", envID)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
