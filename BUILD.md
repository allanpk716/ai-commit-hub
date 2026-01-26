# Windows 构建说明

## 快速开始

### 标准构建

项目已包含预生成的 Windows 资源文件（`.syso`），但推荐使用构建脚本以避免资源冲突警告：

**推荐方式（无警告）**：
```bash
scripts\build-windows.bat
```

**快速方式（有警告但可用）**：
```bash
wails build
```

**注意**：`wails build` 会产生链接器警告（duplicate leaf），但生成的 exe 文件正常可用。使用构建脚本可以避免这些警告。

### 图标资源说明

项目包含的 `rsrc_windows_amd64.syso` 文件提供了以下图标尺寸：
- **16x16** - 小图标（列表视图）
- **32x32** - 标准图标
- **48x48** - 大图标
- **64x64** - 超大图标
- **128x128** - 超超大图标
- **256x256** - 高 DPI 显示

这些资源文件已提交到版本控制，所有开发者 clone 仓库后都可以直接构建。

## 如何更新图标

如果你需要更新应用图标：

1. 替换源图标文件：`assets/icons/appicon.png`
2. 运行图标生成脚本：
   ```bash
   python scripts/prepare_icons.py
   go-winres make
   ```
3. 提交更新后的 `rsrc_windows_amd64.syso` 文件
4. 重新构建：`wails build`

## 多尺寸图标支持

应用现在支持完整的 Windows 图标尺寸，包括：
- **16x16** - 小图标（列表视图）
- **32x32** - 标准图标
- **48x48** - 大图标
- **64x64** - 超大图标
- **128x128** - 超超大图标
- **256x256** - 高 DPI 显示

这确保了应用在 Windows 文件管理器的任何视图大小下都能显示清晰的图标。

## 快速构建

使用自动化脚本构建（推荐）：

```bash
scripts\build-windows.bat
```

这个脚本会自动：
1. 安装 go-winres 工具（如果需要）
2. 使用 Python 生成多尺寸 PNG 图标
3. 生成 Windows 资源文件（.syso）
4. 构建前端
5. 使用 `go build` 编译后端（避免资源冲突警告）
6. 验证图标是否嵌入

## 手动构建

如果需要手动构建：

```bash
# 1. 生成多尺寸图标
python scripts\prepare_icons.py

# 2. 生成资源文件
go-winres make

# 3. 构建前端
cd frontend && npm run build && cd ..

# 4. 构建应用（推荐使用 go build 避免警告）
go build -o build/bin/ai-commit-hub.exe .

# 或者使用 wails build（会有链接器警告但可用）
# wails build
```

## 文件结构

```
ai-commit-hub/
├── assets/icons/appicon.png      # 源图标（最高分辨率）
├── build/windows/icon.ico        # Windows ICO 文件
├── winres/                       # go-winres 工作目录
│   ├── winres.json              # 资源配置文件
│   ├── icon16.png               # 自动生成的 16x16 图标
│   ├── icon32.png               # 自动生成的 32x32 图标
│   ├── icon48.png               # 自动生成的 48x48 图标
│   ├── icon64.png               # 自动生成的 64x64 图标
│   ├── icon128.png              # 自动生成的 128x128 图标
│   └── icon256.png              # 自动生成的 256x256 图标
└── scripts/
    ├── build-windows.bat        # 自动化构建脚本
    ├── prepare_icons.py         # 生成多尺寸图标
    └── clear-icon-cache.bat     # 清除 Windows 图标缓存
```

## 常见问题

### Q: 小图标正常，大图标显示不正确？

A: 这是 ICO 文件缺少多尺寸的问题。新的构建系统已经解决：

```bash
scripts\build-windows.bat
```

### Q: 构建后看不到图标？

A: Windows 图标缓存问题。运行：

```bash
scripts\clear-icon-cache.bat
```

### Q: 某些尺寸的图标模糊？

A: 确保：
1. `assets/icons/appicon.png` 是高分辨率源图（建议 512x512 或更高）
2. 重新运行 `scripts\build-windows.bat`

### Q: 如何更新图标？

A: 只需替换 `assets/icons/appicon.png`，然后重新构建：

```bash
scripts\build-windows.bat
```

## 技术细节

### 图标生成流程

1. **源图**：`assets/icons/appicon.png` (高分辨率)
2. **多尺寸生成**：Python PIL 将源图缩放到 6 个标准尺寸
3. **资源打包**：go-winres 将所有 PNG 打包到 .syso 文件
4. **嵌入 exe**：Go 编译器自动将 .syso 链接到可执行文件

### 为什么不用 rsrc？

rsrc 工具只支持单个 ICO 文件，难以生成高质量的多尺寸图标。go-winres 支持：
- 多个 PNG 文件作为输入
- 自动生成所有标准尺寸
- 更好的质量控制
- 内置 manifest 和版本信息支持

### 资源冲突说明

使用 `wails build` 时会出现链接器警告：
```
.rsrc merge failure: duplicate leaf: type: 3 (ICON) name: X lang: 0
```

**原因**：
- Wails 构建时会自动生成 Windows 资源（包含图标和 manifest）
- go-winres 也生成了图标资源
- 链接器在合并资源时发现重复定义

**影响**：
- 警告是非致命的，exe 文件可以正常生成和使用
- 图标仍然能正确显示

**解决方案**：
- **推荐**：使用 `scripts\build-windows.bat`，它会用 `go build` 避免冲突
- **快速**：直接用 `wails build`，忽略警告
- 注意：`build/windows/icon.ico` 已重命名为 `.bak` 以避免冲突

## 相关工具

- **go-winres** - Windows 资源文件生成工具
  - 安装：`go install github.com/tc-hib/go-winres@latest`
  - 文档：https://github.com/tc-hib/go-winres

- **Python PIL** - 图像处理库
  - 用于生成多尺寸 PNG 图标
  - 安装：`pip install pillow`

## 验证图标

### 方法 1：文件管理器

1. 打开 `build\bin\` 目录
2. 右键 → 查看 → 选择不同的图标大小
3. 验证所有尺寸下图标都清晰

### 方法 2：桌面快捷方式

1. 创建桌面快捷方式
2. 右键快捷方式 → 属性 → 更改图标
3. 查看不同尺寸的预览

### 方法 3：PowerShell

```powershell
Add-Type -AssemblyName System.Drawing
$icon = [System.Drawing.Icon]::ExtractAssociatedIcon("build\bin\ai-commit-hub.exe")
Write-Host "Icon size: $($icon.Width) x $($icon.Height)"
```

## 更新日志

- **2026-01-26** - 解决 Wails/go-winres 资源冲突，使用 go build 避免警告
- **2026-01-26** - 将资源文件加入版本控制，简化构建流程
- **2026-01-26** - 切换到 go-winres，支持完整的多尺寸图标
- **2026-01-26** - 修复 Windows 图标缓存问题
- **2026-01-26** - 添加自动化构建脚本
