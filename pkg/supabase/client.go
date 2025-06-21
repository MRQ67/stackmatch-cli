package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MRQ67/stackmatch-cli/pkg/auth"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	supabase "github.com/supabase-community/supabase-go"
)

// Client represents a Supabase client for the StackMatch CLI
type Client struct {
	*supabase.Client
	url string
	key string
}


// NewClient creates a new Supabase client
func NewClient(url, key string, accessToken ...string) (*Client, error) {
	// Create client options
	opts := &supabase.ClientOptions{
		Headers: make(map[string]string),
	}
	
	// Set access token if provided
	if len(accessToken) > 0 && accessToken[0] != "" {
		opts.Headers["Authorization"] = "Bearer " + accessToken[0]
		
		// Get user ID from the token (format: xxxx-xxxx-xxxx-xxxx)
		// We'll extract it from the token claims if needed
	}

	client, err := supabase.NewClient(url, key, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &Client{
		Client: client,
		url:    url,
		key:    key,
	}, nil
}

// SaveEnvironment saves an environment to Supabase with the given name and visibility
// If name is empty, it will use a default name
// isPublic determines if the environment is visible to other users
func (c *Client) SaveEnvironment(ctx context.Context, env *types.EnvironmentData, name string, isPublic bool) (string, error) {
	// Get the current user ID from the context
	userID := ""
	if user, ok := ctx.Value("user").(*auth.User); ok && user != nil {
		userID = user.ID
	}

	if userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}

	// Set the scan date to now if not set
	if env.ScanDate.IsZero() {
		env.ScanDate = time.Now()
	}

	// Convert environment data to JSON
	envJSON, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("failed to marshal environment data: %w", err)
	}

	// Log the data being saved (truncated for brevity)
	log.Printf("Saving environment data (truncated): %s", string(envJSON)[:min(100, len(envJSON))])

	// Use provided name or default to a generic name with timestamp
	if name == "" {
		name = fmt.Sprintf("Environment %s", time.Now().Format("2006-01-02 15:04"))
	}

	// Prepare the data to insert
	insertData := map[string]interface{}{
		"name":      name,
		"data":      json.RawMessage(envJSON),
		"is_public": isPublic,
		"user_id":   userID,
	}

	// Create a new client with the service role key
	serviceClient, err := NewClient(c.url, c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create service client: %w", err)
	}

	// Insert the data using the service client
	var result []map[string]interface{}
	_, err = serviceClient.Client.From("environments").
		Insert(insertData, false, "", "", "").
		ExecuteTo(&result)

	if err != nil {
		return "", fmt.Errorf("failed to save environment: %w", err)
	}

	// Extract the ID from the result
	if len(result) == 0 || result[0]["id"] == nil {
		return "", fmt.Errorf("no ID returned from insert")
	}

	envID, ok := result[0]["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid ID type returned: %T", result[0]["id"])
	}

	return envID, nil
}

// envRow represents a row in the environments table
type envRow struct {
	ID        string          `json:"id"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Data      json.RawMessage `json:"data"`
}

// GetEnvironment retrieves an environment from Supabase by ID
func (c *Client) GetEnvironment(ctx context.Context, id string) (*types.EnvironmentData, error) {
	var rows []envRow

	// Remove the "eq." prefix from the ID if it exists
	if len(id) > 3 && id[:3] == "eq." {
		id = id[3:]
	}

	// Select only the data column and filter by ID
	_, err := c.From("environments").
		Select("data", "", false).
		Eq("id", id).
		ExecuteTo(&rows)

	if err != nil {
		return nil, fmt.Errorf("failed to get environment from Supabase: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("environment not found with id: %s", id)
	}

	// Unmarshal the JSON data into EnvironmentData
	var envData types.EnvironmentData
	if err := json.Unmarshal(rows[0].Data, &envData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment data: %w", err)
	}

	return &envData, nil
}

// ListEnvironments retrieves a list of all environments
func (c *Client) ListEnvironments(ctx context.Context) ([]types.EnvironmentData, error) {
	var envs []types.EnvironmentData

	// Execute the query
	_, err := c.From("environments").Select("*", "exact", false).ExecuteTo(&envs)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	return envs, nil
}

// GetEnvironmentHistory retrieves the version history of an environment
func (c *Client) GetEnvironmentHistory(ctx context.Context, envID string, limit int) ([]types.EnvironmentHistory, error) {
	var history []types.EnvironmentHistory

	// Build the query
	query := c.From("environment_history").Select("*", "exact", false)
	
	// Add environment ID filter if provided
	if envID != "" {
		query = query.Eq("environment_id", envID)
	}
	
	// Order by created_at in descending order
	// Note: The Supabase Go client's Order method expects a column name and an optional ascending parameter
	// We'll use raw SQL for ordering to ensure it works as expected
	query = query.Order("created_at desc", nil)
	
	// Apply limit if specified
	// The second parameter is the foreign table name, which is empty for the main table
	if limit > 0 {
		query = query.Limit(limit, "")
	}

	// Execute the query
	_, err := query.ExecuteTo(&history)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment history: %w", err)
	}

	return history, nil
}

// envWithData represents the structure of an environment row in the database
type envWithData struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	UserID    string          `json:"user_id"`
	IsPublic  bool            `json:"is_public"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// FindEnvironmentByUserAndName finds an environment by username and environment name
func (c *Client) FindEnvironmentByUserAndName(ctx context.Context, username, envName string) (*types.EnvironmentData, error) {
	// First, find the user ID by username
	var users []map[string]interface{}
	_, err := c.From("profiles").
		Select("id", "exact", false).
		Eq("username", username).
		ExecuteTo(&users)

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user '%s' not found", username)
	}

	userID, ok := users[0]["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Find the environment by name and user ID
	var envRows []envWithData
	_, err = c.From("environments").
		Select("*", "exact", false).
		Eq("name", envName).
		Eq("user_id", userID).
		ExecuteTo(&envRows)

	if err != nil {
		return nil, fmt.Errorf("failed to find environment: %w", err)
	}

	if len(envRows) == 0 {
		return nil, fmt.Errorf("environment '%s' not found for user '%s'", envName, username)
	}

	// Unmarshal the JSON data into EnvironmentData
	var envData types.EnvironmentData
	if err := json.Unmarshal(envRows[0].Data, &envData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment data: %w", err)
	}

	return &envData, nil
}
