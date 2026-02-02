package version

import (
	"fmt"
	"regexp"
	"strconv"
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
	versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
)

// ParseVersion 解析版本字符串，返回主版本、次版本、修订号
// 支持格式："v1.2.3" 或 "1.2.3"
// 注意：不支持预发布版本号（如 v1.2.3-beta）或构建元数据（如 v1.2.3+build123）
// 格式错误时返回错误
func ParseVersion(version string) (major, minor, patch int, err error) {
	// 移除 v 前缀
	version = vPrefixRegex.ReplaceAllString(version, "")

	// 使用预编译的正则表达式匹配版本号格式
	matches := versionRegex.FindStringSubmatch(version)

	if matches == nil {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, err
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, err
	}

	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, err
	}

	return major, minor, patch, nil
}
