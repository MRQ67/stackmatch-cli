package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/scanner"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/spf13/cobra"
)

// promptForVisibility asks the user if the environment should be public
func promptForVisibility() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Make this environment public? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no")
		}
	}
}

// scanEnvironment scans the current development environment
func scanEnvironment() *types.EnvironmentData {
	envData := &types.EnvironmentData{
		Tools:              make(map[string]string),
		PackageManagers:    make(map[string]string),
		CodeEditors:        make(map[string]string),
		ConfiguredLanguages: make(map[string]string),
		ConfigFiles:        []string{},
		System:             types.SystemInfo{},
	}

	// Set scan timestamp
	envData.ScanDate = time.Now()

	// Detect system information
	scanner.DetectSystemInfo(&envData.System)

	// Detect package managers
	scanner.DetectPackageManagers(envData)

	// Detect programming languages
	scanner.DetectProgrammingLanguages(envData)

	// Detect development tools
	scanner.DetectTools(envData)
	// Detect code editors and IDEs
	scanner.DetectEditors(envData)
	// Scan for configuration files
	scanner.DetectConfigFiles(envData)

	return envData
}

var (
	isPublic bool
)

var pushCmd = &cobra.Command{
	Use:   "push [name] [flags]",
	Short: "Scan & push environment data to Supabase",
	Long: `Scans the current development environment and uploads the configuration to Supabase.
This requires authentication and Supabase URL/API key to be set.

If a name is not provided as an argument, you will be prompted to enter one.`,
	Args:  cobra.MaximumNArgs(1),
	PreRunE: requireAuth,
	Run: func(cmd *cobra.Command, args []string) {
		// Use the global authenticated Supabase client
		if supabaseClient == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Scan the environment
		envData := scanEnvironment()

		// Get the current user from the session
		user := auth.GetCurrentUser()
		if user == nil {
			log.Fatal("Not authenticated. Please run 'stackmatch login' first.")
		}

		// Get environment name from args or prompt
		envName := ""
		if len(args) > 0 {
			envName = args[0]
		}

		// If no name provided, prompt for one
		if envName == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter a name for this environment: ")
			input, _ := reader.ReadString('\n')
			envName = strings.TrimSpace(input)

			if envName == "" {
				envName = fmt.Sprintf("Environment %s", time.Now().Format("2006-01-02 15:04"))
			}
		}

		// Get visibility setting
		isEnvPublic := isPublic
		if !cmd.Flags().Changed("public") {
			// Prompt for visibility if not set via flag
			var err error
			isEnvPublic, err = promptForVisibility()
			if err != nil {
				log.Fatalf("Failed to get visibility preference: %v", err)
			}
		}

		// Add user to context
		ctx := context.WithValue(context.Background(), "user", user)

		// Upload to Supabase
		id, err := supabaseClient.SaveEnvironment(ctx, envData, envName, isEnvPublic)
		if err != nil {
			log.Fatalf("Failed to save environment: %v", err)
		}

		visibility := "private"
		if isEnvPublic {
			visibility = "public"
		}
		fmt.Printf("Successfully saved %s environment '%s' with ID: %s\n", visibility, envName, id)
	},
}

func init() {
	pushCmd.Flags().BoolVarP(&isPublic, "public", "p", false, "Make the environment publicly accessible")
	rootCmd.AddCommand(pushCmd)
}
