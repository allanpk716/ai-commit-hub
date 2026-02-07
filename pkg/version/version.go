package version

import (
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/mod/semver"
)

var (
	// Version 应用版本号，构建时通过 -ldflags 注入
	Version = "dev"
	// CommitSHA Git 提交 SHA，构建时通过 -ldflags 注入
	CommitSHA = "unknown"
	// BuildTime 构建时间，构建时通过 -ldflags 注入
	BuildTime = "unknown"
)

// 预编译正则表达式以提高性能
var (
	vPrefixRegex = regexp.MustCompile(`^v`)
	// 支持预发布版本的正则表达式: v1.2.3-beta.1 或 1.2.3-alpha.2
	versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
)

// ParseVersion 解析版本字符串，返回主版本、次版本、修订号和预发布标识符
// 支持格式："v1.2.3", "1.2.3", "v1.2.3-beta.1", "1.2.3-alpha.2"
// 格式错误时返回错误
func ParseVersion(version string) (major, minor, patch int, prerelease string, err error) {
	// 移除 v 前缀
	version = vPrefixRegex.ReplaceAllString(version, "")

	// 使用预编译的正则表达式匹配版本号格式
	matches := versionRegex.FindStringSubmatch(version)

	if matches == nil {
		return 0, 0, 0, "", fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, "", err
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, "", err
	}

	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, "", err
	}

	// 预发布标识符（如果有）
 prerelease = matches[4]

	return major, minor, patch, prerelease, nil
}

// NormalizeVersion 规范化版本号，确保以 "v" 开头（semver 库要求）
func NormalizeVersion(version string) string {
	// 移除空格
	version = regexp.MustCompile(`\s+`).ReplaceAllString(version, "")
	// 确保以 "v" 开头
	if !regexp.MustCompile(`^v`).MatchString(version) {
		return "v" + version
	}
	return version
}

// SafeCompareVersions 安全比较两个版本号，使用 semver 库
// 返回: 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
// 无效版本时返回 0
func SafeCompareVersions(v1, v2 string) int {
	normV1 := NormalizeVersion(v1)
	normV2 := NormalizeVersion(v2)

	// 验证版本号格式
	if !semver.IsValid(normV1) || !semver.IsValid(normV2) {
		return 0
	}

	return semver.Compare(normV1, normV2)
}

// IsPrerelease 检查是否为预发布版本
func IsPrerelease(version string) bool {
	normVersion := NormalizeVersion(version)
	if !semver.IsValid(normVersion) {
		return false
	}
	return semver.Prerelease(normVersion) != ""
}

// CompareVersions 比较两个版本号（兼容旧接口，内部使用 semver）
// 返回: 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
func CompareVersions(v1, v2 string) int {
	return SafeCompareVersions(v1, v2)
}

// GetVersion 获取当前版本号
// 开发模式返回 "dev-uncommitted"
// 生产模式返回 "v{major}.{minor}.{patch}"
func GetVersion() string {
	if Version == "dev" {
		return "dev-uncommitted"
	}
	return "v" + Version
}

// GetFullVersion 获取完整版本信息
// 格式: "v1.0.0 (abc1234 2024-01-15)"
// 开发模式返回 "dev-uncommitted"
func GetFullVersion() string {
	if Version == "dev" {
		return "dev-uncommitted"
	}
	return fmt.Sprintf("v%s (%s %s)", Version, CommitSHA, BuildTime)
}

// IsDevVersion 判断是否为开发版本
func IsDevVersion() bool {
	return Version == "dev"
}
