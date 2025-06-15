package scanner

import (
	"regexp"
	"testing"
)

func TestParseVersion(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		regex    *regexp.Regexp
		expected string
	}{
		{
			name:     "Go Version",
			output:   "go version go1.18.3 windows/amd64",
			regex:    regexp.MustCompile(`go version go([\d\.]+)`),
			expected: "1.18.3",
		},
		{
			name:     "Node.js Version",
			output:   "v16.15.0",
			regex:    regexp.MustCompile(`v?([\d\.]+)`),
			expected: "16.15.0",
		},
		{
			name:     "Python Version",
			output:   "Python 3.10.4",
			regex:    regexp.MustCompile(`Python ([\d\.]+)`),
			expected: "3.10.4",
		},
		{
			name:     "Git Version",
			output:   "git version 2.36.1.windows.1",
			regex:    regexp.MustCompile(`git version (\d+(?:\.\d+)*)`),
			expected: "2.36.1",
		},
		{
			name:     "Homebrew Version",
			output:   "Homebrew 3.5.2",
			regex:    regexp.MustCompile(`Homebrew ([\d\.]+)`),
			expected: "3.5.2",
		},
		{
			name:     "No Match",
			output:   "Some random string",
			regex:    regexp.MustCompile(`version ([\d\.]+)`),
			expected: "",
		},
		{
			name:     "Nil Regex",
			output:   "Some random string",
			regex:    nil,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := parseVersion(tc.output, tc.regex)
			if actual != tc.expected {
				t.Errorf("expected version '%s', but got '%s'", tc.expected, actual)
			}
		})
	}
}
