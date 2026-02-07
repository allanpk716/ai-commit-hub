@echo off
REM 测试自动更新功能脚本
REM 使用方法：双击运行或在命令行执行

echo ========================================
echo AI Commit Hub - 自动更新测试模式
echo ========================================
echo.
echo 此脚本将启动应用并启用测试模式
echo 测试模式会模拟检测到新版本，可以测试下载功能
echo.

REM 检查 wails 是否安装
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 wails 命令
    echo 请先安装 wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    pause
    exit /b 1
)

REM 设置测试模式环境变量
set AI_COMMIT_HUB_TEST_MODE=true

echo [✓] 测试模式已启用
echo.
echo 测试信息：
echo   - 测试版本：v1.0.0-alpha.1
echo   - 下载 URL：GitHub Releases
echo   - 文件大小：~60MB
echo.
echo 开始启动应用...
echo.

REM 启动开发服务器
wails dev
