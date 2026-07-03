#!/bin/bash

# md2word Docker 多平台构建脚本

set -e

echo "🐳 使用 Docker 多阶段构建进行跨平台编译..."

# 创建构建目录
mkdir -p dist

# 检查 Docker 是否可用
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装或不可用"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 构建信息
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 支持的平台列表
declare -A PLATFORMS=(
    ["windows-amd64"]="windows amd64"
    ["linux-amd64"]="linux amd64"
    ["darwin-amd64"]="darwin amd64"
    ["darwin-arm64"]="darwin arm64"
)

echo "📦 开始多平台构建..."

# 为每个平台构建
for platform in "${!PLATFORMS[@]}"; do
    IFS=' ' read -r goos goarch <<< "${PLATFORMS[$platform]}"
    
    echo "🔨 构建 ${platform} (${goos}/${goarch})..."
    
    # 构建 Docker 镜像并提取二进制文件
    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg GOOS="${goos}" \
        --build-arg GOARCH="${goarch}" \
        -f Dockerfile.build \
        -t md2word-builder:${platform} \
        .
    
    # 创建临时容器并复制文件
    container_id=$(docker create md2word-builder:${platform})
    
    # 确定文件扩展名
    if [ "${goos}" = "windows" ]; then
        binary_name="md2word-${platform}.exe"
        docker_binary="/md2word-${goos}-${goarch}.exe"
    else
        binary_name="md2word-${platform}"
        docker_binary="/md2word-${goos}-${goarch}"
    fi
    
    # 复制二进制文件
    if docker cp "${container_id}:${docker_binary}" "dist/${binary_name}" 2>/dev/null; then
        echo "✅ ${platform} 构建成功: dist/${binary_name}"
    else
        echo "❌ ${platform} 构建失败"
    fi
    
    # 清理临时容器
    docker rm "${container_id}" >/dev/null 2>&1
    
    # 清理镜像（可选）
    docker rmi md2word-builder:${platform} >/dev/null 2>&1
done

# 构建当前平台的 GUI 版本
echo ""
echo "🖥️  构建 GUI 版本（当前平台）..."
go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o dist/md2word-gui ./cmd/md2word-gui

if [ -f "dist/md2word-gui" ]; then
    echo "✅ GUI 版本构建成功"
else
    echo "❌ GUI 版本构建失败"
fi

echo ""
echo "✅ Docker 多平台构建完成！"
echo "📁 构建结果:"

# 显示构建结果
if [ -d "dist" ]; then
    ls -lh dist/
    
    echo ""
    echo "📊 文件大小统计:"
    du -h dist/* | sort -hr
    
    echo ""
    echo "🔍 文件验证:"
    for file in dist/md2word-*; do
        if [ -f "$file" ]; then
            file_info=$(file "$file" 2>/dev/null || echo "unknown")
            echo "  $(basename "$file"): $file_info"
        fi
    done
else
    echo "❌ 构建失败，未找到 dist 目录"
    exit 1
fi

echo ""
echo "💡 使用说明:"
echo "   Windows: dist/md2word-windows-amd64.exe input.md output.docx"
echo "   Linux:   dist/md2word-linux-amd64 input.md output.docx"
echo "   macOS:   dist/md2word-darwin-amd64 input.md output.docx"
echo "   GUI:     dist/md2word-gui"
echo ""
echo "📦 分发建议:"
echo "   - 这些是静态链接的二进制文件"
echo "   - 可以直接分发给用户，无需依赖"
echo "   - 建议创建压缩包进行分发"

# 创建分发包
echo ""
echo "📦 创建分发包..."
cd dist

# 为每个平台创建压缩包
for file in md2word-*; do
    if [ -f "$file" ] && [ "$file" != "md2word-gui" ]; then
        platform=$(echo "$file" | sed 's/md2word-//' | sed 's/\.exe$//')
        
        if [[ "$file" == *.exe ]]; then
            zip -q "${platform}.zip" "$file"
            echo "✅ 创建 Windows 分发包: ${platform}.zip"
        else
            tar -czf "${platform}.tar.gz" "$file"
            echo "✅ 创建分发包: ${platform}.tar.gz"
        fi
    fi
done

cd ..
echo "🎉 所有构建和打包完成！"