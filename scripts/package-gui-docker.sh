#!/bin/bash

# md2word GUI 版本 Docker 跨平台编译脚本（使用标准 Go 镜像）

set -e

echo "🐳 使用 Docker 进行 GUI 编译..."

# 检查 Docker 是否可用
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装或不可用"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 创建编译目录
rm -rf dist-gui-docker
mkdir -p dist-gui-docker

# 构建信息
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "📦 开始 Docker 编译..."

# 尝试的镜像列表（按优先级排序）
IMAGES=(
    "golang:latest"
    "golang:1.24-alpine"
    "golang:1.23-alpine"
    "golang:1.22-alpine" 
    "golang:alpine"
)

# 选择可用的镜像
SELECTED_IMAGE=""
for image in "${IMAGES[@]}"; do
    echo "🔍 检查镜像: $image"
    if docker pull "$image" 2>/dev/null; then
        echo "✅ 成功拉取镜像: $image"
        SELECTED_IMAGE="$image"
        break
    else
        echo "❌ 无法拉取镜像: $image"
    fi
done

if [ -z "$SELECTED_IMAGE" ]; then
    echo "❌ 无法拉取任何 Go 镜像，尝试使用本地镜像..."
    # 检查本地是否有可用的 Go 镜像
    LOCAL_IMAGES=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep golang | head -1)
    if [ -n "$LOCAL_IMAGES" ]; then
        SELECTED_IMAGE="$LOCAL_IMAGES"
        echo "✅ 使用本地镜像: $SELECTED_IMAGE"
    else
        echo "❌ 没有可用的 Go 镜像"
        exit 1
    fi
fi

echo "🚀 使用镜像: $SELECTED_IMAGE"

# 检测镜像的架构
CURRENT_ARCH=$(docker run --rm "$SELECTED_IMAGE" go env GOARCH 2>/dev/null || echo "arm64")
echo "🏗️  检测到架构: $CURRENT_ARCH"

# 支持的平台配置（编译当前架构）
PLATFORMS="linux-${CURRENT_ARCH}:linux:${CURRENT_ARCH}"

# 为每个平台编译
for platform_config in $PLATFORMS; do
    if [ -z "$platform_config" ]; then
        continue
    fi
    
    IFS=':' read -r platform goos goarch <<< "$platform_config"
    
    echo ""
    echo "🔨 编译 GUI ${platform} (${goos}/${goarch})..."
    
    # 设置输出文件名
    output_file="md2word-gui-${platform}"
    if [ "$goos" = "windows" ]; then
        output_file="${output_file}.exe"
    fi
    
    # 根据镜像类型选择不同的安装命令
    if [[ "$SELECTED_IMAGE" == *"alpine"* ]]; then
        # Alpine 系统使用 apk
        INSTALL_CMD="apk add --no-cache gcc musl-dev pkgconfig mesa-dev libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxext-dev libxfixes-dev alsa-lib-dev"
    else
        # Debian/Ubuntu 系统使用 apt-get
        INSTALL_CMD="apt-get update -qq && apt-get install -y -qq gcc pkg-config libgl1-mesa-dev xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxext-dev libxfixes-dev libasound2-dev"
    fi
    
    # 使用选定的镜像进行编译（不设置 GOOS/GOARCH，编译当前架构）
    if docker run --rm \
        -v "$(pwd)":/workspace \
        -w /workspace \
        -e CGO_ENABLED=1 \
        "$SELECTED_IMAGE" \
        sh -c "
            $INSTALL_CMD && \
            go mod tidy && \
            go build -ldflags='${LDFLAGS}' -o 'dist-gui-docker/${output_file}' ./cmd/md2word-gui
        "; then
        
        if [ -f "dist-gui-docker/${output_file}" ]; then
            file_size=$(du -h "dist-gui-docker/${output_file}" | cut -f1)
            echo "✅ GUI ${platform} 编译成功 (${file_size})"
        else
            echo "❌ GUI ${platform} 编译失败 - 文件未生成"
        fi
    else
        echo "❌ GUI ${platform} 编译失败"
        
        # 如果编译失败，尝试显示更多调试信息
        echo "🔍 尝试获取更多调试信息..."
        docker run --rm \
            -v "$(pwd)":/workspace \
            -w /workspace \
            "$SELECTED_IMAGE" \
            sh -c "go version && echo 'Go 模块信息:' && go mod download && go list -m all | head -5"
    fi
done

# 检查编译结果
echo ""
echo "📊 编译结果检查:"
if [ -d "dist-gui-docker" ]; then
    echo "✅ 编译完成！文件列表:"
    ls -lh dist-gui-docker/
    
    echo ""
    echo "🔍 验证文件:"
    for file in dist-gui-docker/md2word-gui-*; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo "✅ $(basename "$file") (${size})"
        fi
    done
else
    echo "❌ 编译失败，未找到 dist-gui-docker 目录"
    exit 1
fi

# 创建分发包
echo ""
echo "📦 创建分发包..."
cd dist-gui-docker

# 为每个平台创建压缩包
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
echo "🎉 Docker GUI 编译完成！"
echo "� 编译结果位于: d$ist-gui-docker/"
echo "🖼️  使用的镜像: $SELECTED_IMAGE"
echo "🏗️  编译架构: $CURRENT_ARCH"
echo ""
echo "💡 使用说明:"
echo "   Linux:   dist-gui-docker/md2word-gui-linux-${CURRENT_ARCH}"
echo ""
echo "📦 分发文件:"
ls -lh dist-gui-docker/*.tar.gz dist-gui-docker/*.zip 2>/dev/null || echo "   (分发包创建中...)"

echo ""
echo "⚠️  GUI 版本 Docker 编译说明:"
echo "   Docker 环境下编译当前容器架构的 Linux 版本。"
echo "   Windows 和 macOS 版本由于平台特定依赖，建议在对应平台上编译。"
echo ""
echo "🔍 测试建议:"
echo "   可以尝试运行编译出的二进制文件来验证是否正常工作"