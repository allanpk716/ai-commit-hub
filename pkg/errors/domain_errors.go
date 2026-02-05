package errors

import "fmt"

// ValidationError 表示数据验证错误
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// GitOperationError 表示 Git 操作错误
type GitOperationError struct {
	Operation string
	Path      string
	Err       error
}

func (e *GitOperationError) Error() string {
	return fmt.Sprintf("git operation '%s' failed at %s: %v", e.Operation, e.Path, e.Err)
}

func (e *GitOperationError) Unwrap() error {
	return e.Err
}

// NewGitOperationError 创建 Git 操作错误
func NewGitOperationError(operation, path string, err error) *GitOperationError {
	return &GitOperationError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}

// AIProviderError 表示 AI Provider 错误
type AIProviderError struct {
	Provider string
	Message  string
	Err      error
}

func (e *AIProviderError) Error() string {
	return fmt.Sprintf("AI provider '%s' error: %s", e.Provider, e.Message)
}

func (e *AIProviderError) Unwrap() error {
	return e.Err
}

// NewAIProviderError 创建 AI Provider 错误
func NewAIProviderError(provider, message string, err error) *AIProviderError {
	return &AIProviderError{
		Provider: provider,
		Message:  message,
		Err:      err,
	}
}

// IsNotFoundError 检查是否是"未找到"错误
func IsNotFoundError(err error) bool {
	return err != nil && (err.Error() == "record not found" || err.Error() == "not found")
}

// IsValidationError 检查是否是验证错误
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsGitError 检查是否是 Git 操作错误
func IsGitError(err error) bool {
	_, ok := err.(*GitOperationError)
	return ok
}

// IsAIProviderError 检查是否是 AI Provider 错误
func IsAIProviderError(err error) bool {
	_, ok := err.(*AIProviderError)
	return ok
}
