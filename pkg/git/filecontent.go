package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/WQGroup/logger"
)

// FileContentResult 文件内容读取结果
type FileContentResult struct {
	Content  string // 文件内容（文本文件）
	IsBinary bool   // 是否为二进制文件
}

// ReadFileContent 读取文件内容并判断是否为二进制
func ReadFileContent(repoPath, filePath string) (FileContentResult, error) {
	fullPath := filepath.Join(repoPath, filePath)
	logger.Infof("[ReadFileContent] 开始读取文件: %s", fullPath)
	logger.Infof("[ReadFileContent] repoPath: %s, filePath: %s", repoPath, filePath)

	// 读取文件内容
	data, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Errorf("[ReadFileContent] 读取文件失败: %v", err)
		return FileContentResult{}, fmt.Errorf("读取文件失败: %w", err)
	}

	logger.Infof("[ReadFileContent] 文件大小: %d 字节", len(data))

	// 判断是否为二进制文件（复用现有的 isBinary 函数）
	if isBinary(data) {
		logger.Infof("[ReadFileContent] 检测到二进制文件")
		return FileContentResult{
			Content:  "",
			IsBinary: true,
		}, nil
	}

	content := string(data)
	logger.Infof("[ReadFileContent] 文本文件内容长度: %d 字符", len(content))
	logger.Infof("[ReadFileContent] 内容预览 (前100字符): %s", truncateString(content, 100))

	return FileContentResult{
		Content:  content,
		IsBinary: false,
	}, nil
}

// truncateString 截断字符串用于日志
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
