package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/WQGroup/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Downloader 下载器
type Downloader struct {
	client        *http.Client
	downloadDir   string
	onProgress    ProgressFunc
	maxRetries    int           // 最大重试次数
	retryInterval time.Duration // 重试间隔
	ctx           context.Context
	proxyURL      string // 代理 URL
	cancelled     bool   // 取消标志
}

// ProgressFunc 进度回调函数
type ProgressFunc func(downloaded, total int64)

// NewDownloader 创建下载器
func NewDownloader(downloadDir string) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		downloadDir:   downloadDir,
		maxRetries:    3,
		retryInterval: 5 * time.Second,
	}
}

// NewResumableDownloader 创建支持断点续传的下载器
func NewResumableDownloader(downloadDir string) *Downloader {
	return NewDownloader(downloadDir)
}

// SetContext 设置上下文（用于发送事件）
func (d *Downloader) SetContext(ctx context.Context) {
	d.ctx = ctx
}

// SetProgressFunc 设置进度回调
func (d *Downloader) SetProgressFunc(fn ProgressFunc) {
	d.onProgress = fn
}

// SetProxy 设置代理
func (d *Downloader) SetProxy(proxyURL string) {
	d.proxyURL = proxyURL

	// 解析代理 URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		logger.Errorf("解析代理 URL 失败: %v", err)
		return
	}

	// 创建带代理的 Transport
	d.client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	logger.Infof("已设置代理: %s", proxyURL)
}

// Download 下载文件（支持断点续传）
func (d *Downloader) Download(url, filename string) (string, error) {
	logger.Infof("开始下载: %s -> %s", url, filename)

	// 确保下载目录存在
	if err := os.MkdirAll(d.downloadDir, 0755); err != nil {
		return "", fmt.Errorf("创建下载目录失败: %w", err)
	}

	// 目标文件路径
	destPath := filepath.Join(d.downloadDir, filename)
	tmpPath := destPath + ".tmp"

	// 检查临时文件是否存在（断点续传）
	var downloaded int64
	if info, err := os.Stat(tmpPath); err == nil {
		downloaded = info.Size()
		logger.Infof("检测到临时文件，已下载: %d bytes", downloaded)
	}

	// 带重试的下载
	var lastErr error
	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if d.cancelled {
			return "", fmt.Errorf("下载已取消")
		}

		if attempt > 0 {
			logger.Infof("重试第 %d 次（共 %d 次）...", attempt, d.maxRetries)
			time.Sleep(d.retryInterval)
		}

		destPath, lastErr = d.downloadWithResume(url, tmpPath, destPath, downloaded)
		if lastErr == nil {
			// 下载成功
			return destPath, nil
		}

		// 检查是否为可恢复错误
		if !isRecoverableError(lastErr) {
			logger.Errorf("遇到不可恢复的错误: %v", lastErr)
			return "", lastErr
		}

		logger.Warnf("下载失败（可恢复）: %v", lastErr)
	}

	return "", fmt.Errorf("下载失败，已重试 %d 次: %w", d.maxRetries, lastErr)
}

// downloadWithResume 执行带断点续传的下载
func (d *Downloader) downloadWithResume(url, tmpPath, destPath string, downloaded int64) (string, error) {
	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Range 头（断点续传）
	if downloaded > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
		logger.Infof("设置断点续传: Range: bytes=%d-", downloaded)
	}

	// 发起请求
	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return "", fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	// 获取文件大小
	totalSize := resp.ContentLength
	if downloaded > 0 && resp.StatusCode == http.StatusPartialContent {
		// 断点续传时，需要从 Content-Range 头获取总大小
		// Content-Range: bytes 0-1023/2048
		totalSize = downloaded + resp.ContentLength
		logger.Infof("断点续传响应: 206 Partial Content, 总大小: %d", totalSize)
	}

	// 打开文件（追加模式）
	destFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer destFile.Close()

	// 创建进度写入器
	writer := &progressWriter{
		writer:       destFile,
		total:        totalSize,
		downloaded:   downloaded,
		onProgress:   d.onProgress,
		lastUpdate:   time.Now(),
		ctx:          d.ctx,
		url:          url,
		startTime:    time.Now(),
		lastWritten:  downloaded,
	}

	// 复制数据
	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	// 验证文件大小
	actualSize := downloaded + written
	if totalSize > 0 && actualSize != totalSize {
		return "", fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", totalSize, actualSize)
	}

	// 显式关闭文件，确保数据写入磁盘
	if err := destFile.Close(); err != nil {
		return "", fmt.Errorf("关闭文件失败: %w", err)
	}

	// 重命名临时文件为最终文件
	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", fmt.Errorf("重命名文件失败: %w", err)
	}

	// 下载完成时强制触发一次进度回调（100%）
	if d.onProgress != nil && totalSize > 0 {
		d.onProgress(actualSize, totalSize)
	}

	logger.Infof("下载完成: %s (%d bytes)", destPath, actualSize)
	return destPath, nil
}

// VerifyHash 验证文件哈希
func (d *Downloader) VerifyHash(filepath, expectedHash string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	if actualHash != expectedHash {
		return fmt.Errorf("哈希不匹配: 期望 %s, 实际 %s", expectedHash, actualHash)
	}

	logger.Infof("文件哈希验证通过: %s", actualHash)
	return nil
}

// progressWriter 进度写入器
type progressWriter struct {
	writer      io.Writer
	total       int64
	downloaded  int64
	written     int64
	onProgress  ProgressFunc
	lastUpdate  time.Time
	ctx         context.Context
	url         string
	startTime   time.Time
	lastWritten int64
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

	totalDownloaded := pw.downloaded + pw.written

	if pw.total > 0 {
		percentage := float64(totalDownloaded) / float64(pw.total) * 100
		if percentage >= 1.0 {
			// 计算速度
			elapsed := now.Sub(pw.startTime).Seconds()
			speed := int64(float64(totalDownloaded-pw.lastWritten) / elapsed)

			// 计算剩余时间
			remaining := pw.total - totalDownloaded
			var eta string
			if speed > 0 {
				etaSeconds := int(float64(remaining) / float64(speed))
				eta = formatDuration(time.Duration(etaSeconds) * time.Second)
			} else {
				eta = "计算中..."
			}

			// 通过 Wails Events 发送进度
			if pw.ctx != nil {
				runtime.EventsEmit(pw.ctx, "download-progress", map[string]interface{}{
					"percentage": float64(totalDownloaded) / float64(pw.total) * 100,
					"downloaded": totalDownloaded,
					"total":      pw.total,
					"speed":      speed,
					"eta":        eta,
					"url":        pw.url,
				})
			}

			// 调用旧的回调（兼容性）
			pw.onProgress(totalDownloaded, pw.total)
			pw.lastUpdate = now
			pw.lastWritten = totalDownloaded

			logger.Debugf("下载进度: %.1f%% (%d/%d bytes), 速度: %s/s, 剩余: %s",
				percentage, totalDownloaded, pw.total, formatBytes(speed), eta)
		}
	}
}

// Cancel 取消下载
func (d *Downloader) Cancel(filename string) error {
	d.cancelled = true

	destPath := filepath.Join(d.downloadDir, filename)
	tmpPath := destPath + ".tmp"

	// 删除临时文件
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除临时文件失败: %w", err)
	}

	// 删除目标文件
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	logger.Infof("已取消下载，删除文件: %s", destPath)
	return nil
}

// isRecoverableError 检查是否为可恢复错误
func isRecoverableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络错误通常是可恢复的
	// 超时错误、连接错误等
	return true
}

// formatBytes 格式化字节大小
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration 格式化时间长度
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分%d秒", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%d小时%d分", int(d.Hours()), int(d.Minutes())%60)
}
