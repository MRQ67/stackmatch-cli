package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/spf13/cobra"
)

// envRow is defined in pull.go

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "List all your environments",
	Long:  `Lists all environments for the currently authenticated user.`,
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current user
		currentUser := auth.GetCurrentUser()
		if currentUser == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Get all environments for the current user
		var envs []envRow
		_, err = supabaseClient.From("environments").
			Select("*", "exact", false).
			Eq("user_id", currentUser.ID).
			ExecuteTo(&envs)

		if err != nil {
			log.Fatalf("Failed to get environments: %v", err)
		}

		if len(envs) == 0 {
			fmt.Println("No environments found. Push your first environment with 'stackmatch push'")
			return
		}

		// Display environments in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCREATED\tSIZE (bytes)")
		for _, env := range envs {
			createdAt := ""
			if !env.CreatedAt.IsZero() {
				createdAt = env.CreatedAt.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(w, "%s\t%s\t%d\n",
				env.Name,
				createdAt,
				len(env.Data),
			)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
