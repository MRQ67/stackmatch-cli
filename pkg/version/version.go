package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version (SemVer)
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

var (
	// versionRegex is a regular expression for parsing semantic versions
	// This is more permissive to handle partial versions like "1" or "1.2"
	versionRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)
)

// Parse parses a version string into a Version struct
func Parse(v string) (*Version, error) {
	matches := versionRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	ver := &Version{}
	var err error

	// Major version is required
	ver.Major, err = strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	// Minor version (default to 0 if not present)
	if matches[2] != "" {
		ver.Minor, err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %w", err)
		}
	}

	// Patch version (default to 0 if not present)
	if matches[3] != "" {
		ver.Patch, err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %w", err)
		}
	}

	// Pre-release and build metadata (optional)
	if len(matches) > 4 {
		ver.PreRelease = matches[4]
	}
	if len(matches) > 5 {
		ver.Build = matches[5]
	}

	return ver, nil
}

// Compare compares this version to another version.
// Returns -1 if v < other, 0 if v == other, or 1 if v > other.
func (v *Version) Compare(other *Version) int {
	// Compare major version
	if v.Major != other.Major {
		return compareInts(v.Major, other.Major)
	}
	// Compare minor version
	if v.Minor != other.Minor {
		return compareInts(v.Minor, other.Minor)
	}
	// Compare patch version
	if v.Patch != other.Patch {
		return compareInts(v.Patch, other.Patch)
	}
	// If one has a pre-release and the other doesn't, the one without is greater
	if v.PreRelease != "" && other.PreRelease == "" {
		return -1
	} else if v.PreRelease == "" && other.PreRelease != "" {
		return 1
	} else if v.PreRelease != other.PreRelease {
		// Compare pre-release versions lexically in ASCII sort order
		if v.PreRelease < other.PreRelease {
			return -1
		}
		return 1
	}
	// Build metadata is ignored in version comparisons
	return 0
}

// Satisfies checks if this version satisfies the given constraint
func (v *Version) Satisfies(constraint string) (bool, error) {
	// Handle empty constraint as "any version"
	if constraint == "" || constraint == "*" {
		return true, nil
	}

	// Handle basic operators: =, >, <, >=, <=
	for _, op := range []string{">=", "<=", ">", "<", "=", "!="} {
		if strings.HasPrefix(constraint, op) {
			verStr := strings.TrimSpace(constraint[len(op):])
			other, err := Parse(verStr)
			if err != nil {
				return false, fmt.Errorf("invalid version in constraint: %w", err)
			}

			cmp := v.Compare(other)
			switch op {
			case ">=":
				return cmp >= 0, nil
			case "<=":
				return cmp <= 0, nil
			case ">":
				return cmp > 0, nil
			case "<":
				return cmp < 0, nil
			case "=":
				return cmp == 0, nil
			case "!=":
				return cmp != 0, nil
			}
		}
	}

	// Handle version range (e.g., "1.2.3 - 2.3.4")
	if strings.Contains(constraint, " - ") {
		parts := strings.SplitN(constraint, " - ", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid version range: %s", constraint)
		}

		lower, err := Parse(strings.TrimSpace(parts[0]))
		if err != nil {
			return false, fmt.Errorf("invalid lower bound in range: %w", err)
		}

		upper, err := Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			return false, fmt.Errorf("invalid upper bound in range: %w", err)
		}

		return v.Compare(lower) >= 0 && v.Compare(upper) <= 0, nil
	}

	// Handle wildcards (e.g., "1.2.x" or "1.*")
	if strings.ContainsAny(constraint, "xX*^") {
		return checkWildcardConstraint(v, constraint)
	}

	// Handle exact match
	target, err := Parse(constraint)
	if err != nil {
		return false, fmt.Errorf("invalid version constraint: %w", err)
	}
	return v.Compare(target) == 0, nil
}

// checkWildcardConstraint handles version constraints with wildcards
func checkWildcardConstraint(v *Version, constraint string) (bool, error) {
	// Handle simple wildcards like * or x
	if constraint == "*" || constraint == "x" || constraint == "X" {
		return true, nil
	}

	// Handle patterns like 1.x or 1.2.x
	if strings.HasSuffix(constraint, ".x") || strings.HasSuffix(constraint, ".X") {
		prefix := strings.TrimSuffix(strings.TrimSuffix(constraint, ".x"), ".X")
		return strings.HasPrefix(v.String(), prefix+"."), nil
	}

	// Handle patterns like 1.2.3-*
	if strings.Contains(constraint, "-*") {
		prefix := strings.TrimSuffix(constraint, "-*")
		return strings.HasPrefix(v.String(), prefix), nil
	}

	// Handle other patterns with x/X/*
	replacer := strings.NewReplacer(
		"x", "[0-9]+",
		"X", "[0-9]+",
		"*", ".*",
	)
	regexStr := replacer.Replace(constraint)
	// Ensure we match the entire version string
	regexStr = "^" + regexStr + "$"

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return false, fmt.Errorf("invalid wildcard pattern: %w", err)
	}

	// Convert version to string and test against the regex
	versionStr := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		versionStr += "-" + v.PreRelease
	}

	return re.MatchString(versionStr), nil
}

// compareInts is a helper function to compare two integers
func compareInts(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// IsValid checks if a version string is valid according to semantic versioning
// It requires exactly major.minor.patch and handles v-prefixed versions
func IsValid(v string) bool {
	if v == "" {
		return false
	}

	// Handle v-prefixed versions
	if strings.HasPrefix(v, "v") {
		v = v[1:]
	}

	// Split into parts, but don't include pre-release/build metadata in the count
	baseVersion := v
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		baseVersion = v[:idx]
	}

	// Must have exactly major.minor.patch
	parts := strings.Split(baseVersion, ".")
	if len(parts) != 3 {
		return false
	}

	// Check major, minor, patch are all numbers
	for i := 0; i < 3; i++ {
		numberPart := parts[i]
		if numberPart == "" {
			return false
		}

		// Allow wildcards for validation
		if numberPart == "x" || numberPart == "X" || numberPart == "*" {
			continue
		}

		if _, err := strconv.Atoi(numberPart); err != nil {
			return false
		}
	}

	return true
}

// String returns the string representation of the version
func (v *Version) String() string {
	versionStr := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		versionStr += "-" + v.PreRelease
	}
	if v.Build != "" {
		versionStr += "+" + v.Build
	}
	return versionStr
}
