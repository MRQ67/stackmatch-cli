package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var (
	downloadOutput string
	downloadID    string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download an environment from Supabase",
	Long: `Downloads an environment configuration from Supabase by ID.
The environment can be saved to a file or printed to stdout.

Requires authentication.`,
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate config
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Initialize Supabase client with config values
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Download from Supabase
		ctx := context.Background()
		env, err := supabaseClient.GetEnvironment(ctx, downloadID)
		if err != nil {
			log.Fatalf("Failed to download environment: %v", err)
		}

		// Convert to JSON
		envJSON, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal environment: %v", err)
		}

		// Output the result
		if downloadOutput != "" {
			if err := os.WriteFile(downloadOutput, envJSON, 0644); err != nil {
				log.Fatalf("Failed to write to file: %v", err)
			}
			fmt.Printf("Environment saved to %s\n", downloadOutput)
		} else {
			fmt.Println(string(envJSON))
		}
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "Output file (default: print to stdout)")
	downloadCmd.Flags().StringVarP(&downloadID, "id", "i", "", "Environment ID to download (required)")
	downloadCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(downloadCmd)
}
