package installer

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/MRQ67/stackmatch-cli/pkg/installer/package_managers"
	"github.com/MRQ67/stackmatch-cli/pkg/types"
	"github.com/MRQ67/stackmatch-cli/pkg/ui"
)

// Type aliases for cleaner code
type (
	// Installer is an alias for types.Installer for backward compatibility
	Installer = types.Installer
	// VersionConstraint is an alias for types.VersionConstraint
	VersionConstraint = types.VersionConstraint
	// PackageVersionInfo is an alias for types.PackageVersionInfo
	PackageVersionInfo = types.PackageVersionInfo
)

// DetectPackageManager detects the best available package manager for the current system
func DetectPackageManager() (Installer, error) {
	// Check package managers in order of preference based on OS
	var managers []Installer

	switch runtime.GOOS {
	case "windows":
		managers = []Installer{
			package_managers.NewChocolatey(),
			package_managers.NewScoop(),
			package_managers.NewWinget(),
		}
	case "darwin":
		managers = []Installer{
			package_managers.NewHomebrew(),
		}
	default: // Linux and others
		managers = []Installer{
			package_managers.NewApt(),
			package_managers.NewDnf(),
			package_managers.NewYum(),
			package_managers.NewPacman(),
			package_managers.NewSnap(),
		}
	}

	// Return the first available package manager
	for _, mgr := range managers {
		if mgr.IsAvailable() {
			return mgr, nil
		}
	}

	return nil, fmt.Errorf("no supported package manager found")
}

// installWithMapping installs a package using the appropriate package name for the installer
func installWithMapping(ctx context.Context, installerInst Installer, pkg string, version ...VersionConstraint) error {
	// Get the package name for this specific package manager
	pmType := installerInst.Type()
	mappedPkg, err := GetPackageName(pkg, pmType)
	if err != nil {
		return fmt.Errorf("package mapping error: %w", err)
	}

	// If we get an empty package name, it means no mapping was found
	// and we should use the original package name
	if mappedPkg == "" {
		mappedPkg = pkg
	}

	// Check if we have a version constraint
	if len(version) > 0 && version[0].Version != "" {
		// First check if the installed version already satisfies the constraint
		info, err := installerInst.CheckVersion(ctx, mappedPkg, version[0])
		if err == nil && info != nil && info.Satisfies {
			// Already installed with a compatible version
			return nil
		}

		// Install specific version
		err = installerInst.InstallVersion(ctx, mappedPkg, version[0])
	} else {
		// Install without version constraint
		err = installerInst.InstallPackage(ctx, mappedPkg)
	}

	if err != nil {
		// If we get a PackageNotFoundError, try with the original package name
		if _, ok := err.(*types.PackageNotFoundError); ok && mappedPkg != pkg {
			if len(version) > 0 && version[0].Version != "" {
				err = installerInst.InstallVersion(ctx, pkg, version[0])
			} else {
				err = installerInst.InstallPackage(ctx, pkg)
			}
		}
		return err
	}

	return nil
}

// InstallPackage installs a package using the best available package manager
func InstallPackage(ctx context.Context, pkg string, version ...VersionConstraint) error {
	installerInst, err := DetectPackageManager()
	if err != nil {
		return err
	}

	// Show confirmation
	versionStr := ""
	if len(version) > 0 && version[0].Version != "" {
		versionStr = " (version: " + version[0].Version + ")"
	}

	ui.PrintInfo("Package manager: %s", installerInst.Name())
	confirmed, err := ui.Confirm(fmt.Sprintf("Install package %s%s?", pkg, versionStr), true)
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}
	if !confirmed {
		return fmt.Errorf("installation cancelled by user")
	}

	// Show progress
	spinner := ui.NewSpinner(fmt.Sprintf("Installing %s...", pkg))
	defer spinner.Close()

	var result error
	if len(version) > 0 {
		result = installWithMapping(ctx, installerInst, pkg, version[0])
	} else {
		result = installWithMapping(ctx, installerInst, pkg)
	}

	if result != nil {
		ui.PrintError(result, "Failed to install %s", pkg)
	} else {
		ui.PrintSuccess("Successfully installed %s", pkg)
		
		// Verify installation
		if len(version) > 0 {
			if err := VerifyInstallation(ctx, installerInst, pkg, &version[0]); err != nil {
				ui.PrintWarning("Verification warning: %v", err)
			}
		} else {
			if err := VerifyInstallation(ctx, installerInst, pkg, nil); err != nil {
				ui.PrintWarning("Verification warning: %v", err)
			}
		}
	}

	return result
}

// InstallPackages installs multiple packages using the best available package manager
func InstallPackages(ctx context.Context, packages []string, versions ...map[string]VersionConstraint) error {
	if len(packages) == 0 && (len(versions) == 0 || len(versions[0]) == 0) {
		return fmt.Errorf("no packages to install")
	}

	installerInst, err := DetectPackageManager()
	if err != nil {
		return err
	}

	// Show summary of packages to install
	ui.PrintInfo("Package manager: %s", installerInst.Name())
	ui.PrintInfo("Packages to install:")
	for _, pkg := range packages {
		ui.PrintInfo("  - %s", pkg)
	}

	versionedPkgs := make(map[string]types.VersionConstraint)
	if len(versions) > 0 {
		for pkg, ver := range versions[0] {
			versionedPkgs[pkg] = ver
			ui.PrintInfo("  - %s (version: %s)", pkg, ver.Version)
		}
	}

	confirmed, err := ui.Confirm("Proceed with installation?", true)
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}
	if !confirmed {
		return fmt.Errorf("installation cancelled by user")
	}

	// Use batchInstall for better progress reporting and verification
	return batchInstall(ctx, installerInst, packages, versionedPkgs)
}

// batchInstall installs multiple packages with progress reporting
func batchInstall(ctx context.Context, installerInst Installer, packages []string, versions map[string]types.VersionConstraint) error {
	// Show progress
	spinner := ui.NewSpinner("Installing packages...")
	defer spinner.Close()

	var failed []string
	// Process regular packages
	for _, pkg := range packages {
		err := installWithMapping(ctx, installerInst, pkg)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", pkg, err))
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to install packages: %s", strings.Join(failed, "; "))
	}

	return nil
}
