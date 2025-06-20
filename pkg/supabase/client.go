package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

// SaveEnvironment saves an environment to Supabase
func (c *Client) SaveEnvironment(ctx context.Context, env *types.EnvironmentData) (string, error) {
	// Set the scan date to now if not set
	if env.ScanDate.IsZero() {
		env.ScanDate = time.Now()
	}

	envJSON, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("failed to marshal environment data: %w", err)
	}

	// Log the data we're about to send (first 200 chars to avoid huge logs)
	logStr := string(envJSON)
	if len(logStr) > 200 {
		logStr = logStr[:200] + "..."
	}
	log.Printf("Saving environment data (truncated): %s", logStr)

	// Create a map with all required fields for the environments table
	// Note: user_id is automatically set by the database trigger
	insertData := map[string]interface{}{
		"name":      "My Environment", // Default name, could be made configurable
		"data":      json.RawMessage(envJSON),
		"is_public": false, // Default to private
	}

	// Try to insert the data
	var result []map[string]interface{}
	_, err = c.From("environments").
		Insert(insertData, false, "", "", "*").
		ExecuteTo(&result)

	if err != nil {
		log.Printf("Error inserting into Supabase: %v", err)
		return "", fmt.Errorf("failed to save environment to Supabase: %w", err)
	}

	log.Printf("Insert result: %+v", result)

	// Try to get the ID from the result if available
	if len(result) > 0 {
		if id, ok := result[0]["id"].(string); ok && id != "" {
			log.Printf("Successfully saved environment with ID: %s", id)
			return id, nil
		}
	}

	// If we couldn't get the ID from the result, try to fetch the most recent environment
	var dbEnvs []map[string]interface{}
	_, err = c.From("environments").
		Select("id", "", false).
		Order("created_at", nil).
		Limit(1, "").
		ExecuteTo(&dbEnvs)
	
	if err != nil {
		log.Printf("Error fetching recent environments: %v", err)
	} else if len(dbEnvs) > 0 {
		if id, ok := dbEnvs[0]["id"].(string); ok && id != "" {
			log.Printf("Found recent environment with ID: %s", id)
			return id, nil
		}
	}

	log.Println("Warning: Could not retrieve ID of inserted record, but insert may have succeeded")
	return "unknown-id", nil
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

// FindEnvironmentByUserAndName finds an environment by username and environment name
func (c *Client) FindEnvironmentByUserAndName(ctx context.Context, username, envName string) (*types.EnvironmentData, error) {
	var envs []types.EnvironmentData

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

	// Then find the environment by name and user ID
	_, err = c.From("environments").
		Select("*", "exact", false).
		Eq("name", envName).
		Eq("user_id", userID).
		ExecuteTo(&envs)

	if err != nil {
		return nil, fmt.Errorf("failed to find environment: %w", err)
	}

	if len(envs) == 0 {
		return nil, fmt.Errorf("environment '%s' not found for user '%s'", envName, username)
	}

	return &envs[0], nil
}
