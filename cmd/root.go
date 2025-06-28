package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/config"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// Global configuration
	cfg *config.Config

	// Supabase configuration flags
	supabaseURL    string
	supabaseAPIKey string

	// Supabase client
	supabaseClient *supabase.Client

	rootCmd = &cobra.Command{
		Use:   "stackmatch",
		Short: "StackMatch: Clone environments, not just code.",
		Long:  `StackMatch is a CLI tool that helps developers scan, export, and import their development environment configurations.
It aims to eliminate "works on my machine" problems by providing a consistent way to manage development setups.`,
	}
)

func init() {
	// Initialize config
	cfg = config.New()

	// Add persistent flags for Supabase
	rootCmd.PersistentFlags().StringVarP(&supabaseURL, "supabase-url", "u", cfg.SupabaseURL, "Supabase project URL")
	rootCmd.PersistentFlags().StringVarP(&supabaseAPIKey, "supabase-key", "k", cfg.SupabaseAPIKey, "Supabase API key")

	// Bind flags to config
	if err := cfg.BindFlags(pflag.CommandLine); err != nil {
		log.Fatalf("Failed to bind flags: %v", err)
	}

	// Initialize Supabase client
	var err error
	supabaseClient, err = initSupabase()
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Add commands directly to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)

	// Persistent pre-run to validate config and handle flags
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Update config from flags if provided
		if err := cfg.BindFlags(pflag.CommandLine); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}

		// Skip config validation for auth commands
		switch cmd.Name() {
		case "login", "logout", "whoami", "register":
			return nil
		}

		// Validate config for other commands
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		// Initialize Supabase client
		var err error
		supabaseClient, err = initSupabase()
		if err != nil {
			return fmt.Errorf("failed to initialize Supabase client: %w", err)
		}

		return nil
	}

	// Save config on successful command execution (but not for auth commands)
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		switch cmd.Name() {
		case "login", "logout", "whoami":
			return nil // Skip saving for auth commands
		}

		if err := cfg.Save(); err != nil {
			log.Printf("Warning: Failed to save config: %v", err)
		}
		return nil
	}
}

// initSupabase initializes the Supabase client with the current configuration
func initSupabase() (*supabase.Client, error) {
	if supabaseURL == "" || supabaseAPIKey == "" {
		return nil, fmt.Errorf("supabase URL and API key must be set")
	}

	// Get access token if user is authenticated
	var accessToken string
	if user := auth.GetCurrentUser(); user != nil {
		accessToken = user.AccessToken
	}

	client, err := supabase.NewClient(supabaseURL, supabaseAPIKey, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	return client, nil
}

// requireAuth is a middleware that ensures the user is authenticated
func requireAuth(cmd *cobra.Command, args []string) error {
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required. Please run 'stackmatch login'")
	}
	return nil
}

// Execute runs the root command
func Execute() {
	// Initialize Supabase client
	var err error
	supabaseClient, err = initSupabase()
	if err != nil {
		log.Printf("Warning: Failed to initialize Supabase client: %v", err)
	}

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
