package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExcludeMode 排除模式类型
type ExcludeMode string

const (
	ExcludeModeExact     ExcludeMode = "exact"     // 精确文件名
	ExcludeModeExtension ExcludeMode = "extension" // 扩展名
	ExcludeModeDirectory ExcludeMode = "directory" // 目录
)

// DirectoryOption 目录选项
type DirectoryOption struct {
	Pattern string `json:"pattern"` // .gitignore 模式
	Label   string `json:"label"`   // 显示标签
}

// GetDirectoryOptions 获取目录层级选项
func GetDirectoryOptions(filePath string) []DirectoryOption {
	// 输入验证：空路径返回空数组
	if filePath == "" {
		return []DirectoryOption{}
	}

	// 转换为 Git 标准路径格式
	gitPath := toGitPath(filePath)
	parts := strings.Split(gitPath, "/")

	var options []DirectoryOption
	var pathBuilder strings.Builder

	// 构建层级选项（排除文件名）
	for i := 0; i < len(parts)-1; i++ {
		if i > 0 {
			pathBuilder.WriteString("/")
		}
		pathBuilder.WriteString(parts[i])

		pattern := pathBuilder.String()
		options = append(options, DirectoryOption{
			Pattern: pattern,
			Label:   pattern,
		})
	}

	// 添加"目录下所有扩展名"选项
	if len(parts) > 1 {
		dir := pathBuilder.String()
		ext := filepath.Ext(filePath)
		options = append(options, DirectoryOption{
			Pattern: dir + "/*" + ext,
			Label:   dir + "/*" + ext,
		})
	}

	return options
}

// GenerateGitIgnorePattern 生成 .gitignore 规则
func GenerateGitIgnorePattern(filePath string, mode ExcludeMode) (string, error) {
	gitPath := toGitPath(filePath)

	switch mode {
	case ExcludeModeExact:
		return gitPath, nil

	case ExcludeModeExtension:
		ext := filepath.Ext(filePath)
		if ext == "" {
			return "", fmt.Errorf("文件没有扩展名")
		}
		return "*" + ext, nil

	case ExcludeModeDirectory:
		dir := filepath.Dir(filePath)
		if dir == "." || dir == "" {
			return "/", nil
		}
		return toGitPath(dir), nil

	default:
		return "", fmt.Errorf("未知的排除模式: %s", mode)
	}
}

// AddToGitIgnoreFile 添加规则到 .gitignore 文件
func AddToGitIgnoreFile(projectPath, pattern string) error {
	gitIgnorePath := filepath.Join(projectPath, ".gitignore")

	// 获取现有文件权限
	existingPerms := os.FileMode(0644)
	if info, err := os.Stat(gitIgnorePath); err == nil {
		existingPerms = info.Mode().Perm()
	}

	// 读取现有内容
	var content []string
	if data, err := os.ReadFile(gitIgnorePath); err == nil {
		content = strings.Split(string(data), "\n")
	}

	// 检查是否已存在
	pattern = strings.TrimSpace(pattern)
	for _, line := range content {
		if strings.TrimSpace(line) == pattern {
			return nil // 已存在，不重复添加
		}
	}

	// 追加新规则
	content = append(content, pattern, "")
	return os.WriteFile(gitIgnorePath, []byte(strings.Join(content, "\n")), existingPerms)
}

// toGitPath 转换为 Git 标准路径格式
func toGitPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
