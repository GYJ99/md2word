#!/bin/bash

# md2word GUI 版本编译脚本（当前平台）

set -e

echo "🚀 GUI 版本编译（当前平台）..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装或不可用"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

# 创建编译目录
rm -rf dist-gui
mkdir -p dist-gui

# 构建信息
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "📦 开始编译..."
echo "🔧 Go 版本: $(go version)"

# 获取当前平台信息
current_os=$(go env GOOS)
current_arch=$(go env GOARCH)

echo ""
echo "🔨 编译当前平台 GUI (${current_os}/${current_arch})..."

output_file="dist-gui/md2word-gui-${current_os}-${current_arch}"
if [ "$current_os" = "windows" ]; then
    output_file="${output_file}.exe"
fi

# 编译当前平台的 GUI 版本
if go build -ldflags="${LDFLAGS}" -o "${output_file}" ./cmd/md2word-gui; then
    file_size=$(du -h "${output_file}" | cut -f1)
    echo "✅ GUI ${current_os}-${current_arch} 编译成功 (${file_size})"
else
    echo "❌ GUI ${current_os}-${current_arch} 编译失败"
    exit 1
fi

echo ""
echo "⚠️  GUI 版本交叉编译限制说明:"
echo "   由于 Fyne 框架依赖 OpenGL 和平台特定的 C 库，"
echo "   GUI 版本无法进行交叉编译，只能在目标平台上编译。"
echo ""
echo "💡 如需其他平台的 GUI 版本，请在对应平台上运行:"
echo "   Windows: go build -o md2word-gui.exe ./cmd/md2word-gui"
echo "   Linux:   go build -o md2word-gui ./cmd/md2word-gui"
echo "   macOS:   go build -o md2word-gui ./cmd/md2word-gui"

# 检查编译结果
echo ""
echo "📊 编译结果:"
if [ -d "dist-gui" ]; then
    ls -lh dist-gui/
    
    echo ""
    echo "🔍 验证文件:"
    for file in dist-gui/md2word-gui-*; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo "✅ $(basename "$file") (${size})"
        fi
    done
else
    echo "❌ 编译失败，未找到 dist-gui 目录"
    exit 1
fi

# 创建分发包
echo ""
echo "📦 创建分发包..."
cd dist-gui

for file in md2word-gui-*; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        
        if [[ "$filename" == *.exe ]]; then
            # Windows 使用 zip
            platform=$(echo "$filename" | sed 's/md2word-gui-//' | sed 's/\.exe$//')
            zip -q "gui-${platform}.zip" "$filename"
            echo "📦 创建 Windows GUI 分发包: gui-${platform}.zip"
        else
            # Unix 系统使用 tar.gz
            platform=$(echo "$filename" | sed 's/md2word-gui-//')
            tar -czf "gui-${platform}.tar.gz" "$filename"
            echo "📦 创建 GUI 分发包: gui-${platform}.tar.gz"
        fi
    fi
done

cd ..

echo ""
echo "🎉 GUI 版本编译完成！"
echo "📁 编译结果位于: dist-gui/"
echo ""
echo "💡 使用说明:"
echo "   运行: ${output_file}"
echo ""
echo "📦 分发文件:"
ls -lh dist-gui/*.zip dist-gui/*.tar.gz 2>/dev/null || echo "   无分发包"