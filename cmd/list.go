package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your environments stored in Supabase",
	Long:  `Lists all of the environments that you have pushed to Supabase.`,
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current user
		user := auth.GetCurrentUser()
		if user == nil {
			log.Fatal("You must be logged in to list your environments.")
		}

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey, user.AccessToken)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// List the environments
		environments, err := supabaseClient.ListEnvironments(context.Background(), user.ID)
		if err != nil {
			log.Fatalf("Failed to list environments: %v", err)
		}

		// Print the environments
		if len(environments) == 0 {
			fmt.Println("You don't have any environments stored in Supabase.")
			return
		}

		fmt.Println("Your environments:")
		for _, env := range environments {
			fmt.Printf("- %s\n", env.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
