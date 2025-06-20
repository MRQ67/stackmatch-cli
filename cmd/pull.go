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
	pullOutput string
)

var pullCmd = &cobra.Command{
	Use:   "pull <environment-id>",
	Short: "Download an environment from Supabase",
	Long: `Downloads an environment configuration from Supabase by ID.
The environment can be saved to a file or printed to stdout.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Get environment ID from args
		envID := args[0]

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Download from Supabase
		ctx := context.Background()
		env, err := supabaseClient.GetEnvironment(ctx, envID)
		if err != nil {
			log.Fatalf("Failed to get environment: %v", err)
		}

		// Convert to JSON
		envJSON, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal environment: %v", err)
		}

		// Output to file or stdout
		if pullOutput != "" {
			if err := os.WriteFile(pullOutput, envJSON, 0o644); err != nil {
				log.Fatalf("Failed to write to file: %v", err)
			}
			fmt.Printf("Environment saved to %s\n", pullOutput)
		} else {
			fmt.Println(string(envJSON))
		}
	},
}

func init() {
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", "", "Output file (default: stdout)")
}
