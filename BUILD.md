# Windows 构建说明

## 图标已成功嵌入！

构建的应用程序 `build\bin\ai-commit-hub.exe` 已成功嵌入了自定义图标。

## 快速构建

使用自动化脚本构建：

```bash
scripts\build-windows.bat
```

这个脚本会自动：
1. 安装 rsrc 工具（如果需要）
2. 生成 icon.syso 文件
3. 构建 Wails 应用
4. 验证图标是否嵌入

## 图标相关文件

- `build/windows/icon.ico` - Windows 图标源文件
- `icon.syso` - 自动生成的资源文件（不需要手动编辑）
- `assets/build/windows/icon.ico` - 图标备份

## 常见问题

### Q: 构建后看不到图标？

A: 这是 Windows 图标缓存问题。运行：

```bash
scripts\clear-icon-cache.bat
```

### Q: wails dev 有图标，但 wails build 没有？

A: 确保在构建前生成了 `icon.syso` 文件。使用 `scripts\build-windows.bat` 脚本会自动处理。

### Q: 图标没有更新？

A: Windows 会缓存图标。尝试：
1. 运行 `scripts\clear-icon-cache.bat`
2. 重启文件管理器（Win+R → `restart`）
3. 或注销后重新登录

## 验证图标

运行验证脚本：

```powershell
powershell -ExecutionPolicy Bypass -File test_icon.ps1
```

或手动验证：

```powershell
Add-Type -AssemblyName System.Drawing
$icon = [System.Drawing.Icon]::ExtractAssociatedIcon("build\bin\ai-commit-hub.exe")
$bitmap = $icon.ToBitmap()
$bitmap.Save("verify_icon.png")
```

## 技术细节

Windows 应用图标需要通过 `.syso` 文件嵌入。这是 Go 编译器的标准方式：

1. 使用 `rsrc` 工具将 `.ico` 文件转换为 `.syso`
2. `.syso` 文件与 `main.go` 在同一目录
3. Go 编译器自动将 `.syso` 链接到可执行文件

## 相关脚本

- `scripts/build-windows.bat` - 完整构建脚本
- `scripts/clear-icon-cache.bat` - 清除 Windows 图标缓存
- `scripts/init-icons.bat` - 初始化图标文件
