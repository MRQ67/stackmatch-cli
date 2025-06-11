package exporter

import (
	"encoding/json"
	"os"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// WriteJSON serializes the EnvironmentData to a JSON file.
func WriteJSON(data types.EnvironmentData, filename string) error {
	// Marshal the data with pretty printing (indentation)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the specified file
	// The file will be created if it doesn't exist, or truncated if it does.
	// 0644 provides read/write for the owner, and read-only for group/others.
	return os.WriteFile(filename, jsonData, 0644)
}
