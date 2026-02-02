package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/WQGroup/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/src/assets/app-icon.png
var appIconPNG []byte

//go:embed frontend/src/assets/app-icon.ico
var appIconICO []byte

func initLogger() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get home directory: %v", err)
		return
	}

	// 创建日志目录
	logDir := filepath.Join(homeDir, ".ai-commit-hub", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Errorf("Failed to create log directory: %v", err)
		return
	}

	// 配置 logger 输出到文件
	logger.SetLoggerSettings(
		&logger.Settings{
			LogRootFPath:  logDir,
			LogNameBase:   "app.log",
			MaxSizeMB:     100,
			MaxAgeDays:    30,
			FormatterType: "text",
		},
	)

	logger.Info("Logger initialized, log directory:", logDir)
}

func main() {
	// 初始化 logger
	initLogger()

	// Create an instance of the app structure
	app := NewApp()

	// 启动系统托盘 (在 Wails 启动前)
	go app.runSystray()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "AI Commit Hub",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.onBeforeClose, // 新增: 拦截关闭
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		logger.Errorf("Error: %v", err)
	}
}
