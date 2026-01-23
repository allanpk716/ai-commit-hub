package helpers

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// LogAssertions 日志断言
type LogAssertions struct {
	t       *testing.T
	capture *LogCapture
}

// NewLogAssertions 创建日志断言器
func NewLogAssertions(t *testing.T, capture *LogCapture) *LogAssertions {
	t.Helper()
	return &LogAssertions{
		t:       t,
		capture: capture,
	}
}

// AssertLogContains 断言日志包含文本
func (a *LogAssertions) AssertLogContains(substring string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	if !a.capture.Contains(substring) {
		assert.Fail(a.t, "日志不包含预期内容", substring, msgAndArgs)
		return false
	}
	return true
}

// AssertLogNotContains 断言日志不包含文本
func (a *LogAssertions) AssertLogNotContains(substring string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	if a.capture.Contains(substring) {
		assert.Fail(a.t, "日志包含不应出现的内容", substring, msgAndArgs)
		return false
	}
	return true
}

// AssertLogCount 断言日志出现次数
func (a *LogAssertions) AssertLogCount(substring string, expectedCount int, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	actualCount := strings.Count(logs, substring)

	if actualCount != expectedCount {
		assert.Fail(a.t,
			"日志出现次数不符",
			substring, "期望", expectedCount, "次，实际", actualCount, "次",
			msgAndArgs,
		)
		return false
	}
	return true
}

// AssertLogPattern 断言日志匹配正则表达式
func (a *LogAssertions) AssertLogPattern(pattern string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	matched, err := regexp.MatchString(pattern, logs)

	if err != nil || !matched {
		assert.Fail(a.t, "日志不匹配正则表达式", pattern, msgAndArgs)
		return false
	}
	return true
}

// AssertNoErrors 断言没有错误或警告日志
func (a *LogAssertions) AssertNoErrors(msgAndArgs ...interface{}) bool {
	a.t.Helper()
	return !a.capture.ContainsError()
}

// AssertLogSequence 断言日志按顺序出现
func (a *LogAssertions) AssertLogSequence(substrings ...string) bool {
	a.t.Helper()
	logs := a.capture.GetLogs()
	lastIndex := -1

	for _, substr := range substrings {
		index := strings.Index(logs, substr)
		if index == -1 || index <= lastIndex {
			assert.Fail(a.t,
				"日志顺序不正确",
				"期望按顺序出现:", substrings,
			)
			return false
		}
		lastIndex = index
	}
	return true
}
