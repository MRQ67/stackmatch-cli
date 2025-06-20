package package_managers

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

// basePackageManager provides common functionality for all package managers
type basePackageManager struct {
	name           string
	pmType        types.PackageManagerType
	executableName string
}

func (b *basePackageManager) Name() string {
	return b.name
}

// Type returns the package manager type
func (b *basePackageManager) Type() types.PackageManagerType {
	return b.pmType
}

func (b *basePackageManager) IsAvailable() bool {
	_, err := exec.LookPath(b.executableName)
	return err == nil
}

// runCommand is a helper method to run shell commands
func (b *basePackageManager) runCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, b.executableName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

