#!/bin/bash

# Docker快速启动脚本

set -e

echo "=== Fanap Docker 快速启动 ==="

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装，请先安装Docker"
    echo "访问: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "Docker版本: $(docker --version)"

# 构建镜像
echo ""
echo "步骤1: 构建Docker镜像"
docker build -t fanap:latest .

if [ $? -ne 0 ]; then
    echo "错误: Docker镜像构建失败"
    exit 1
fi

echo "✓ Docker镜像构建成功"

# 选择配置模式
echo ""
echo "步骤2: 选择配置模式"
echo "1) 默认配置 (40°C-75°C)"
echo "2) 激进散热 (35°C-60°C)"
echo "3) 静音模式 (50°C-80°C)"
echo "4) 自定义配置"
read -p "请选择 [1-4]: " choice

case $choice in
    1)
        LOW_TEMP="40.0"
        HIGH_TEMP="75.0"
        MIN_PWM="50"
        MAX_PWM="255"
        ;;
    2)
        LOW_TEMP="35.0"
        HIGH_TEMP="60.0"
        MIN_PWM="0"
        MAX_PWM="255"
        ;;
    3)
        LOW_TEMP="50.0"
        HIGH_TEMP="80.0"
        MIN_PWM="30"
        MAX_PWM="200"
        ;;
    4)
        read -p "低温阈值 (如: 40.0): " LOW_TEMP
        read -p "高温阈值 (如: 75.0): " HIGH_TEMP
        read -p "最小PWM (0-255, 如: 50): " MIN_PWM
        read -p "最大PWM (0-255, 如: 255): " MAX_PWM
        ;;
    *)
        echo "无效选择，使用默认配置"
        LOW_TEMP="40.0"
        HIGH_TEMP="75.0"
        MIN_PWM="50"
        MAX_PWM="255"
        ;;
esac

# 选择是否启用详细日志
echo ""
read -p "启用详细日志？ [y/N]: " verbose_choice
if [[ $verbose_choice =~ ^[Yy]$ ]]; then
    VERBOSE="true"
else
    VERBOSE="false"
fi

# 显示配置
echo ""
echo "=== 配置信息 ==="
echo "温度阈值: ${LOW_TEMP}°C - ${HIGH_TEMP}°C"
echo "PWM范围: ${MIN_PWM} - ${MAX_PWM}"
echo "详细日志: ${VERBOSE}"

# 停止旧容器（如果存在）
echo ""
echo "步骤3: 停止旧容器（如果存在）"
if docker ps -a --format '{{.Names}}' | grep -q "^fanap$"; then
    echo "停止现有fanap容器..."
    docker stop fanap 2>/dev/null
    docker rm fanap 2>/dev/null
    echo "✓ 旧容器已清理"
else
    echo "没有运行中的fanap容器"
fi

# 运行新容器
echo ""
echo "步骤4: 启动新容器"
docker run -d \
  --name fanap \
  --restart unless-stopped \
  --device=/sys/class/hwmon:/sys/class/hwmon \
  --device=/sys/class/thermal:/sys/class/thermal \
  -e FANAP_LOW_TEMP=$LOW_TEMP \
  -e FANAP_HIGH_TEMP=$HIGH_TEMP \
  -e FANAP_MIN_PWM=$MIN_PWM \
  -e FANAP_MAX_PWM=$MAX_PWM \
  -e FANAP_VERBOSE=$VERBOSE \
  -e FANAP_INTERVAL=5s \
  fanap:latest

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Fanap容器启动成功!"
    echo ""
    echo "管理命令:"
    echo "  查看日志: docker logs -f fanap"
    echo "  停止容器: docker stop fanap"
    echo "  重启容器: docker restart fanap"
    echo "  删除容器: docker rm fanap"
    echo ""
    echo "配置可以随时通过重新运行此脚本调整"
else
    echo "错误: Fanap容器启动失败"
    exit 1
fi
