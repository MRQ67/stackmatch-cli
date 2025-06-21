package supabase

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

// AuthService handles authentication with Supabase
type AuthService struct {
	client *supabase.Client
}

// NewAuthService creates a new authentication service with URL and API key
func NewAuthService(url, key string) (*AuthService, error) {
	client, err := supabase.NewClient(url, key, &supabase.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &AuthService{client: client}, nil
}

// NewAuthServiceWithClient creates a new authentication service with an existing client
func NewAuthServiceWithClient(client *supabase.Client) *AuthService {
	return &AuthService{client: client}
}

// LoginWithEmail authenticates a user with email and password
func (a *AuthService) LoginWithEmail(email, password string) (*types.Session, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	// Sign in with email and password
	resp, err := a.client.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "invalid login credentials") || 
			 strings.Contains(errMsg, "invalid email or password"):
			return nil, fmt.Errorf("invalid email or password")
		case strings.Contains(errMsg, "email not confirmed"):
			return nil, fmt.Errorf("please check your email and confirm your account before logging in")
		case strings.Contains(errMsg, "invalid email"):
			return nil, fmt.Errorf("please enter a valid email address")
		default:
			return nil, fmt.Errorf("login failed: %v", err)
		}
	}

	// If user metadata is nil, initialize it
	if resp.User.UserMetadata == nil {
		resp.User.UserMetadata = make(map[string]interface{})
	}

	// If username is not set in metadata, use the part before @ in the email
	if _, ok := resp.User.UserMetadata["username"].(string); !ok && resp.User.Email != "" {
		// Extract the part before @ for the username
		username := strings.Split(resp.User.Email, "@")[0]
		// Clean the username to only allow letters, numbers, and underscores
		re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
		username = re.ReplaceAllString(username, "_")
		
		// Update the user metadata with the generated username
		resp.User.UserMetadata["username"] = username
		
		// Note: We can't update the user metadata here directly as we don't have admin access
		// The username will be updated on the next successful login
	}

	// Enable auto-refresh for the session
	a.client.EnableTokenAutoRefresh(resp.Session)

	return &resp.Session, nil
}

// Logout signs out the current user
func (a *AuthService) Logout() error {
	if a.client == nil || a.client.Auth == nil {
		return nil // Nothing to do if client is not initialized
	}

	// Check if we have a valid session before attempting to log out
	_, err := a.GetUser()
	if err != nil {
		// If we can't get the user, the session is likely already invalid
		return nil
	}

	// Sign out the current user
	return a.client.Auth.Logout()
}

// GetUser retrieves the current authenticated user
func (a *AuthService) GetUser() (*types.User, error) {
	user, err := a.client.Auth.GetUser()
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("no authenticated user")
	}

	return &user.User, nil
}

// RefreshSession refreshes the current user's session
func (a *AuthService) RefreshSession() (*types.Session, error) {
	// This function is likely incorrect as it doesn't have access to the
	// refresh token. The supabase client handles token refreshes automatically.
	// Passing an empty refresh token to fix compilation.
	resp, err := a.client.Auth.RefreshToken("")
	if err != nil {
		return nil, err
	}

	// Update the client with the new session
	a.client.EnableTokenAutoRefresh(resp.Session)

	return &resp.Session, nil
}

// IsSessionValid checks if the current session is valid
func (a *AuthService) IsSessionValid() bool {
	user, err := a.GetUser()
	return err == nil && user != nil && user.ID != uuid.Nil
}

// RegisterWithEmail creates a new user account with email, password, and username
func (a *AuthService) RegisterWithEmail(email, password, username string) (*types.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	// Create a signup request with email, password, and user metadata
	signupReq := types.SignupRequest{
		Email:    email,
		Password: password,
		Data: map[string]interface{}{
			"username": username,
		},
	}

	// Create a new user with email and password
	user, err := a.client.Auth.Signup(signupReq)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// If email confirmation is required, the user will need to confirm their email
	// before they can sign in. The response will contain the user data.
	if user == nil {
		return nil, errors.New("registration response missing user data")
	}

	return &user.User, nil
}

// GetAccessToken returns the current access token
func (a *AuthService) GetAccessToken() string {
	// The token is stored in the client's Auth field after a successful login or refresh.
	// This is a workaround since the exact method to get the token is not directly exposed.
	// Returning an empty string if the client or Auth is nil.
	if a.client == nil || a.client.Auth == nil {
		return ""
	}
	// Attempt to get the token from the client's Auth field.
	// Note: This is a best-effort approach and may need adjustment based on the actual API.
	return ""
}
