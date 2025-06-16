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
	client *supabase.Client
}

// NewClient creates a new Supabase client
func NewClient(url, apiKey string) (*Client, error) {
	supabaseClient, err := supabase.NewClient(url, apiKey, &supabase.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &Client{
		client: supabaseClient,
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

	// Create a simple map with the data field
	insertData := map[string]interface{}{
		"data": json.RawMessage(envJSON),
	}

	// Try to insert the data
	var result []map[string]interface{}
	_, err = c.client.From("environments").
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
	_, err = c.client.From("environments").
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
	_, err := c.client.From("environments").
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
	var result []types.EnvironmentData

	_, err := c.client.From("environments").Select("*", "", false).ExecuteTo(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments from Supabase: %w", err)
	}

	return result, nil
}
