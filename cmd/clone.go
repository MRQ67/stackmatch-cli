package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <username>/<env-name>",
	Short: "Clone another user's environment",
	Long: `Clones an environment from another user and applies it locally.
Format should be 'username/env-name'.`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse username and env name
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			log.Fatal("Invalid format. Use: username/env-name")
		}
		username := parts[0]
		envName := parts[1]

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Find environment by username and name
		ctx := context.Background()
		sourceEnv, err := supabaseClient.FindEnvironmentByUserAndName(ctx, username, envName)
		if err != nil {
			log.Fatalf("Failed to find environment: %v", err)
		}

		// Get the current user from the session
		user := auth.GetCurrentUser()
		if user == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Add user to context
		ctx = context.WithValue(ctx, "user", user)

		// Convert the environment data to JSON for display
		envJSON, err := json.MarshalIndent(sourceEnv, "", "  ")
		if err != nil {
			log.Fatalf("Failed to format environment data: %v", err)
		}

		// Print the environment data
		fmt.Printf("Environment '%s' from user '%s':\n%s\n", envName, username, string(envJSON))
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
