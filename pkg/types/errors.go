package types

import "fmt"

// PackageAlreadyInstalledError is returned when a package is already installed
type PackageAlreadyInstalledError struct {
	Package string
}

func (e *PackageAlreadyInstalledError) Error() string {
	return fmt.Sprintf("package %s is already installed", e.Package)
}

// PackageNotFoundError is returned when a package is not found in the repository
type PackageNotFoundError struct {
	Package string
}

func (e *PackageNotFoundError) Error() string {
	return fmt.Sprintf("package %s not found in repository", e.Package)
}
