package version

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"1.2.3", "1.2.3", false},
		{"v1.2.3", "1.2.3", false},
		{"1.2.3-alpha", "1.2.3-alpha", false},
		{"1.2.3-alpha.1", "1.2.3-alpha.1", false},
		{"1.2.3+build", "1.2.3+build", false},
		{"1.2.3-alpha.1+build.123", "1.2.3-alpha.1+build.123", false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			v, err := Parse(tc.input)
			if tc.hasError {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}

			got := v.String()
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.3.0", -1},
		{"1.3.0", "1.2.3", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.beta", "1.0.0-alpha.1", 1}, // beta > 1 in ASCII
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_%s", tc.a, tc.b), func(t *testing.T) {
			va, err := Parse(tc.a)
			if err != nil {
				t.Fatalf("failed to parse version %q: %v", tc.a, err)
			}

			vb, err := Parse(tc.b)
			if err != nil {
				t.Fatalf("failed to parse version %q: %v", tc.b, err)
			}


			got := va.Compare(vb)
			if got != tc.expected {
				t.Errorf("compare(%q, %q): expected %d, got %d", tc.a, tc.b, tc.expected, got)
			}
		})
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
		hasError   bool
	}{
		// Exact version
		{"1.2.3", "1.2.3", true, false},
		{"1.2.3", "=1.2.3", true, false},
		{"1.2.3", "1.2.4", false, false},

		// Comparison operators
		{"1.2.3", ">1.2.2", true, false},
		{"1.2.3", ">1.2.3", false, false},
		{"1.2.3", ">=1.2.3", true, false},
		{"1.2.3", "<1.2.4", true, false},
		{"1.2.3", "<1.2.3", false, false},
		{"1.2.3", "<=1.2.3", true, false},
		{"1.2.3", "!=1.2.4", true, false},
		{"1.2.3", "!=1.2.3", false, false},

		// Ranges
		{"1.2.3", "1.2.0 - 1.3.0", true, false},
		{"1.2.3", "1.0.0 - 1.2.2", false, false},
		{"1.2.3", "1.2.3 - 1.2.3", true, false},

		// Wildcards
		{"1.2.3", "1.2.x", true, false},
		{"1.2.3", "1.x", true, false},
		{"1.2.3", "1.2.3-*", true, false},
		{"1.2.3", "1.2.4-*", false, false},

		// Invalid constraints
		{"1.2.3", "invalid", false, true},
		{"1.2.3", "1.2.3.4", false, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_%s", tc.version, tc.constraint), func(t *testing.T) {
			v, err := Parse(tc.version)
			if err != nil {
				t.Fatalf("failed to parse version %q: %v", tc.version, err)
			}

			result, err := v.Satisfies(tc.constraint)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for constraint %q, got nil", tc.constraint)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for constraint %q: %v", tc.constraint, err)
			}

			if result != tc.expected {
				t.Errorf("satisfies(%q, %q): expected %v, got %v", tc.version, tc.constraint, tc.expected, result)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"1.2.3", true},
		{"v1.2.3", true},
		{"1.2.3-alpha", true},
		{"1.2.3-alpha.1", true},
		{"1.2.3+build", true},
		{"1.2.3-alpha.1+build.123", true},
		{"1", false},
		{"1.2", false},
		{"1.2.3.4", false},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			got := IsValid(tc.version)
			if got != tc.expected {
				t.Errorf("IsValid(%q): expected %v, got %v", tc.version, tc.expected, got)
			}
		})
	}
}
