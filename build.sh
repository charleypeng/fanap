#!/bin/bash

# 构建脚本 - 用于在macOS上交叉编译Linux版本

set -e

echo "=== Fanap 构建脚本 ==="

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go未安装，请先安装Go"
    exit 1
fi

echo "Go版本: $(go version)"

# 创建输出目录
mkdir -p build

# 构建Linux x64版本
echo "正在构建Linux x64版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/fanap-linux-amd64 .

# 创建tar包
echo "正在创建发布包..."
cd build
tar -czf fanap-linux-amd64.tar.gz fanap-linux-amd64
cd ..

echo ""
echo "✓ 构建完成!"
echo ""
echo "文件列表:"
echo "  build/fanap-linux-amd64 ($(du -h build/fanap-linux-amd64 | cut -f1))"
echo "  build/fanap-linux-amd64.tar.gz ($(du -h build/fanap-linux-amd64.tar.gz | cut -f1))"
echo ""
echo "部署方式:"
echo ""
echo "1. 直接部署（推荐用于QNAP）:"
echo "   上传到目标系统并运行"
echo ""
echo "2. Docker部署（推荐用于通用Linux）:"
echo "   ./build-docker.sh"
echo ""
echo "详细文档:"
echo "   - 直接部署: README.md"
echo "   - Docker部署: DOCKER.md"

