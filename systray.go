package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/allanpk716/ai-commit-hub/pkg/constants"
	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SystrayManager 管理系统托盘
type SystrayManager struct {
	app         *App
	quitOnce    sync.Once
	windowShown bool
	currentIcon []byte
	iconLoaded  bool
}

// NewSystrayManager 创建托盘管理器
func NewSystrayManager(app *App) *SystrayManager {
	return &SystrayManager{
		app: app,
	}
}

// Start 启动系统托盘
func (sm *SystrayManager) Start() {
	systray.Run(sm.onReady, sm.onExit)
}

// onReady 托盘准备就绪回调
func (sm *SystrayManager) onReady() {
	sm.setInitialIcon()
	sm.setupMenu()
}

// setInitialIcon 设置初始图标
func (sm *SystrayManager) setInitialIcon() {
	time.Sleep(constants.SystrayInitDelay)

	if sm.currentIcon != nil {
		systray.SetIcon(sm.currentIcon)
		sm.iconLoaded = true
		return
	}

	sm.loadAndSetIcon()
}

// loadAndSetIcon 加载并设置图标
func (sm *SystrayManager) loadAndSetIcon() {
	for attempt := 0; attempt < constants.MaxIconRetryAttempts; attempt++ {
		icon, err := sm.loadAppIcon()
		if err != nil {
			time.Sleep(constants.IconRetryDelay)
			continue
		}

		sm.currentIcon = icon
		systray.SetIcon(icon)
		sm.iconLoaded = true
		time.Sleep(constants.IconSettleDelay)
		return
	}

	fmt.Println("Warning: Failed to load tray icon after multiple attempts")
}

// setupMenu 设置托盘菜单
func (sm *SystrayManager) setupMenu() {
	systray.SetTitle("AI Commit Hub")
	systray.SetTooltip("AI Commit Hub - 双击显示窗口")

	showWindow := systray.AddMenuItem("显示窗口", "显示主窗口")
	quitButton := systray.AddMenuItem("退出应用", "完全退出应用")

	go func() {
		for {
			select {
			case <-showWindow.ClickedCh:
				sm.ShowWindow()
			case <-quitButton.ClickedCh:
				sm.Quit()
			}
		}
	}()
}

// ShowWindow 显示主窗口
func (sm *SystrayManager) ShowWindow() {
	if sm.windowShown {
		return
	}

	sm.windowShown = true
	runtime.WindowShow(sm.app.ctx)
	time.Sleep(constants.WindowShowDelay)
	sm.windowShown = false
}

// Quit 退出应用
func (sm *SystrayManager) Quit() {
	sm.quitOnce.Do(func() {
		systray.Quit()
	})
}

// onExit 托盘退出回调
func (sm *SystrayManager) onExit() {
	sm.app.onSystrayExit()
}

// loadAppIcon 加载应用图标
func (sm *SystrayManager) loadAppIcon() ([]byte, error) {
	// 调用 App 的 getTrayIcon 方法
	return sm.app.getTrayIcon(), nil
}

// loadDarwinIcon 加载 macOS 图标（占位实现）
func (sm *SystrayManager) loadDarwinIcon() ([]byte, error) {
	return sm.app.getTrayIcon(), nil
}

// loadDefaultIcon 加载默认图标（占位实现）
func (sm *SystrayManager) loadDefaultIcon() ([]byte, error) {
	return sm.app.getTrayIcon(), nil
}
