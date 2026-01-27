package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileContentResult 文件内容读取结果
type FileContentResult struct {
	Content  string // 文件内容（文本文件）
	IsBinary bool   // 是否为二进制文件
}

// ReadFileContent 读取文件内容并判断是否为二进制
func ReadFileContent(repoPath, filePath string) (FileContentResult, error) {
	fullPath := filepath.Join(repoPath, filePath)

	// 读取文件内容
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return FileContentResult{}, fmt.Errorf("读取文件失败: %w", err)
	}

	// 判断是否为二进制文件（复用现有的 isBinary 函数）
	if isBinary(data) {
		return FileContentResult{
			Content:  "",
			IsBinary: true,
		}, nil
	}

	return FileContentResult{
		Content:  string(data),
		IsBinary: false,
	}, nil
}
