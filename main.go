package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/allanpk716/ai-commit-hub/pkg/version"
	"github.com/sirupsen/logrus"
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
	// 获取可执行文件所在目录（程序根目录）
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取可执行文件路径失败: %v\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	// 创建日志目录
	logDir := filepath.Join(exeDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
		return
	}

	// 输出调试信息
	fmt.Fprintf(os.Stderr, "[DEBUG] 日志目录: %s\n", logDir)

	// 配置 logger 输出到文件
	if err := logger.SetLoggerSettingsWithError(
		&logger.Settings{
			LogRootFPath:     logDir,
			LogNameBase:      "app", // logger 库会自动添加 .log 后缀
			MaxSizeMB:        100,
			MaxAgeDays:       30,
			FormatterType:    "withField",
			TimestampFormat:  "2006-01-02 15:04:05.000",
			DisableTimestamp: false,
			DisableLevel:     false,
			DisableCaller:    true,
			Level:            logrus.InfoLevel, // 明确设置日志级别为 Info
		},
	); err != nil {
		fmt.Fprintf(os.Stderr, "配置 logger 失败: %v\n", err)
		return
	}

	logger.Infof("日志初始化完成，日志目录: %s", logDir)
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
	defer logger.Close() // 确保程序退出时刷新日志

	// 输出版本信息
	logger.WithField("version", version.GetVersion()).Info("AI Commit Hub starting up...")
	logger.WithField("info", version.GetFullVersion()).Debug("Full version info")

	// 初始化数据库以读取窗口状态
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get home directory: %v", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	dbPath := filepath.Join(configDir, "ai-commit-hub.db")

	// 初始化数据库
	dbConfig := &repository.DatabaseConfig{Path: dbPath}
	if err := repository.InitializeDatabase(dbConfig); err != nil {
		logger.Warnf("Failed to initialize database for window state: %v", err)
		// 继续使用默认窗口大小
	}

	// 读取保存的窗口状态
	var initialWidth, initialHeight int = 1280, 800 // 默认值

	windowStateRepo := repository.NewWindowStateRepository()
	if state, err := windowStateRepo.GetByKey("window.main"); err == nil {
		// 成功读取到窗口状态
		if state.Width > 0 && state.Height > 0 {
			initialWidth = state.Width
			initialHeight = state.Height
			logger.WithFields(map[string]interface{}{
				"position":  fmt.Sprintf("(%d,%d)", state.X, state.Y),
				"size":      fmt.Sprintf("%dx%d", state.Width, state.Height),
				"maximized": state.Maximized,
			}).Info("从数据库恢复窗口状态")
			// 输出到控制台
			fmt.Printf("[MAIN] Window state from DB: Position(%d,%d) Size(%dx%d) Maximized=%v\n",
				state.X, state.Y, state.Width, state.Height, state.Maximized)
			fmt.Printf("[MAIN] Will create window with size: %dx%d\n", initialWidth, initialHeight)
		}
		// 注意：窗口位置需要在 startup() 中使用 runtime.WindowSetPosition() 设置
	} else {
		logger.Info("未找到保存的窗口状态，使用默认值")
		fmt.Printf("[MAIN] No saved state found, using default size: %dx%d\n", initialWidth, initialHeight)
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "AI Commit Hub",
		Width:  initialWidth,
		Height: initialHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.onBeforeClose, // 新增: 拦截关闭
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:              "e3984e08-28dc-4e3d-b70a-45e961589cdc",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
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
