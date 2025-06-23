package package_managers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/version"
)

type chocolatey struct {
	*basePackageManager
}

// NewChocolatey creates a new Chocolatey package manager instance
func NewChocolatey() types.Installer {
	c := &chocolatey{
		basePackageManager: &basePackageManager{
			name:           "Chocolatey",
			pmType:         types.TypeChocolatey,
			executableName: "choco",
			versionCommand: "list --local-only --exact",
			versionRegex:   `([0-9]+\.[0-9]+(?:\.[0-9]+(?:\.[0-9]+)?)?)`,
			installWithFlags: true,
		},
	}
	c.installPackageFunc = c.installPackage
	c.installMultipleFunc = c.installMultiple
	c.uninstallPackageFunc = c.uninstallPackage
	return c
}

// installPackage installs a single package
func (c *chocolatey) installPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := c.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --yes to avoid prompts
	_, err = c.runCommand(ctx, "install", "--yes", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

// uninstallPackage uninstalls a package using Chocolatey
func (c *chocolatey) uninstallPackage(ctx context.Context, pkg string) error {
	// First check if installed
	installed, err := c.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if !installed {
		return nil // Already uninstalled
	}

	// Uninstall the package with --yes to avoid prompts
	_, err = c.runCommand(ctx, "uninstall", "--yes", pkg)
	if err != nil {
		return fmt.Errorf("failed to uninstall package: %w", err)
	}

	return nil
}

// InstallPackage implements the Installer interface
func (c *chocolatey) InstallPackage(ctx context.Context, pkg string) error {
	return c.installPackage(ctx, pkg)
}

// InstallVersion installs a specific version of a package
func (c *chocolatey) InstallVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) error {
	// Check if already installed with the required version
	info, err := c.CheckVersion(ctx, pkg, constraint)
	if err != nil {
		return fmt.Errorf("failed to check package version: %w", err)
	}

	if info.Satisfies {
		return nil // Already installed with the required version
	}

	// Get available versions
	versions, err := c.getAvailableVersions(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to get available versions: %w", err)
	}

	// Find a version that satisfies the constraint
	var selectedVersion string
	for _, v := range versions {
		ver, err := version.Parse(v)
		if err != nil {
			continue
		}
		if satisfies, _ := ver.Satisfies(constraint.Version); satisfies {
			selectedVersion = v
			break
		}
	}

	if selectedVersion == "" {
		return fmt.Errorf("no version found matching constraint: %s", constraint.Version)
	}

	// Install the specific version
	_, err = c.runCommand(ctx, "install", pkg, "--version", selectedVersion, "-y")
	if err != nil {
		return fmt.Errorf("failed to install package version %s: %w", selectedVersion, err)
	}

	return nil
}

// installMultiple installs multiple packages in a single operation
func (c *chocolatey) installMultiple(ctx context.Context, packages []string) error {
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

// InstallMultiple implements the Installer interface
func (c *chocolatey) InstallMultiple(ctx context.Context, packages []string) error {
	return c.installMultiple(ctx, packages)
}

// InstallMultipleVersions installs multiple packages with specific versions
func (c *chocolatey) InstallMultipleVersions(ctx context.Context, packages map[string]types.VersionConstraint) error {
	if len(packages) == 0 {
		return nil
	}

	// Install each package with its version constraint
	for pkg, constraint := range packages {
		if constraint.Version != "" {
			if err := c.InstallVersion(ctx, pkg, constraint); err != nil {
				return fmt.Errorf("failed to install %s@%s: %w", pkg, constraint.Version, err)
			}
		} else {
			if err := c.InstallPackage(ctx, pkg); err != nil {
				return fmt.Errorf("failed to install %s: %w", pkg, err)
			}
		}
	}

	return nil
}

// UninstallPackage uninstalls a package
func (c *chocolatey) UninstallPackage(ctx context.Context, pkg string) error {
	return c.uninstallPackage(ctx, pkg)
}

// getAvailableVersions gets all available versions for a package
func (c *chocolatey) getAvailableVersions(ctx context.Context, pkg string) ([]string, error) {
	output, err := c.runCommand(ctx, "find", pkg, "--all-versions")
	if err != nil {
		return nil, fmt.Errorf("failed to find package versions: %w", err)
	}

	// Parse the output to extract versions
	re := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?(?:\-[\w\.]+)?)`)
	matches := re.FindAllStringSubmatch(output, -1)

	var versions []string
	for _, match := range matches {
		if len(match) > 1 {
			versions = append(versions, match[1])
		}
	}

	return versions, nil
}

// GetInstalledVersion gets the installed version of a package
func (c *chocolatey) GetInstalledVersion(ctx context.Context, pkg string) (*types.PackageVersionInfo, error) {
	output, err := c.runCommand(ctx, "list", "--local-only", pkg)
	if err != nil {
		// If the package is not installed, return empty version
		if strings.Contains(err.Error(), "The package was not found") {
			return &types.PackageVersionInfo{
				Name: pkg,
			}, nil
		}
		return nil, fmt.Errorf("failed to get installed version: %w", err)
	}

	// Parse the output which is in format: "pkg version"
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return &types.PackageVersionInfo{
			Name: pkg,
		}, nil
	}

	// Find the line with our package
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower(pkg)) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return &types.PackageVersionInfo{
					Name:    pkg,
					Version: parts[1],
				}, nil
			}
		}
	}

	return &types.PackageVersionInfo{
		Name: pkg,
	}, nil
}

// CheckVersion checks if the installed package satisfies the version constraint
func (c *chocolatey) CheckVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) (*types.PackageVersionInfo, error) {
	// First get the installed version
	info, err := c.GetInstalledVersion(ctx, pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed version: %w", err)
	}

	// If not installed, return early
	if info.Version == "" {
		info.Satisfies = false
		info.Constraint = constraint.Version
		return info, nil
	}

	// Parse the installed version
	installedVer, err := version.Parse(info.Version)
	if err != nil {
		info.Satisfies = false
		info.Constraint = constraint.Version
		return info, nil
	}

	// Check if it satisfies the constraint
	satisfies, err := installedVer.Satisfies(constraint.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version constraint: %w", err)
	}

	info.Satisfies = satisfies
	info.Constraint = constraint.Version
	return info, nil
}

func (c *chocolatey) UpdatePackageManager(ctx context.Context) error {
	_, err := c.runCommand(ctx, "upgrade", "chocolatey", "-y")
	return err
}

// checkIfInstalled overrides the base implementation with Chocolatey-specific logic
func (c *chocolatey) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	info, err := c.GetInstalledVersion(ctx, pkg)
	if err != nil {
		return false, err
	}
	return info.Version != "", nil
}
