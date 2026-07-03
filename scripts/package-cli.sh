#!/bin/bash

# md2word 命令行版本跨平台打包脚本（本地构建）

set -e

echo "🚀 命令行版本跨平台打包（本地构建）..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装或不可用"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

# 创建打包目录
rm -rf dist-cli
mkdir -p dist-cli

# 构建信息
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "📦 开始跨平台构建..."
echo "🔧 Go 版本: $(go version)"

# 支持的平台配置
PLATFORMS="windows-amd64:windows:amd64:.exe linux-amd64:linux:amd64: darwin-amd64:darwin:amd64: darwin-arm64:darwin:arm64: linux-arm64:linux:arm64:"

for platform_config in $PLATFORMS; do
    if [ -z "$platform_config" ]; then
        continue
    fi
    
    IFS=':' read -r platform goos goarch ext <<< "$platform_config"
    
    echo "🔨 构建 ${platform} (${goos}/${goarch})..."
    
    output_file="dist-cli/md2word-${platform}${ext}"
    
    # 设置环境变量并构建
    if CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build \
        -ldflags="${LDFLAGS}" \
        -o "${output_file}" \
        ./cmd/md2word; then
        
        file_size=$(du -h "${output_file}" | cut -f1)
        echo "✅ ${platform} 构建成功 (${file_size})"
    else
        echo "❌ ${platform} 构建失败"
    fi
done

# 检查构建结果
echo ""
echo "📊 构建结果检查:"
if [ -d "dist-cli" ]; then
    echo "✅ 构建成功！文件列表:"
    ls -lh dist-cli/
    
    echo ""
    echo "🔍 验证文件:"
    for file in dist-cli/md2word-*; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo "✅ $(basename "$file") (${size})"
        fi
    done
else
    echo "❌ 构建失败，未找到 dist-cli 目录"
    exit 1
fi

# 创建分发包
echo ""
echo "📦 创建分发包..."
cd dist-cli

# 为每个平台创建压缩包
for file in md2word-*; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        
        if [[ "$filename" == *.exe ]]; then
            # Windows 使用 zip
            platform=$(echo "$filename" | sed 's/md2word-//' | sed 's/\.exe$//')
            zip -q "${platform}.zip" "$filename"
            echo "📦 创建 Windows 分发包: ${platform}.zip"
        else
            # Unix 系统使用 tar.gz
            platform=$(echo "$filename" | sed 's/md2word-//')
            tar -czf "${platform}.tar.gz" "$filename"
            echo "📦 创建分发包: ${platform}.tar.gz"
        fi
    fi
done

cd ..

echo ""
echo "🎉 命令行版本跨平台打包完成！"
echo "📁 打包结果位于: dist-cli/"
echo ""
echo "💡 使用说明:"
echo "   Windows: dist-cli/md2word-windows-amd64.exe input.md output.docx"
echo "   Linux:   dist-cli/md2word-linux-amd64 input.md output.docx"
echo "   macOS:   dist-cli/md2word-darwin-amd64 input.md output.docx"
echo ""
echo "📦 分发文件:"
ls -lh dist-cli/*.zip dist-cli/*.tar.gz 2>/dev/null || echo "   (分发包创建中...)"