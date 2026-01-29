package pushover

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "Equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "v1 less than v2",
			v1:       "1.0.0",
			v2:       "1.0.1",
			expected: -1,
		},
		{
			name:     "v1 greater than v2",
			v1:       "1.0.1",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2 (major version)",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 (minor version)",
			v1:       "1.2.0",
			v2:       "1.1.0",
			expected: 1,
		},
		{
			name:     "v1 with prerelease",
			v1:       "1.0.0-alpha",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "v1 unknown",
			v1:       "unknown",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "both unknown",
			v1:       "unknown",
			v2:       "unknown",
			expected: 0,
		},
		{
			name:     "v1 empty",
			v1:       "",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "Complex version comparison",
			v1:       "1.2.3",
			v2:       "1.2.4",
			expected: -1,
		},
		{
			name:     "Different length versions",
			v1:       "1.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "Both are unknown",
			v1:       "unknown",
			v2:       "unknown",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseVersionNumbers(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []int
	}{
		{
			name:     "Simple version",
			version:  "1.2.3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "Two part version",
			version:  "1.2",
			expected: []int{1, 2},
		},
		{
			name:     "Single part version",
			version:  "1",
			expected: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionNumbers(tt.version)
			if len(result) != len(tt.expected) {
				t.Errorf("parseVersionNumbers(%q) length = %d, want %d", tt.version, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseVersionNumbers(%q)[%d] = %d, want %d", tt.version, i, result[i], tt.expected[i])
				}
			}
		})
	}
}
