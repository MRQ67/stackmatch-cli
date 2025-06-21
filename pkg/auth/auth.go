package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

// User represents an authenticated user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// FromSupabaseSession creates a User from Supabase session
func FromSupabaseSession(session *types.Session) *User {
	if session == nil || session.User.Email == "" {
		return nil
	}

	// Convert UUID to string
	userID := ""
	if session.User.ID != (uuid.UUID{}) { // Check for zero UUID
		userID = session.User.ID.String()
	}

	// Calculate expiration time
	expiresIn := 3600 // Default to 1 hour if not specified
	if session.ExpiresIn > 0 {
		expiresIn = int(session.ExpiresIn)
	}

	// Extract username from user metadata if available
	username := ""
	if session.User.UserMetadata != nil {
		// Marshal the UserMetadata to JSON and then unmarshal it to a map
		metadataJSON, err := json.Marshal(session.User.UserMetadata)
		if err == nil {
			var meta map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &meta); err == nil {
				if uname, ok := meta["username"].(string); ok {
					username = uname
				}
			}
		}

		// If username is still empty, try to get it from the email
		if username == "" && session.User.Email != "" {
			username = session.User.Email
		}
	}

	return &User{
		ID:           userID,
		Email:        session.User.Email,
		Username:     username,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Second * time.Duration(expiresIn)),
	}
}

var (
	currentUser *User
	mu          sync.RWMutex
	sessionFile string
	initialized bool

	// ErrNotAuthenticated is returned when a user is not authenticated
	ErrNotAuthenticated = fmt.Errorf("not authenticated")
	// ErrSessionExpired is returned when the session has expired
	ErrSessionExpired = fmt.Errorf("session expired")
)

func init() {
	// Set up session file path with proper permissions
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	sessionFile = filepath.Join(home, ".stackmatch", "session.json")
	initialized = true
}

// SaveSession saves a user session to disk
func SaveSession(user *User) error {
	if user == nil {
		return fmt.Errorf("cannot save nil user session")
	}

	// Update the current user in memory
	mu.Lock()
	currentUser = user
	mu.Unlock()

	// Marshal the user data
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(sessionFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to a temporary file first
	tempFile := sessionFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	// Atomically rename the temp file
	if err := os.Rename(tempFile, sessionFile); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// Logout removes the current user session
func Logout() error {
	mu.Lock()
	defer mu.Unlock()

	currentUser = nil

	// Remove session file
	if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	return nil
}

// GetCurrentUser returns the currently authenticated user, if any
func GetCurrentUser() *User {
	if !initialized {
		return nil
	}

	// Check memory first with a read lock
	mu.RLock()
	user := currentUser
	mu.RUnlock()

	if user != nil {
		if time.Until(user.ExpiresAt) > 5*time.Minute {
			return user
		}
		// Session is about to expire, try to refresh
		if refreshed := tryRefreshSession(user); refreshed != nil {
			return refreshed
		}
	}

	// Try to load from disk
	user, err := loadSession()
	if err != nil {
		return nil
	}

	// Check if session is still valid
	if time.Now().After(user.ExpiresAt) {
		_ = os.Remove(sessionFile)
		return nil
	}

	// Update in-memory cache
	mu.Lock()
	currentUser = user
	mu.Unlock()

	return user
}

// tryRefreshSession attempts to refresh an expiring session
func tryRefreshSession(user *User) *User {
	if user == nil || user.RefreshToken == "" {
		return nil
	}

	// TODO: Implement token refresh using Supabase client
	// This requires the Supabase client to be available in this package
	// For now, we'll just return nil to indicate refresh wasn't possible
	return nil
}

// IsAuthenticated checks if there is a valid user session
func IsAuthenticated() bool {
	return GetCurrentUser() != nil
}

// GetUserFromContext retrieves the user from a context
func GetUserFromContext(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value("user").(*User); ok {
		return user
	}
	return nil
}

// RequireAuth returns an error if no user is authenticated
func RequireAuth() error {
	if !IsAuthenticated() {
		return ErrNotAuthenticated
	}
	return nil
}

// loadSession loads a user session from disk
func loadSession() (*User, error) {
	if !initialized {
		return nil, fmt.Errorf("auth package not initialized")
	}

	// Check if session file exists
	fileInfo, err := os.Stat(sessionFile)
	if os.IsNotExist(err) {
		return nil, ErrNotAuthenticated
	} else if err != nil {
		return nil, fmt.Errorf("failed to access session file: %w", err)
	}

	// Check for empty file
	if fileInfo.Size() == 0 {
		_ = os.Remove(sessionFile)
		return nil, fmt.Errorf("session file is empty")
	}

	// Read the file
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		_ = os.Remove(sessionFile)
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal the user data
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		_ = os.Remove(sessionFile)
		return nil, fmt.Errorf("invalid session data: %w", err)
	}

	// Validate the loaded user
	if user.ID == "" || user.AccessToken == "" || user.Email == "" {
		_ = os.Remove(sessionFile)
		return nil, fmt.Errorf("invalid session data: missing required fields")
	}

	return &user, nil
}

// ClearSession removes the current session
func ClearSession() error {
	mu.Lock()
	defer mu.Unlock()

	currentUser = nil
	if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	return nil
}

// GetAccessToken returns the current access token if valid
func GetAccessToken() (string, error) {
	user := GetCurrentUser()
	if user == nil {
		return "", ErrNotAuthenticated
	}

	if time.Until(user.ExpiresAt) < 0 {
		return "", ErrSessionExpired
	}

	return user.AccessToken, nil
}
