#!/bin/bash

# Docker镜像构建脚本

set -e

echo "=== Fanap Docker 构建脚本 ==="

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装，请先安装Docker"
    echo "访问: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "Docker版本: $(docker --version)"

# 构建Linux版本（如果还没有）
if [ ! -f "build/fanap-linux-amd64" ]; then
    echo "正在构建Linux版本..."
    ./build.sh
fi

# 构建Docker镜像
echo "正在构建Docker镜像..."
docker build -t fanap:latest .

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Docker镜像构建成功!"
    echo ""
    echo "镜像信息:"
    docker images | grep fanap
    echo ""
  echo "运行方法:"
  echo "  docker run -d \\"
  echo "    --name fanap \\"
  echo "    --privileged \\"
  echo "    -e FANAP_VERBOSE=true \\"
  echo "    fanap:latest"
  echo ""
  echo "说明: 程序需要 --privileged 模式访问硬件设备"
  echo ""
  echo "或使用docker-compose:"
  echo "  docker-compose up -d"
    echo ""
    echo "查看日志:"
    echo "  docker logs -f fanap"
else
    echo ""
    echo "✗ Docker镜像构建失败"
    exit 1
fi
