package pushover

import (
	"testing"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Git describe format with commit info",
			version:  "v1.6.0-1-g3871faa",
			expected: "v1.6.0",
		},
		{
			name:     "Git describe format with multiple commits",
			version:  "v2.0.0-15-gabc123",
			expected: "v2.0.0",
		},
		{
			name:     "Clean version without v prefix",
			version:  "1.6.0-1-g3871faa",
			expected: "1.6.0",
		},
		{
			name:     "Pure version tag",
			version:  "v1.6.0",
			expected: "v1.6.0",
		},
		{
			name:     "Prerelease version should not be modified",
			version:  "v1.6.0-alpha",
			expected: "v1.6.0-alpha",
		},
		{
			name:     "Beta version should not be modified",
			version:  "v1.6.0-beta.1",
			expected: "v1.6.0-beta.1",
		},
		{
			name:     "Empty string",
			version:  "",
			expected: "",
		},
		{
			name:     "Version without patch",
			version:  "v1.6-1-g3871faa",
			expected: "v1.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanVersion(tt.version)
			if result != tt.expected {
				t.Errorf("cleanVersion(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}


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

func TestCleanVersionAndCompare(t *testing.T) {
	tests := []struct {
		name               string
		hookVersion        string
		extensionVersion   string
		needsUpdate        bool
	}{
		{
			name:             "Hook version matches extension version (Git describe format)",
			hookVersion:      "v1.6.0-1-g3871faa",
			extensionVersion: "v1.6.0",
			needsUpdate:      false,
		},
		{
			name:             "Hook version is older than extension version",
			hookVersion:      "v1.5.0-3-gabc123",
			extensionVersion: "v1.6.0",
			needsUpdate:      true,
		},
		{
			name:             "Hook version is newer than extension version",
			hookVersion:      "v1.7.0-1-gdef456",
			extensionVersion: "v1.6.0",
			needsUpdate:      false,
		},
		{
			name:             "Both versions are clean tags",
			hookVersion:      "v1.6.0",
			extensionVersion: "v1.6.0",
			needsUpdate:      false,
		},
		{
			name:             "Hook with prerelease, extension with release",
			hookVersion:      "v1.6.0-alpha",
			extensionVersion: "v1.6.0",
			needsUpdate:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理 Hook 版本
			cleanedHookVersion := cleanVersion(tt.hookVersion)

			// 比较版本
			result := CompareVersions(cleanedHookVersion, tt.extensionVersion)
			needsUpdate := result < 0

			if needsUpdate != tt.needsUpdate {
				t.Errorf("cleanVersion(%q) = %q, CompareVersions(%q, %q) = %d, needsUpdate=%v, want %v",
					tt.hookVersion, cleanedHookVersion,
					cleanedHookVersion, tt.extensionVersion,
					result, needsUpdate, tt.needsUpdate)
			}
		})
	}
}

