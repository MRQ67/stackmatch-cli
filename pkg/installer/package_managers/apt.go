package package_managers

import (
	"context"
	"fmt"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/version"
)

type apt struct {
	*basePackageManager
}

// NewApt creates a new APT package manager instance
func NewApt() types.Installer {
	pm := &apt{
		basePackageManager: &basePackageManager{
			name:           "APT",
			pmType:        types.TypeApt,
			executableName: "apt",
			versionCommand: "apt-cache",
			versionRegex:   `(\d+:)?([\d.~+-]+)(-[\w.+-]+)?`,
		},
	}
	pm.installPackageFunc = pm.installPackage
	pm.installMultipleFunc = pm.installMultiple
	return pm
}

// installPackage installs a single package
func (a *apt) installPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := a.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package with --assume-yes to avoid prompts
	_, err = a.runCommand(ctx, "install", "--assume-yes", pkg)
	if err != nil {
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

// InstallPackage implements the Installer interface
func (a *apt) InstallPackage(ctx context.Context, pkg string) error {
	return a.installPackage(ctx, pkg)
}

// InstallVersion installs a specific version of a package
func (a *apt) InstallVersion(ctx context.Context, pkg string, version types.VersionConstraint) error {
	// Check if the package is already installed with the required version
	info, err := a.CheckVersion(ctx, pkg, version)
	if err != nil {
		return fmt.Errorf("failed to check package version: %w", err)
	}

	if info.Satisfies {
		return nil // Already installed with the required version
	}

	// Format the package with version (e.g., "package=1.2.3")
	versionedPkg := fmt.Sprintf("%s=%s", pkg, version.Version)
	
	// Install the specific version
	_, err = a.runCommand(ctx, "install", "--assume-yes", "--allow-downgrades", versionedPkg)
	if err != nil {
		return fmt.Errorf("failed to install package version %s: %w", version.Version, err)
	}

	return nil
}

// installMultiple installs multiple packages in a single operation
func (a *apt) installMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// APT can install multiple packages in one command
	args := append([]string{"install", "--assume-yes"}, packages...)
	_, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

// InstallMultiple implements the Installer interface
func (a *apt) InstallMultiple(ctx context.Context, packages []string) error {
	return a.installMultiple(ctx, packages)
}

// InstallMultipleVersions installs multiple packages with specific versions
func (a *apt) InstallMultipleVersions(ctx context.Context, packages map[string]types.VersionConstraint) error {
	if len(packages) == 0 {
		return nil
	}

	// Prepare the package list with versions
	var pkgs []string
	for pkg, ver := range packages {
		if ver.Version != "" {
			pkg = fmt.Sprintf("%s=%s", pkg, ver.Version)
		}
		pkgs = append(pkgs, pkg)
	}

	// Install all packages with versions in one command
	args := append([]string{"install", "--assume-yes", "--allow-downgrades"}, pkgs...)
	_, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages with versions: %w", err)
	}

	return nil
}

// GetInstalledVersion gets the installed version of a package
func (a *apt) GetInstalledVersion(ctx context.Context, pkg string) (*types.PackageVersionInfo, error) {
	// First try to get the installed version using dpkg
	output, err := a.runCommand(ctx, "dpkg-query", "-W", "-f=${Version}\\n${Status}\\n", pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to query package version: %w", err)
	}

	// Parse the output to get version and status
	lines := strings.SplitN(output, "\n", 2)
	if len(lines) < 2 {
		return &types.PackageVersionInfo{
			Name: pkg,
		}, nil
	}

	// Check if the package is installed
	if !strings.Contains(lines[1], "install ok installed") {
		return &types.PackageVersionInfo{
			Name: pkg,
		}, nil
	}

	// Clean up the version string
	ver := strings.TrimSpace(lines[0])
	// Remove architecture if present (e.g., "1:2.0.0-1_amd64" -> "1:2.0.0-1")
	if idx := strings.LastIndex(ver, "_"); idx > 0 {
		ver = ver[:idx]
	}

	return &types.PackageVersionInfo{
		Name:    pkg,
		Version: ver,
	}, nil
}

// CheckVersion checks if the installed package satisfies the version constraint
func (a *apt) CheckVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) (*types.PackageVersionInfo, error) {
	// First get the installed version
	info, err := a.GetInstalledVersion(ctx, pkg)
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

func (a *apt) UpdatePackageManager(ctx context.Context) error {
	// Update package lists
	_, err := a.runCommand(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Upgrade all packages
	_, err = a.runCommand(ctx, "upgrade", "--assume-yes")
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %w", err)
	}

	return nil
}

// checkIfInstalled overrides the base implementation with APT-specific logic
func (a *apt) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	// dpkg -s returns 0 if package is installed
	_, err := a.runCommand(ctx, "dpkg", "-s", pkg)
	if err == nil {
		return true, nil
	}

	// Check if the error is because the package is not installed
	output, _ := a.runCommand(ctx, "dpkg", "-l", pkg)
	if strings.Contains(output, pkg) && strings.Contains(output, "ii") {
		return true, nil
	}

	return false, nil
}
