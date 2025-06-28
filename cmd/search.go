package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for public environments in Supabase",
	Long:  `Searches for public environments in Supabase that you can clone.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the search query
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Search for environments
		environments, err := supabaseClient.SearchEnvironments(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to search for environments: %v", err)
		}

		// Print the environments
		if len(environments) == 0 {
			fmt.Println("No public environments found.")
			return
		}

		fmt.Println("Public environments:")
		for _, env := range environments {
			fmt.Printf("- %s by %s\n", env.Name, env.Username)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
