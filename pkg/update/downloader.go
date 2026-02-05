package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/WQGroup/logger"
)

// Downloader 下载器
type Downloader struct {
	client      *http.Client
	downloadDir string
	onProgress  ProgressFunc
}

// ProgressFunc 进度回调函数
type ProgressFunc func(downloaded, total int64)

// NewDownloader 创建下载器
func NewDownloader(downloadDir string) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		downloadDir: downloadDir,
	}
}

// SetProgressFunc 设置进度回调
func (d *Downloader) SetProgressFunc(fn ProgressFunc) {
	d.onProgress = fn
}

// Download 下载文件
func (d *Downloader) Download(url, filename string) (string, error) {
	logger.Infof("开始下载: %s -> %s", url, filename)

	// 确保下载目录存在
	if err := os.MkdirAll(d.downloadDir, 0o755); err != nil {
		return "", fmt.Errorf("创建下载目录失败: %w", err)
	}

	// 创建目标文件
	destPath := filepath.Join(d.downloadDir, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer destFile.Close()

	// 发起 HTTP 请求
	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	// 获取文件大小
	totalSize := resp.ContentLength

	// 创建写入器和跟踪器
	writer := &progressWriter{
		writer:     destFile,
		total:      totalSize,
		onProgress: d.onProgress,
		lastUpdate: time.Now(),
	}

	// 复制数据
	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	// 验证文件大小
	if totalSize > 0 && written != totalSize {
		return "", fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", totalSize, written)
	}

	// 下载完成时强制触发一次进度回调（100%）
	if d.onProgress != nil && totalSize > 0 {
		d.onProgress(written, totalSize)
	}

	logger.Infof("下载完成: %s (%d bytes)", destPath, written)
	return destPath, nil
}

// progressWriter 进度写入器
type progressWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	onProgress ProgressFunc
	lastUpdate time.Time
}

// Write 实现 io.Writer
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.written += int64(n)
		pw.reportProgress()
	}
	return n, err
}

// reportProgress 报告进度
func (pw *progressWriter) reportProgress() {
	if pw.onProgress == nil {
		return
	}

	now := time.Now()
	if now.Sub(pw.lastUpdate) < 100*time.Millisecond {
		return
	}

	if pw.total > 0 {
		percentage := float64(pw.written) / float64(pw.total) * 100
		if percentage >= 1.0 {
			pw.onProgress(pw.written, pw.total)
			pw.lastUpdate = now

			logger.Debugf("下载进度: %.1f%% (%d/%d bytes)",
				percentage, pw.written, pw.total)
		}
	}
}

// Cancel 取消下载
func (d *Downloader) Cancel(filename string) error {
	destPath := filepath.Join(d.downloadDir, filename)
	if err := os.Remove(destPath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	logger.Infof("已取消下载，删除文件: %s", destPath)
	return nil
}
