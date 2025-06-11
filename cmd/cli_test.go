package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestScanCommand(t *testing.T) {
	cmd := exec.Command(cliBinaryPath, "scan")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run scan command: %v\nOutput: %s", err, string(output))
	}

	var envData types.EnvironmentData
	if err := json.Unmarshal(output, &envData); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
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
