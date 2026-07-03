#!/bin/bash
# md2word 一键打包脚本：编译当前平台的 CLI 与 GUI
# 输出到 dist-cli/ 和 dist-gui/

set -euo pipefail

VERSION="${VERSION:-1.0.0}"
BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

current_os=$(go env GOOS)
current_arch=$(go env GOARCH)
platform="${current_os}-${current_arch}"

echo "==> 平台: ${platform}"
echo "==> 版本: ${VERSION} (${BUILD_TIME})"

# ---------- CLI ----------
rm -rf dist-cli
mkdir -p dist-cli
cli_ext=""
[ "${current_os}" = "windows" ] && cli_ext=".exe"
cli_out="dist-cli/md2word-${platform}${cli_ext}"

echo ""
echo "==> 编译 CLI: ${cli_out}"
CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o "${cli_out}" ./cmd/md2word
echo "    成功 ($(du -h "${cli_out}" | cut -f1))"

# ---------- GUI ----------
rm -rf dist-gui
mkdir -p dist-gui
gui_out="dist-gui/md2word-gui-${platform}${cli_ext}"

echo ""
echo "==> 编译 GUI: ${gui_out}"
go build -ldflags="${LDFLAGS}" -o "${gui_out}" ./cmd/md2word-gui
echo "    成功 ($(du -h "${gui_out}" | cut -f1))"

# ---------- 分发包 ----------
echo ""
echo "==> 创建分发包..."
cd dist-cli
if [ "${current_os}" = "windows" ]; then
    zip -q "${platform}.zip" "md2word-${platform}${cli_ext}"
else
    tar -czf "${platform}.tar.gz" "md2word-${platform}${cli_ext}"
fi
cd ../dist-gui
if [ "${current_os}" = "windows" ]; then
    zip -q "gui-${platform}.zip" "md2word-gui-${platform}${cli_ext}"
else
    tar -czf "gui-${platform}.tar.gz" "md2word-gui-${platform}${cli_ext}"
fi
cd ..

echo ""
echo "==> 完成。"
echo "    CLI: dist-cli/md2word-${platform}${cli_ext}"
echo "    GUI: dist-gui/md2word-gui-${platform}${cli_ext}"
ls -lh dist-cli/ dist-gui/
