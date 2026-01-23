package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogAssertions_AssertLogContains(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("test message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogContains("test message"))
}

func TestLogAssertions_AssertLogNotContains(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("test message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogNotContains("not found"))
}

func TestLogAssertions_AssertNoErrors(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("info message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertNoErrors())

	capture.Error("error message")
	assert.False(t, assertions.AssertNoErrors())
}

func TestLogAssertions_AssertLogCount(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("test")
	capture.Info("test")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogCount("test", 2))
}

func TestLogAssertions_AssertLogPattern(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("[INFO] test message")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogPattern(`\[INFO\].*test`))
}

func TestLogAssertions_AssertLogSequence(t *testing.T) {
	capture := NewLogCapture(t)
	capture.Info("first")
	capture.Info("second")
	capture.Info("third")

	assertions := NewLogAssertions(t, capture)
	assert.True(t, assertions.AssertLogSequence("first", "second", "third"))
}
