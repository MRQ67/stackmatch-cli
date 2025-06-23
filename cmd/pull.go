package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/spf13/cobra"
)

var (
	pullOutput string
	listOnly  bool
)

// envRow represents a row from the environments table
type envRow struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
	UserID    string          `json:"user_id"`
	IsPublic  bool            `json:"is_public"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

var pullCmd = &cobra.Command{
	Use:   "pull [environment-name]",
	Short: "Pull your latest or a specific environment",
	Long: `Pulls your most recent environment or a specific named environment.

Examples:
  stackmatch pull           # Pull latest environment
  stackmatch pull my-env    # Pull environment named 'my-env'
`,
	Args:  cobra.MaximumNArgs(1),
	PreRunE: requireAuth,
	Run: runPullCommand,
}

func runPullCommand(cmd *cobra.Command, args []string) {
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

	ctx := context.Background()
	var envData json.RawMessage
	envName := ""
	if len(args) > 0 {
		envName = args[0]
	}

	// Query for environment
	env, err := getEnvironment(ctx, supabaseClient, currentUser.ID, envName)
	if err != nil {
		log.Fatal(err)
	}

	// Extract environment data
	envData = env.Data

	// Output based on flags
	if listOnly {
		// Display environment details
		fmt.Printf("Environment: %s\n", env.Name)
		fmt.Printf("ID: %s\n", env.ID)
		fmt.Printf("Created: %s\n", env.CreatedAt.Format(time.RFC1123))
		fmt.Printf("Size: %d bytes\n", len(envData))
		return
	}

	if pullOutput != "" {
		if err := os.WriteFile(pullOutput, envData, 0o644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		fmt.Printf("Environment '%s' saved to %s\n", env.Name, pullOutput)
	} else {
		fmt.Println(string(envData))
	}
}

// getEnvironment retrieves either the latest or a named environment for the user
func getEnvironment(ctx context.Context, client *supabase.Client, userID, envName string) (*envRow, error) {
	var envs []envRow
	query := client.From("environments").
		Select("*", "exact", false).
		Eq("user_id", userID)

	if envName != "" {
		query = query.Eq("name", envName)
	}

	_, err := query.
		Limit(1, "").
		ExecuteTo(&envs)

	if err != nil {
		return nil, fmt.Errorf("failed to query environments: %w", err)
	}

	if len(envs) == 0 {
		if envName != "" {
			return nil, fmt.Errorf("environment '%s' not found. Use 'stackmatch list' to see available environments", envName)
		}
		return nil, fmt.Errorf("no environments found. Push your first environment with 'stackmatch push'")
	}

	return &envs[0], nil
}

func init() {
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", "", "Save output to a file instead of stdout")
	pullCmd.Flags().BoolVarP(&listOnly, "list-only", "l", false, "Only list environment details without downloading")
	rootCmd.AddCommand(pullCmd)
}
