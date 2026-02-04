package errors

import "fmt"

// AppInitError 表示应用初始化错误
type AppInitError struct {
	OriginalErr error
}

func (e *AppInitError) Error() string {
	return fmt.Sprintf("app not initialized: %v", e.OriginalErr)
}

func (e *AppInitError) Unwrap() error {
	return e.OriginalErr
}

// NewAppInitError 创建初始化错误
func NewAppInitError(err error) *AppInitError {
	return &AppInitError{OriginalErr: err}
}

// CheckInit 检查初始化错误并返回
func CheckInit(initErr error) error {
	if initErr != nil {
		return &AppInitError{OriginalErr: initErr}
	}
	return nil
}
