package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/version"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/src/assets/app-icon.png
var appIconPNG []byte

//go:embed build/windows/icon.ico
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
	// 检查命令行参数（版本标志）
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Println(version.GetFullVersion())
			os.Exit(0)
		case "-h", "--help":
			fmt.Println("AI Commit Hub - Git commit message generator")
			fmt.Println("\nUsage:")
			fmt.Println("  ai-commit-hub [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  -v, --version  Show version information")
			fmt.Println("  -h, --help     Show this help message")
			os.Exit(0)
		}
	}

	// 初始化 logger
	initLogger()

	// 输出版本信息
	logger.Info("AI Commit Hub starting up...", "version", version.GetVersion())
	logger.Debug("Full version info", "info", version.GetFullVersion())

	// Create an instance of the app structure
	app := NewApp()

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
