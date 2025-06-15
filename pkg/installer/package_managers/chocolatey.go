package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
)

type chocolatey struct {
	*basePackageManager
}

// NewChocolatey creates a new Chocolatey package manager instance
func NewChocolatey() types.Installer {
	return &chocolatey{
		basePackageManager: &basePackageManager{
			name:           "Chocolatey",
			pmType:         types.TypeChocolatey,
			executableName: "choco",
		},
	}
}

func (c *chocolatey) InstallPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := c.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package
	_, err = c.runCommand(ctx, "install", pkg, "-y")
	if err != nil {
		// Check if package was not found
		if strings.Contains(err.Error(), "The package was not found") {
			return &types.PackageNotFoundError{
				Package: pkg,
			}
		}
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

func (c *chocolatey) InstallMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Chocolatey can install multiple packages in one command
	args := append([]string{"install"}, packages...)
	args = append(args, "-y") // Assume yes to all prompts

	_, err := c.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (c *chocolatey) UpdatePackageManager(ctx context.Context) error {
	_, err := c.runCommand(ctx, "upgrade", "chocolatey", "-y")
	return err
}

// checkIfInstalled overrides the base implementation with Chocolatey-specific logic
func (c *chocolatey) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := c.runCommand(ctx, "list", "--local-only", pkg)
	if err != nil {
		return false, err
	}

	// Check if the package appears in the list of installed packages
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 1 {
		// First line is header, check subsequent lines
		for _, line := range lines[1:] {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), strings.ToLower(pkg)) {
				return true, nil
			}
		}
	}

	return false, nil
}
