package cmd

import (
	"context"
	"fmt"
	"log"
	"text/tabwriter"
	"os"

	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
)

var (
	logLimit int
)

var logCmd = &cobra.Command{
	Use:   "log [environment-id]",
	Short: "View environment history",
	Long:  `Shows the history of an environment, including all versions and changes.`,
	Args:  cobra.MaximumNArgs(1),
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current environment ID from config or prompt
		envID := "" // TODO: Get current environment ID from config

		// Initialize Supabase client
		supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey)
		if err != nil {
			log.Fatalf("Failed to initialize Supabase client: %v", err)
		}

		// Get environment history
		ctx := context.Background()
		history, err := supabaseClient.GetEnvironmentHistory(ctx, envID, logLimit)
		if err != nil {
			log.Fatalf("Failed to get environment history: %v", err)
		}

		// Display history in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tName\tVersion\tCreated At\tUpdated By")
		for _, env := range history {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				env.ID,
				env.Name,
				env.Version,
				env.CreatedAt.Format("2006-01-02 15:04:05"),
				env.UpdatedBy,
			)
		}
		w.Flush()
	},
}

func init() {
	logCmd.Flags().IntVar(&logLimit, "limit", 10, "Maximum number of history entries to show")
	rootCmd.AddCommand(logCmd)
}
