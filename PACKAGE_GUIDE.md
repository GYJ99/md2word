# md2word 打包指南

## 一键打包（推荐）

```bash
./scripts/build.sh
```

脚本会自动：
1. 检测当前操作系统和架构（`go env GOOS/GOARCH`）
2. 编译 CLI（`CGO_ENABLED=0`，纯静态二进制）
3. 编译 GUI（需 CGO，仅当前平台）
4. 把版本号和构建时间注入二进制（`main.Version`、`main.BuildTime`）
5. 生成分发包：Unix 用 `.tar.gz`，Windows 用 `.zip`

## 输出结构

```
dist-cli/
├── md2word-darwin-arm64        # CLI 二进制
└── darwin-arm64.tar.gz         # 分发包

dist-gui/
├── md2word-gui-darwin-arm64    # GUI 二进制
└── gui-darwin-arm64.tar.gz     # 分发包
```

文件名格式：`md2word[-gui]-{goos}-{goarch}[.exe]`

## 自定义版本号

```bash
VERSION=1.0.1 ./scripts/build.sh
```

构建完成后，GUI 状态栏会显示注入的版本号：
```
v1.0.1                                    2026-07-03 18:21:30
```

## 跨平台说明

| 类型 | 是否支持跨平台编译 | 原因 |
|------|-------------------|------|
| CLI  | ✅ 支持 | `CGO_ENABLED=0` 静态链接，无系统依赖 |
| GUI  | ❌ 不支持 | Fyne 依赖 OpenGL 和平台特定的 C 库（GLFW、Cocoa、Win32） |

如需 Windows/Linux GUI 版本：
- 在对应平台上运行 `./scripts/build.sh`
- 或在 CI 中按平台分别构建（`.github/workflows/build-gui.yml`）

## 手动编译

```bash
# CLI（可指定 GOOS/GOARCH）
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64   go build -ldflags="-s -w" -o md2word         ./cmd/md2word
CGO_ENABLED=0 GOOS=windows GOARCH=amd64  go build -ldflags="-s -w" -o md2word.exe     ./cmd/md2word
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  go build -ldflags="-s -w" -o md2word         ./cmd/md2word

# GUI（必须 CGO_ENABLED=1 且与构建机平台一致）
go build -ldflags="-s -w -X main.Version=1.0.0" -o md2word-gui ./cmd/md2word-gui
```

`ldflags` 推荐使用 `-s -w` 去除调试信息，体积可减少约 30%。

## 产物大小参考

| 平台 | CLI | GUI |
|------|-----|-----|
| macOS arm64 | ~15M | ~30M |
| Linux amd64 | ~15M | ~30M |
| Windows amd64 | ~15M | ~30M |

## CI/CD

`.github/workflows/build-gui.yml` 在 main/develop 分支和 `v*` tag 触发，会同时构建 3 大平台（windows-latest / macos-latest / ubuntu-latest）的 CLI 和 GUI，并自动创建 GitHub Release。
