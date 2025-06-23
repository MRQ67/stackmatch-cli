package package_managers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/version"
)

type homebrew struct {
	*basePackageManager
}

// NewHomebrew creates a new Homebrew package manager instance
func NewHomebrew() types.Installer {
	hb := &homebrew{
		basePackageManager: &basePackageManager{
			name:            "Homebrew",
			pmType:          types.TypeHomebrew,
			executableName:   "brew",
			versionCommand:   "info --json=v2",
			versionRegex:     `"version":"([^"]+)"`,
			installWithFlags: true,
		},
	}
	hb.installPackageFunc = hb.installPackage
	hb.installMultipleFunc = hb.installMultiple
	hb.uninstallPackageFunc = hb.uninstallPackage
	return hb
}

// installPackage installs a single package
func (h *homebrew) installPackage(ctx context.Context, pkg string) error {
	// First check if already installed
	installed, err := h.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if installed {
		return &types.PackageAlreadyInstalledError{Package: pkg}
	}

	// Install the package
	_, err = h.runCommand(ctx, "install", pkg)
	if err != nil {
		// Check if package was not found
		if strings.Contains(err.Error(), "No available formula or cask") {
			return &types.PackageNotFoundError{
				Package: pkg,
			}
		}
		return fmt.Errorf("failed to install package: %w", err)
	}

	return nil
}

// uninstallPackage uninstalls a package using Homebrew
func (h *homebrew) uninstallPackage(ctx context.Context, pkg string) error {
	// First check if installed
	installed, err := h.checkIfInstalled(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if !installed {
		return nil // Already uninstalled
	}

	// Uninstall the package
	_, err = h.runCommand(ctx, "uninstall", "--ignore-dependencies", pkg)
	if err != nil {
		return fmt.Errorf("failed to uninstall package: %w", err)
	}

	return nil
}

// InstallPackage implements the Installer interface
func (h *homebrew) InstallPackage(ctx context.Context, pkg string) error {
	return h.installPackage(ctx, pkg)
}

// InstallVersion installs a specific version of a package
func (h *homebrew) InstallVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) error {
	// Check if the package is already installed with the required version
	info, err := h.CheckVersion(ctx, pkg, constraint)
	if err != nil {
		return fmt.Errorf("failed to check package version: %w", err)
	}

	if info.Satisfies {
		return nil // Already installed with the required version
	}

	// Get available versions
	versions, err := h.getAvailableVersions(ctx, pkg)
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
	_, err = h.runCommand(ctx, "install", fmt.Sprintf("%s@%s", pkg, selectedVersion))
	if err != nil {
		return fmt.Errorf("failed to install package version %s: %w", selectedVersion, err)
	}

	return nil
}

// installMultiple installs multiple packages in a single operation
func (h *homebrew) installMultiple(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Homebrew can install multiple packages in one command
	args := append([]string{"install"}, packages...)
	_, err := h.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

// InstallMultiple implements the Installer interface
func (h *homebrew) InstallMultiple(ctx context.Context, packages []string) error {
	return h.installMultiple(ctx, packages)
}

// InstallMultipleVersions installs multiple packages with specific versions
func (h *homebrew) InstallMultipleVersions(ctx context.Context, packages map[string]types.VersionConstraint) error {
	if len(packages) == 0 {
		return nil
	}

	// Install each package with its version constraint
	for pkg, constraint := range packages {
		if constraint.Version != "" {
			if err := h.InstallVersion(ctx, pkg, constraint); err != nil {
				return fmt.Errorf("failed to install %s@%s: %w", pkg, constraint.Version, err)
			}
		} else {
			if err := h.InstallPackage(ctx, pkg); err != nil {
				return fmt.Errorf("failed to install %s: %w", pkg, err)
			}
		}
	}

	return nil
}

// getAvailableVersions gets all available versions for a package
func (h *homebrew) getAvailableVersions(ctx context.Context, pkg string) ([]string, error) {
	output, err := h.runCommand(ctx, "info", "--json=v1", "--installed", pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to get package info: %w", err)
	}

	// Parse the JSON output to extract versions
	// This is a simplified version - in a real implementation, you'd want to properly parse the JSON
	re := regexp.MustCompile(`"version":"([^"]+)"`)
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
func (h *homebrew) GetInstalledVersion(ctx context.Context, pkg string) (*types.PackageVersionInfo, error) {
	output, err := h.runCommand(ctx, "list", "--versions", pkg)
	if err != nil {
		// If the package is not installed, return empty version
		if strings.Contains(err.Error(), "No available formula or cask") ||
			strings.Contains(err.Error(), "No such keg") {
			return &types.PackageVersionInfo{
				Name: pkg,
			}, nil
		}
		return nil, fmt.Errorf("failed to get installed version: %w", err)
	}

	// Parse the output which is in format: "pkg version1 version2 ..."
	parts := strings.Fields(output)
	if len(parts) < 2 {
		return &types.PackageVersionInfo{
			Name: pkg,
		}, nil
	}

	// Return the first version (most recent)
	return &types.PackageVersionInfo{
		Name:    pkg,
		Version: parts[1],
	}, nil
}

// CheckVersion checks if the installed package satisfies the version constraint
func (h *homebrew) CheckVersion(ctx context.Context, pkg string, constraint types.VersionConstraint) (*types.PackageVersionInfo, error) {
	// First get the installed version
	info, err := h.GetInstalledVersion(ctx, pkg)
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

func (h *homebrew) UpdatePackageManager(ctx context.Context) error {
	_, err := h.runCommand(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to update Homebrew: %w", err)
	}
	return nil
}

// checkIfInstalled overrides the base implementation with Homebrew-specific logic
func (h *homebrew) checkIfInstalled(ctx context.Context, pkg string) (bool, error) {
	output, err := h.runCommand(ctx, "list", "--versions", pkg)
	if err != nil {
		// If the package is not installed, list will return an error
		if strings.Contains(err.Error(), "No available formula or cask") ||
			strings.Contains(err.Error(), "No such keg") {
			return false, nil
		}
		return false, err
	}

	// If we get output, the package is installed
	return strings.TrimSpace(output) != "", nil
}
