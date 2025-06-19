package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/supabase"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)



var (
	email    string
	password string
)

func init() {
	// Add auth commands to root
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(registerCmd)

	// Login flags
	loginCmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password (optional, will prompt if not provided)")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Supabase",
	Long:  `Authenticate with Supabase using email and password`,
	Run: func(cmd *cobra.Command, args []string) {
		if supabaseClient == nil {
			fmt.Fprintln(os.Stderr, "Error: Supabase client not initialized. Please check your configuration.")
			os.Exit(1)
		}

		// Check if already logged in
		if user := auth.GetCurrentUser(); user != nil {
			fmt.Printf("Already logged in as %s\n", user.Email)
			return
		}

		// Prompt for email
		fmt.Print("Email: ")
		var email string
		_, err := fmt.Scanln(&email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading email: %v\n", err)
			os.Exit(1)
		}

		// Prompt for password (hidden)
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println() // Print newline after password input

		password := string(bytePassword)

		// Initialize auth service
		authService, err := supabase.NewAuthService(supabaseURL, supabaseAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing auth service: %v\n", err)
			os.Exit(1)
		}

		// Call the auth service to handle login
		session, err := authService.LoginWithEmail(email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
			os.Exit(1)
		}

		// Convert session to auth.User and save
		authUser := auth.FromSupabaseSession(session)
		if authUser == nil {
			fmt.Fprintln(os.Stderr, "Failed to create user session")
			os.Exit(1)
		}

		if err := auth.SaveSession(authUser); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save session: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully logged in as %s\n", authUser.Email)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out the current user",
	Long:  `Sign out the currently authenticated user`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user is logged in
		user := auth.GetCurrentUser()
		if user == nil {
			fmt.Println("No active session found")
			return
		}

		// First try to sign out from Supabase while we still have the token
		if supabaseClient != nil {
			authService := supabase.NewAuthServiceWithClient(supabaseClient.Client)
			if err := authService.Logout(); err != nil {
				// Don't treat this as a fatal error since we'll clear the local session anyway
				fmt.Fprintf(os.Stderr, "Warning: Failed to sign out from Supabase: %v\n", err)
			}
		}

		// Clear local session
		if err := auth.ClearSession(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing session: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully logged out")
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Long:  `Register a new account with email and password`,
	Run: func(cmd *cobra.Command, args []string) {
		if supabaseClient == nil {
			fmt.Fprintln(os.Stderr, "Error: Supabase client not initialized. Please check your configuration.")
			os.Exit(1)
		}

		// Check if already logged in
		if user := auth.GetCurrentUser(); user != nil {
			fmt.Printf("Already logged in as %s. Please log out before registering a new account.\n", user.Email)
			return
		}

		// Prompt for email
		fmt.Print("Email: ")
		var email string
		_, err := fmt.Scanln(&email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading email: %v\n", err)
			os.Exit(1)
		}

		// Prompt for password (hidden)
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println() // Print newline after password input

		password := string(bytePassword)

		// Initialize auth service
		authService := supabase.NewAuthServiceWithClient(supabaseClient.Client)

		// Call the auth service to handle registration
		_, err = authService.RegisterWithEmail(email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Registration failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Registration successful! Please check your email (%s) to verify your account.\n", email)
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the current logged-in user",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current user from auth package
		user := auth.GetCurrentUser()
		if user == nil {
			fmt.Println("Not logged in")
			return
		}

		// Display basic user info
		fmt.Printf("Logged in as: %s\n", user.Email)
		if user.ID != "" {
			fmt.Printf("User ID: %s\n", user.ID)
		}

		// Show session info
		if !user.ExpiresAt.IsZero() {
			fmt.Printf("Session expires: %s\n", user.ExpiresAt.Format("2006-01-02 15:04:05"))
			remaining := time.Until(user.ExpiresAt).Round(time.Minute)
			if remaining > 0 {
				fmt.Printf("Time remaining: %s\n", remaining)
			} else {
				fmt.Println("Session has expired")
			}
		}

		// If we have a Supabase client, try to get fresh user info
		if supabaseClient != nil {
			authService := supabase.NewAuthServiceWithClient(supabaseClient.Client)
			supabaseUser, err := authService.GetUser()
			if err == nil {
				fmt.Println("\nFresh user info from Supabase:")
				if !supabaseUser.EmailConfirmedAt.IsZero() {
					fmt.Printf("Email confirmed: %v\n", !supabaseUser.EmailConfirmedAt.IsZero())
				}
				if !supabaseUser.LastSignInAt.IsZero() {
					fmt.Printf("Last sign in: %s\n", supabaseUser.LastSignInAt.Format("2006-01-02 15:04:05"))
				}
			}
		}
	},
}
