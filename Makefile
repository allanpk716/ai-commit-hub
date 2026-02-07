.PHONY: test test-unit test-integration test-e2e test-frontend test-all coverage clean help updater build

# 默认目标
help:
	@echo "AI Commit Hub - 测试命令"
	@echo ""
	@echo "可用命令:"
	@echo "  make test-all       - 运行所有测试"
	@echo "  make test-unit      - 运行后端单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make coverage       - 生成覆盖率报告"
	@echo "  make clean          - 清理测试文件"
	@echo "  make updater        - 构建更新器"
	@echo "  make build          - 构建完整应用程序"

# 运行所有测试
test-all:
	@echo "运行所有测试..."
	@scripts/run-tests.bat

# 后端单元测试
test-unit:
	@echo "运行后端单元测试..."
	@go test ./pkg/git/... -v
	@go test ./pkg/service/... -v
	@go test ./pkg/repository/... -v
	@go test ./tests/helpers/... -v

# 集成测试
test-integration:
	@echo "运行集成测试..."
	@go test ./tests/integration/... -v

# 覆盖率报告
coverage:
	@echo "生成覆盖率报告..."
	@go test ./pkg/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

# 构建更新器（必须在主程序构建前执行）
updater:
	@echo "构建更新器..."
	@go build -o pkg/update/updater/updater.exe ./pkg/update/updater
	@echo "更新器构建完成: pkg/update/updater/updater.exe"

# 构建完整应用程序（包括更新器）
build: updater
	@echo "构建主应用程序..."
	@wails build
	@echo "构建完成"

# 清理测试文件
clean:
	@echo "清理测试文件..."
	@if exist tmp\test-results rmdir /s /q tmp\test-results
	@if exist coverage.out del /q coverage.out
	@if exist coverage.html del /q coverage.html
