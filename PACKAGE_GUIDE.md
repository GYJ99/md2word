# md2word 跨平台编译指南

## 🎯 三个专用编译脚本

### 1. 命令行版本跨平台编译
```bash
./scripts/package-cli.sh
```
**功能**: 本地跨平台编译，支持所有平台
**输出**: `dist-cli/` 目录
- ✅ Windows: `md2word-windows-amd64.exe` (15M)
- ✅ Linux: `md2word-linux-amd64`, `md2word-linux-arm64` (15M/14M)
- ✅ macOS: `md2word-darwin-amd64`, `md2word-darwin-arm64` (15M)
- 📦 分发包: `.zip` (Windows) 和 `.tar.gz` (Unix)

### 2. GUI 版本编译（当前平台）
```bash
./scripts/package-gui.sh
```
**功能**: 编译当前平台的 GUI 版本
**输出**: `dist-gui/` 目录
- ✅ 当前平台: `md2word-gui-{os}-{arch}` (30M)
- 📦 分发包: `gui-{os}-{arch}.tar.gz` 或 `.zip`

### 3. GUI 版本 Docker 编译（Linux）
```bash
./scripts/package-gui-docker.sh
```
**功能**: 使用 Docker 编译 Linux 版本的 GUI
**输出**: `dist-gui-docker/` 目录
- ✅ Linux ARM64: `md2word-gui-linux-arm64` (29M)
- 📦 分发包: `gui-linux-arm64.tar.gz` (12M)

**⚠️ GUI 版本限制**: 由于 Fyne 框架依赖 OpenGL 和平台特定的 C 库，GUI 版本无法进行完整跨平台编译。Docker 方案可以编译 Linux 版本，其他平台需要在目标平台上编译。

## 🚀 快速使用

```bash
# 编译命令行版本（所有平台）
./scripts/package-cli.sh

# 编译 GUI 版本（当前平台）
./scripts/package-gui.sh

# 编译 GUI 版本（Docker Linux）
./scripts/package-gui-docker.sh

# 测试运行
./dist-cli/md2word-darwin-arm64 --help
./dist-gui/md2word-gui-darwin-arm64
./dist-gui-docker/md2word-gui-linux-arm64
```

## 📁 输出结构

```
dist-cli/           # 命令行版本（跨平台）
├── md2word-windows-amd64.exe
├── md2word-linux-amd64
├── md2word-darwin-amd64
└── *.tar.gz, *.zip

dist-gui/           # GUI 版本（当前平台）
├── md2word-gui-darwin-arm64
└── gui-darwin-arm64.tar.gz

dist-gui-docker/    # GUI 版本（Docker Linux）
├── md2word-gui-linux-arm64
└── gui-linux-arm64.tar.gz
```

## 💡 多平台 GUI 编译

如需其他平台的 GUI 版本，请在对应平台上运行：

```bash
# Windows 上
go build -o md2word-gui.exe ./cmd/md2word-gui

# Linux 上
go build -o md2word-gui ./cmd/md2word-gui

# macOS 上
go build -o md2word-gui ./cmd/md2word-gui
```

## 📋 脚本说明

| 脚本 | 功能 | 支持平台 | 输出目录 | 文件大小 |
|------|------|----------|----------|----------|
| `package-cli.sh` | 命令行版本编译 | 跨平台 | `dist-cli/` | 15M |
| `package-gui.sh` | GUI 版本编译 | 当前平台 | `dist-gui/` | 30M |
| `package-gui-docker.sh` | GUI Docker 编译 | Linux | `dist-gui-docker/` | 29M |

## ✅ 测试结果

### 命令行版本（跨平台编译成功）
- ✅ Windows AMD64: 15M
- ✅ Linux AMD64: 15M  
- ✅ Linux ARM64: 14M
- ✅ macOS AMD64: 15M
- ✅ macOS ARM64: 15M

### GUI 版本（当前平台编译成功）
- ✅ macOS ARM64: 30M

### GUI 版本（Docker 编译成功）
- ✅ Linux ARM64: 29M (使用 golang:latest 镜像)

## 🔧 技术说明

- **CLI 版本**: 使用 `CGO_ENABLED=0` 进行静态编译，支持完整跨平台编译
- **GUI 版本**: 使用 Fyne 框架，依赖 CGO 和平台特定库
  - 本地编译：仅支持当前平台编译
  - Docker 编译：支持 Linux 版本编译，自动检测容器架构
- **分发包**: 自动创建压缩包，Windows 使用 ZIP，Unix 系统使用 TAR.GZ

## 🐳 Docker 编译详情

Docker GUI 编译脚本特性：
- 自动尝试多个 Go 镜像（golang:latest, golang:1.24-alpine 等）
- 自动检测容器架构并编译对应版本
- 自动安装所需的 GUI 开发依赖
- 支持 Alpine 和 Debian/Ubuntu 基础镜像

简洁实用，专注编译！