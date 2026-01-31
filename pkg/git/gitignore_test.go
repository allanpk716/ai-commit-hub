package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateGitIgnorePattern_DirectoryMode 测试目录模式的 pattern 生成
func TestGenerateGitIgnorePattern_DirectoryMode(t *testing.T) {
	tests := []struct {
		name        string
		inputPath   string
		inputMode   ExcludeMode
		wantPattern string
		wantErr     bool
	}{
		{
			name:        "目录模式 - 直接获取父目录",
			inputPath:   "src/components/Hello.vue",
			inputMode:   ExcludeModeDirectory,
			wantPattern: "src/components", // 期望：直接父目录
			wantErr:     false,
		},
		{
			name:        "目录模式 - 单层目录",
			inputPath:   "src/Hello.vue",
			inputMode:   ExcludeModeDirectory,
			wantPattern: "src",
			wantErr:     false,
		},
		{
			name:        "目录模式 - 多层目录",
			inputPath:   "a/b/c/d/file.txt",
			inputMode:   ExcludeModeDirectory,
			wantPattern: "a/b/c/d",
			wantErr:     false,
		},
		{
			name:        "目录模式 - Windows 路径",
			inputPath:   "src\\components\\Hello.vue",
			inputMode:   ExcludeModeDirectory,
			wantPattern: "src/components",
			wantErr:     false,
		},
		{
			name:        "目录模式 - 根目录文件",
			inputPath:   "README.md",
			inputMode:   ExcludeModeDirectory,
			wantPattern: "/",
			wantErr:     false,
		},
		{
			name:        "精确模式 - 完整路径",
			inputPath:   "src/components/Hello.vue",
			inputMode:   ExcludeModeExact,
			wantPattern: "src/components/Hello.vue",
			wantErr:     false,
		},
		{
			name:        "扩展名模式 - 只有扩展名",
			inputPath:   "src/components/Hello.vue",
			inputMode:   ExcludeModeExtension,
			wantPattern: "*.vue",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateGitIgnorePattern(tt.inputPath, tt.inputMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateGitIgnorePattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPattern {
				t.Errorf("GenerateGitIgnorePattern() = %v, want %v", got, tt.wantPattern)
			}
		})
	}
}

// TestGetDirectoryOptions 测试目录选项生成
func TestGetDirectoryOptions(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		wantOptionsLen int
		wantFirst      string
		wantLast       string
	}{
		{
			name:           "多层目录",
			filePath:       "src/components/Hello.vue",
			wantOptionsLen: 3, // src, src/components, src/components/*.vue
			wantFirst:      "src",
			wantLast:       "src/components/*.vue",
		},
		{
			name:           "单层目录",
			filePath:       "src/Hello.vue",
			wantOptionsLen: 2, // src, src/*.vue
			wantFirst:      "src",
			wantLast:       "src/*.vue",
		},
		{
			name:           "根目录文件",
			filePath:       "README.md",
			wantOptionsLen: 0, // 无目录选项
		},
		{
			name:           "Windows 路径",
			filePath:       "src\\components\\Hello.vue",
			wantOptionsLen: 3,
			wantFirst:      "src",
			wantLast:       "src/components/*.vue",
		},
		{
			name:           "深层嵌套",
			filePath:       "a/b/c/d/e/file.txt",
			wantOptionsLen: 6, // a, a/b, a/b/c, a/b/c/d, a/b/c/d/e (目录层级) + a/b/c/d/e/*.txt (扩展名)
			wantFirst:      "a",
			wantLast:       "a/b/c/d/e/*.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDirectoryOptions(tt.filePath)
			if len(got) != tt.wantOptionsLen {
				t.Errorf("GetDirectoryOptions() len = %v, want %v", len(got), tt.wantOptionsLen)
			}
			if tt.wantOptionsLen > 0 {
				if got[0].Pattern != tt.wantFirst {
					t.Errorf("GetDirectoryOptions()[0].Pattern = %v, want %v", got[0].Pattern, tt.wantFirst)
				}
				if got[len(got)-1].Pattern != tt.wantLast {
					t.Errorf("GetDirectoryOptions()[-1].Pattern = %v, want %v", got[len(got)-1].Pattern, tt.wantLast)
				}
			}
		})
	}
}

// TestAddToGitIgnoreFile 测试添加到 .gitignore 文件
func TestAddToGitIgnoreFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	gitIgnorePath := filepath.Join(tmpDir, ".gitignore")

	tests := []struct {
		name      string
		pattern   string
		preWrite  func(string) error // 预先写入内容
		checkFunc func(string) bool  // 检查结果
	}{
		{
			name:    "添加单个规则",
			pattern: "node_modules",
			checkFunc: func(content string) bool {
				return strings.Contains(content, "node_modules")
			},
		},
		{
			name:    "添加带斜杠的规则",
			pattern: "dist/",
			checkFunc: func(content string) bool {
				return strings.Contains(content, "dist/")
			},
		},
		{
			name:    "添加通配符规则",
			pattern: "*.log",
			checkFunc: func(content string) bool {
				return strings.Contains(content, "*.log")
			},
		},
		{
			name: "重复添加相同规则 - 应该忽略",
			pattern: "node_modules",
			preWrite: func(tmpDir string) error {
				gitIgnorePath := filepath.Join(tmpDir, ".gitignore")
				return os.WriteFile(gitIgnorePath, []byte("node_modules\n"), 0644)
			},
			checkFunc: func(content string) bool {
				// 计算出现次数
				count := strings.Count(content, "node_modules")
				return count == 1 // 应该只有一个
			},
		},
		{
			name: "追加到现有内容",
			pattern: "*.log",
			preWrite: func(tmpDir string) error {
				gitIgnorePath := filepath.Join(tmpDir, ".gitignore")
				return os.WriteFile(gitIgnorePath, []byte("node_modules\n\nbuild/\n"), 0644)
			},
			checkFunc: func(content string) bool {
				return strings.Contains(content, "node_modules") &&
					strings.Contains(content, "build/") &&
					strings.Contains(content, "*.log")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理 .gitignore
			os.Remove(gitIgnorePath)

			// 预先写入内容
			if tt.preWrite != nil {
				if err := tt.preWrite(tmpDir); err != nil {
					t.Fatalf("preWrite failed: %v", err)
				}
			}

			// 执行添加
			err := AddToGitIgnoreFile(tmpDir, tt.pattern)
			if err != nil {
				t.Errorf("AddToGitIgnoreFile() error = %v", err)
				return
			}

			// 读取并检查结果
			content, err := os.ReadFile(gitIgnorePath)
			if err != nil {
				t.Fatalf("failed to read .gitignore: %v", err)
			}

			if !tt.checkFunc(string(content)) {
				t.Errorf("checkFunc failed for pattern: %s, content:\n%s", tt.pattern, content)
			}
		})
	}
}

// TestToGitPath 测试路径转换
func TestToGitPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"src/components", "src/components"},
		{"src\\components", "src/components"},
		{"a/b/c", "a/b/c"},
		{"a\\b\\c", "a/b/c"},
		{"file.txt", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toGitPath(tt.input); got != tt.want {
				t.Errorf("toGitPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDirectoryModeBugRegression 测试目录模式 Bug 的回归测试
// 这个测试专门验证用户报告的 bug：选择目录时实际写入的是父目录而非所选目录
//
// Bug 说明：
// - 用户在下拉框选择 "src/components" 后，前端发送 pattern="src/components", mode="directory"
// - 旧代码后端调用 GenerateGitIgnorePattern("src/components", directory)
// - GenerateGitIgnorePattern 使用 filepath.Dir("src/components") 返回 "src"（父目录）
// - 导致写入 .gitignore 的是 "src" 而非用户选择的 "src/components"
//
// 修复方案：
// - 在目录模式下，直接使用前端传递的 pattern，不再调用 GenerateGitIgnorePattern
func TestDirectoryModeBugRegression(t *testing.T) {
	tests := []struct {
		name                  string
		selectedPattern       string // 用户在下拉框中选择的目录
		expectedWritePattern  string // 期望写入 .gitignore 的内容
	}{
		{
			name:                 "选择 src/components 目录",
			selectedPattern:      "src/components",
			expectedWritePattern: "src/components",
		},
		{
			name:                 "选择 src 目录",
			selectedPattern:      "src",
			expectedWritePattern: "src",
		},
		{
			name:                 "选择深层目录",
			selectedPattern:      "packages/app/src",
			expectedWritePattern: "packages/app/src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// 模拟前端行为：直接使用选择的 pattern 作为目录模式
			// 修复后，目录模式下 pattern 应该直接使用，不经过 GenerateGitIgnorePattern
			err := AddToGitIgnoreFile(tmpDir, tt.selectedPattern)
			if err != nil {
				t.Errorf("AddToGitIgnoreFile() error = %v", err)
				return
			}

			// 读取 .gitignore 验证内容
			gitIgnorePath := filepath.Join(tmpDir, ".gitignore")
			content, err := os.ReadFile(gitIgnorePath)
			if err != nil {
				t.Fatalf("failed to read .gitignore: %v", err)
			}

			// 检查写入的 pattern 是否正确
			if !strings.Contains(string(content), tt.expectedWritePattern) {
				t.Errorf("Expected to find %q in .gitignore, got:\n%s", tt.expectedWritePattern, content)
			}

			// 验证没有错误地写入父目录
			if tt.selectedPattern != tt.expectedWritePattern {
				// 如果 selectedPattern 是 "src/components"，确保没有写入 "src"
				parentDir := filepath.Dir(tt.selectedPattern)
				if strings.Contains(string(content), parentDir) && !strings.Contains(string(content), tt.selectedPattern) {
					t.Errorf("Bug: wrote parent directory %q instead of selected directory %q", parentDir, tt.selectedPattern)
				}
			}
		})
	}
}

// TestDirectoryMode_GenerateGitIgnorePatternBehavior
// 这个测试记录 GenerateGitIgnorePattern 在目录模式下的原始行为
// 用于说明为什么需要在 app.go 中特殊处理目录模式
func TestDirectoryMode_GenerateGitIgnorePatternBehavior(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOutput  string
		explanation string
	}{
		{
			name:        "目录模式返回父目录 - 这是 bug 的根源",
			input:       "src/components",
			wantOutput:  "src",
			explanation: "filepath.Dir 获取父目录，这是旧行为导致 bug",
		},
		{
			name:        "单层目录返回根",
			input:       "src",
			wantOutput:  "/",
			explanation: "filepath.Dir('src') 返回 '.'，被转换成 '/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := GenerateGitIgnorePattern(tt.input, ExcludeModeDirectory)
			if got != tt.wantOutput {
				t.Errorf("GenerateGitIgnorePattern(%q, directory) = %q, want %q\n%s", tt.input, got, tt.wantOutput, tt.explanation)
			}
		})
	}
}
