package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [environment_name]",
	Short: "Delete an environment from Supabase",
	Long:  `Deletes an environment that you have pushed to Supabase.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the environment name
		envName := args[0]

		// Get the current user
		user := auth.GetCurrentUser()
		if user == nil {
			log.Fatal("You must be logged in to delete an environment.")
		}

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey, user.AccessToken)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Delete the environment
		if err := supabaseClient.DeleteEnvironment(context.Background(), envName, user.ID); err != nil {
			log.Fatalf("Failed to delete environment: %v", err)
		}

		fmt.Printf("Environment '%s' deleted successfully.\n", envName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}