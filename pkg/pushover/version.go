package pushover

import (
	"strconv"
	"strings"
)

// CompareVersions 比较两个版本字符串
// 返回值: -1 表示 v1 < v2, 0 表示 v1 == v2, 1 表示 v1 > v2
// 支持语义化版本格式 (例如: "1.2.3", "2.0.0-alpha")
func CompareVersions(v1, v2 string) int {
	// 处理 unknown 或空字符串
	if v1 == "" || v1 == "unknown" {
		if v2 == "" || v2 == "unknown" {
			return 0
		}
		return -1
	}
	if v2 == "" || v2 == "unknown" {
		return 1
	}

	// 分割主版本号和预发布标签
	v1Parts := strings.SplitN(v1, "-", 2)
	v2Parts := strings.SplitN(v2, "-", 2)

	v1Main := v1Parts[0]
	v2Main := v2Parts[0]

	// 比较主版本号
	result := compareVersionParts(v1Main, v2Main)
	if result != 0 {
		return result
	}

	// 主版本号相同，比较预发布标签
	// 没有预发布标签的版本更高
	if len(v1Parts) == 1 && len(v2Parts) == 1 {
		return 0
	}
	if len(v1Parts) == 1 {
		return 1
	}
	if len(v2Parts) == 1 {
		return -1
	}

	// 比较预发布标签
	v1Pre := v1Parts[1]
	v2Pre := v2Parts[1]

	// 简单的字符串比较（可以后续优化为更复杂的预发布版本比较）
	if v1Pre < v2Pre {
		return -1
	}
	if v1Pre > v2Pre {
		return 1
	}
	return 0
}

// compareVersionParts 比较版本号的主部分 (例如 "1.2.3")
func compareVersionParts(v1, v2 string) int {
	v1Numbers := parseVersionNumbers(v1)
	v2Numbers := parseVersionNumbers(v2)

	maxLen := len(v1Numbers)
	if len(v2Numbers) > maxLen {
		maxLen = len(v2Numbers)
	}

	for i := 0; i < maxLen; i++ {
		v1Num := 0
		v2Num := 0

		if i < len(v1Numbers) {
			v1Num = v1Numbers[i]
		}
		if i < len(v2Numbers) {
			v2Num = v2Numbers[i]
		}

		if v1Num < v2Num {
			return -1
		}
		if v1Num > v2Num {
			return 1
		}
	}

	return 0
}

// parseVersionNumbers 将版本字符串解析为数字数组
func parseVersionNumbers(version string) []int {
	parts := strings.Split(version, ".")
	numbers := make([]int, 0, len(parts))

	for _, part := range parts {
		// 移除可能的非数字前缀/后缀
		part = strings.TrimSpace(part)
		if num, err := strconv.Atoi(part); err == nil {
			numbers = append(numbers, num)
		}
	}

	return numbers
}
