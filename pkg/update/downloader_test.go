package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir)

	if downloader == nil {
		t.Fatal("NewDownloader() 返回 nil")
	}

	if downloader.downloadDir != tempDir {
		t.Errorf("downloadDir = %s, want %s", downloader.downloadDir, tempDir)
	}
}

func TestDownload_Success(t *testing.T) {
	// 创建测试服务器
	content := "test content for download"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	// 创建下载器
	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir)

	progressCalls := 0
	downloader.SetProgressFunc(func(downloaded, total int64) {
		progressCalls++
	})

	// 执行下载
	destPath, err := downloader.Download(server.URL, "test.txt")
	if err != nil {
		t.Fatalf("Download() 失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatalf("文件未创建: %s", destPath)
	}

	// 读取并验证内容
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if string(data) != content {
		t.Errorf("文件内容不匹配: got %s, want %s", string(data), content)
	}

	// 验证进度回调
	if progressCalls == 0 {
		t.Error("进度回调未被调用")
	}

	// 清理
	os.Remove(destPath)
}

func TestDownload_Error(t *testing.T) {
	// 创建测试服务器（返回错误）
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir)

	_, err := downloader.Download(server.URL, "test.txt")
	if err == nil {
		t.Error("预期返回错误，但返回成功")
	}
}

func TestDownload_ProgressTracking(t *testing.T) {
	// 创建大文件测试进度跟踪
	largeContent := strings.Repeat("x", 1024*100) // 100KB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(largeContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir)

	progressUpdates := 0
	downloader.SetProgressFunc(func(downloaded, total int64) {
		progressUpdates++
	})

	destPath, err := downloader.Download(server.URL, "large.txt")
	if err != nil {
		t.Fatalf("Download() 失败: %v", err)
	}

	// 验证文件大小
	info, _ := os.Stat(destPath)
	if info.Size() != int64(len(largeContent)) {
		t.Errorf("文件大小不匹配: got %d, want %d", info.Size(), len(largeContent))
	}

	// 验证进度更新（本地测试服务器速度快，可能只有 1 次）
	if progressUpdates < 1 {
		t.Errorf("进度更新次数过少: got %d, want >= 1", progressUpdates)
	}

	os.Remove(destPath)
}

func TestDownload_Cancel(t *testing.T) {
	tempDir := t.TempDir()
	downloader := NewDownloader(tempDir)

	// 创建一个假文件
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0o644)

	// 取消下载
	err := downloader.Cancel("test.txt")
	if err != nil {
		t.Fatalf("Cancel() 失败: %v", err)
	}

	// 验证文件已删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("文件仍存在，应该已被删除")
	}
}
