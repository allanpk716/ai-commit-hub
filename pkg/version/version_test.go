package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{
			name:      "标准版本号带v前缀",
			version:   "v1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "标准版本号不带v前缀",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "空字符串",
			version:   "",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   true,
		},
		{
			name:      "格式错误",
			version:   "invalid",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, _, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
				t.Errorf("ParseVersion() = %v.%v.%v, want %v.%v.%v",
					major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
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
		{"v1大于v2_修订版本不同", "v1.2.3", "v1.2.2", 1},
		{"v1小于v2_修订版本不同", "v1.2.2", "v1.2.3", -1},
		{"v1等于v2", "v1.2.3", "v1.2.3", 0},
		{"主版本不同_v1大于v2", "v2.0.0", "v1.9.9", 1},
		{"主版本不同_v1小于v2", "v1.9.9", "v2.0.0", -1},
		{"次版本不同_v1大于v2", "v1.3.0", "v1.2.9", 1},
		{"次版本不同_v1小于v2", "v1.2.9", "v1.3.0", -1},
		{"修订版本不同_v1大于v2", "v1.2.4", "v1.2.3", 1},
		{"修订版本不同_v1小于v2", "v1.2.3", "v1.2.4", -1},
		{"不带v前缀_v1大于v2", "1.2.3", "1.2.2", 1},
		{"混合前缀_v1带v前缀_v2不带", "v1.2.3", "1.2.2", 1},
		{"混合前缀_v1不带v前缀_v2带", "1.2.3", "v1.2.2", 1},
		{"v1格式错误_视为相等", "invalid", "v1.2.3", 0},
		{"v2格式错误_视为相等", "v1.2.3", "invalid", 0},
		{"两者都格式错误_视为相等", "invalid1", "invalid2", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := CompareVersions(tt.v1, tt.v2); result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseVersionWithPrerelease(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		wantMajor     int
		wantMinor     int
		wantPatch     int
		wantPrerelease string
		wantErr       bool
	}{
		{
			name:          "预发布版本beta",
			version:       "v1.2.3-beta.1",
			wantMajor:     1,
			wantMinor:     2,
			wantPatch:     3,
			wantPrerelease: "beta.1",
			wantErr:       false,
		},
		{
			name:          "预发布版本alpha",
			version:       "1.2.3-alpha.2",
			wantMajor:     1,
			wantMinor:     2,
			wantPatch:     3,
			wantPrerelease: "alpha.2",
			wantErr:       false,
		},
		{
			name:          "预发布版本rc",
			version:       "v2.0.0-rc.1",
			wantMajor:     2,
			wantMinor:     0,
			wantPatch:     0,
			wantPrerelease: "rc.1",
			wantErr:       false,
		},
		{
			name:          "稳定版本无预发布标识",
			version:       "v1.2.3",
			wantMajor:     1,
			wantMinor:     2,
			wantPatch:     3,
			wantPrerelease: "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, prerelease, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch || prerelease != tt.wantPrerelease {
				t.Errorf("ParseVersion() = %v.%v.%v-%v, want %v.%v.%v-%v",
					major, minor, patch, prerelease, tt.wantMajor, tt.wantMinor, tt.wantPatch, tt.wantPrerelease)
			}
		})
	}
}

func TestComparePrereleaseVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"beta大于alpha", "v1.0.0-beta.1", "v1.0.0-alpha.2", 1},
		{"alpha.2大于alpha.1", "v1.0.0-alpha.2", "v1.0.0-alpha.1", 1},
		{"稳定版大于预发布版", "v1.0.0", "v1.0.0-beta.1", 1},
		{"预发布版小于稳定版", "v1.0.0-beta.1", "v1.0.0", -1},
		{"相同预发布版本", "v1.0.0-beta.1", "v1.0.0-beta.1", 0},
		{"rc大于beta", "v1.0.0-rc.1", "v1.0.0-beta.2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := CompareVersions(tt.v1, tt.v2); result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"beta版本", "v1.0.0-beta.1", true},
		{"alpha版本", "v1.0.0-alpha.2", true},
		{"rc版本", "v1.0.0-rc.1", true},
		{"稳定版本", "v1.0.0", false},
		{"不带v前缀的预发布版", "1.0.0-beta.1", true},
		{"不带v前缀的稳定版", "1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsPrerelease(tt.version); result != tt.expected {
				t.Errorf("IsPrerelease(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"已带v前缀", "v1.0.0", "v1.0.0"},
		{"不带v前缀", "1.0.0", "v1.0.0"},
		{"带空格", " 1.0.0 ", "v1.0.0"},
		{"预发布版本", "1.0.0-beta.1", "v1.0.0-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := NormalizeVersion(tt.version); result != tt.expected {
				t.Errorf("NormalizeVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	// 保存原始值
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"开发版本", "dev", "dev-uncommitted"},
		{"生产版本", "1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := GetVersion()
			if result != tt.expected {
				t.Errorf("GetVersion() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestIsDevVersion(t *testing.T) {
	// 保存原始值
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"开发版本", "dev", true},
		{"生产版本", "1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := IsDevVersion()
			if result != tt.expected {
				t.Errorf("IsDevVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}
