package installer

import (
	"context"
	"fmt"

	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/ui"
)

// VerifyInstallation verifies that a package was installed correctly
func VerifyInstallation(ctx context.Context, pkgMgr Installer, pkg string, versionConstraint *types.VersionConstraint) error {
	ui.PrintInfo("Verifying installation of %s...", pkg)

	// Check if package is installed
	info, err := pkgMgr.GetInstalledVersion(ctx, pkg)
	if err != nil {
		return fmt.Errorf("failed to verify installation of %s: %w", pkg, err)
	}

	if info.Version == "" {
		return fmt.Errorf("package %s is not installed", pkg)
	}

	// If version constraint is provided, verify it
	if versionConstraint != nil && versionConstraint.Version != "" {
		versionInfo, err := pkgMgr.CheckVersion(ctx, pkg, *versionConstraint)
		if err != nil {
			return fmt.Errorf("failed to check version for %s: %w", pkg, err)
		}

		if !versionInfo.Satisfies {
			return fmt.Errorf("version check failed for %s: installed %s does not satisfy %s", 
				pkg, info.Version, versionConstraint.Version)
		}

		ui.PrintSuccess("Verified %s version %s (required: %s)", pkg, info.Version, versionConstraint.Version)
	} else {
		ui.PrintSuccess("Verified %s is installed (version: %s)", pkg, info.Version)
	}

	return nil
}

// BatchInstall performs batch installation with progress reporting
func BatchInstall(ctx context.Context, pkgMgr Installer, packages []string, versionedPackages map[string]types.VersionConstraint) error {
	total := len(packages) + len(versionedPackages)
	if total == 0 {
		ui.PrintInfo("No packages to install")
		return nil
	}

	bar := ui.NewProgressBar(total, "Installing packages")
	defer bar.Close()
	bar.Render()

	// Install regular packages
	for _, pkg := range packages {
		if err := bar.Add(1); err != nil {
			ui.PrintWarning("Failed to update progress bar: %v", err)
		}

		ui.PrintInfo("Installing %s...", pkg)
		if err := pkgMgr.InstallPackage(ctx, pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}

		// Verify installation
		if err := VerifyInstallation(ctx, pkgMgr, pkg, nil); err != nil {
			return fmt.Errorf("verification failed for %s: %w", pkg, err)
		}
	}

	// Install versioned packages
	for pkg, constraint := range versionedPackages {
		if err := bar.Add(1); err != nil {
			ui.PrintWarning("Failed to update progress bar: %v", err)
		}

		ui.PrintInfo("Installing %s@%s...", pkg, constraint.Version)
		if err := pkgMgr.InstallVersion(ctx, pkg, constraint); err != nil {
			return fmt.Errorf("failed to install %s@%s: %w", pkg, constraint.Version, err)
		}

		// Verify installation with version constraint
		if err := VerifyInstallation(ctx, pkgMgr, pkg, &constraint); err != nil {
			return fmt.Errorf("verification failed for %s@%s: %w", pkg, constraint.Version, err)
		}
	}

	return nil
}
