package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

var (
	cliBinaryPath string
)

// TestMain handles the setup and teardown for integration tests.
// It builds the CLI binary before running tests and cleans it up afterward.
func TestMain(m *testing.M) {
	// Determine the project root and the path for the test binary.
	projectRoot, err := getProjectRoot()
	if err != nil {
		fmt.Printf("Error getting project root: %v\n", err)
		os.Exit(1)
	}

	cliBinaryPath = filepath.Join(projectRoot, "stackmatch-cli-test")
	if runtime.GOOS == "windows" {
		cliBinaryPath += ".exe"
	}

	// Build the CLI binary.
	buildCmd := exec.Command("go", "build", "-o", cliBinaryPath, projectRoot)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build CLI binary: %v\nOutput: %s\n", err, string(buildOutput))
		os.Exit(1)
	}

	// Run the tests.
	exitCode := m.Run()

	// Clean up the binary.
	os.Remove(cliBinaryPath)

	os.Exit(exitCode)
}

func extractJSONOutput(output []byte) ([]byte, error) {
	// Find the start of JSON (first '{' character)
	start := bytes.IndexByte(output, '{')
	if start == -1 {
		return nil, fmt.Errorf("no JSON data found in output")
	}
	// Return everything from the first '{' to the end
	return output[start:], nil
}

func TestScanCommand(t *testing.T) {
	cmd := exec.Command(cliBinaryPath, "scan")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run scan command: %v\nOutput: %s", err, string(output))
	}

	// Extract just the JSON part of the output
	jsonOutput, err := extractJSONOutput(output)
	if err != nil {
		t.Fatalf("failed to extract JSON from output: %v\nOutput: %s", err, string(output))
	}

	var envData types.EnvironmentData
	if err := json.Unmarshal(jsonOutput, &envData); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v\nJSON: %s", err, string(jsonOutput))
	}

	if envData.System.OS == "" {
		t.Error("expected System.OS to be populated, but it was empty")
	}
	if envData.System.Arch == "" {
		t.Error("expected System.Arch to be populated, but it was empty")
	}
	if envData.StackmatchVersion == "" {
		t.Error("expected StackmatchVersion to be populated, but it was empty")
	}
}

func TestImportCommand_DryRun(t *testing.T) {
	// First, create a test environment file by running scan
	scanCmd := exec.Command(cliBinaryPath, "scan")
	scanOutput, err := scanCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run scan command: %v\nOutput: %s", err, string(scanOutput))
	}

	// Extract just the JSON part of the output
	if _, err := extractJSONOutput(scanOutput); err != nil {
		t.Fatalf("failed to extract JSON from scan output: %v\nOutput: %s", err, string(scanOutput))
	}

	// Create a temporary file for the environment data
	tempFile, err := os.CreateTemp("", "stackmatch-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the JSON part of the output to the temp file
	jsonOutput, err := extractJSONOutput(scanOutput)
	if err != nil {
		t.Fatalf("failed to extract JSON from scan output: %v\nOutput: %s", err, string(scanOutput))
	}
	
	if _, err := tempFile.Write(jsonOutput); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test the import command in dry-run mode (default)
	importCmd := exec.Command(cliBinaryPath, "import", tempFile.Name())
	output, err := importCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run import command: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Check for expected output
	expectedStrings := []string{
		"Environment Summary from",
		"System Information:",
		"Note: This is a dry run. No changes have been made to your system.",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(outputStr, s) {
			t.Errorf("expected output to contain %q, but it didn't. Output: %s", s, outputStr)
		}
	}
}

// TestImportCommand_Installation is a test that would actually install packages.
// This is commented out by default as it would modify the system.
// Uncomment and modify as needed for testing on a disposable environment.
/*
func TestImportCommand_Installation(t *testing.T) {
	// Create a minimal test environment file
	testEnv := types.EnvironmentData{
		StackmatchVersion: "test-version",
		ScanDate:         time.Now(),
		System: types.SystemInfo{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Tools: map[string]string{
			"git": "2.30.0",
		},
	}

	envData, err := json.Marshal(testEnv)
	if err != nil {
		t.Fatalf("failed to marshal test environment: %v", err)
	}

	// Create a temporary file for the environment data
	tempFile, err := os.CreateTemp("", "stackmatch-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the test data to the temp file
	if _, err := tempFile.Write(envData); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Run the import command with --no-dry-run
	importCmd := exec.Command(cliBinaryPath, "import", "--no-dry-run", tempFile.Name())
	output, err := importCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run import command: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Import command output:\n%s", outputStr)

	// Check for expected output
	expectedStrings := []string{
		"Starting installation",
		"Using package manager:",
		"Installation completed in",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(outputStr, s) {
			t.Errorf("expected output to contain %q, but it didn't. Output: %s", s, outputStr)
		}
	}
}
*/

// getProjectRoot finds the project root directory by looking for the go.mod file.
func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		if wd == filepath.Dir(wd) {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		wd = filepath.Dir(wd)
	}
}
