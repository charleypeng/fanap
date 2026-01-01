#!/bin/bash

# 创建模拟的hwmon目录结构用于测试

TEST_DIR="./hwmon-test"

# 清理旧的测试目录
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# 创建hwmon目录
mkdir -p "$TEST_DIR/hwmon0"
mkdir -p "$TEST_DIR/hwmon1"

# 创建hwmon0的设备名称
echo "coretemp-isa-0000" > "$TEST_DIR/hwmon0/name"

# 创建hwmon1的设备名称
echo "acpi-0" > "$TEST_DIR/hwmon1/name"

# 创建hwmon0的温度传感器
echo "45000" > "$TEST_DIR/hwmon0/temp1_input"
echo "Package id 0" > "$TEST_DIR/hwmon0/temp1_label"

echo "42000" > "$TEST_DIR/hwmon0/temp2_input"
echo "Core 0" > "$TEST_DIR/hwmon0/temp2_label"

echo "41000" > "$TEST_DIR/hwmon0/temp3_input"
echo "Core 1" > "$TEST_DIR/hwmon0/temp3_label"

# 创建hwmon0的PWM风扇
echo "120" > "$TEST_DIR/hwmon0/pwm1"
echo "CPU Fan" > "$TEST_DIR/hwmon0/fan1_label"

echo "1" > "$TEST_DIR/hwmon0/pwm1_enable"

# 创建hwmon1的温度传感器
echo "38000" > "$TEST_DIR/hwmon1/temp1_input"
echo "CPU" > "$TEST_DIR/hwmon1/temp1_label"

echo "=== 测试目录创建完成 ==="
echo "测试目录: $TEST_DIR"
echo ""
echo "目录结构:"
ls -la "$TEST_DIR"
echo ""
echo "hwmon0内容:"
ls -la "$TEST_DIR/hwmon0"
echo ""
echo "hwmon1内容:"
ls -la "$TEST_DIR/hwmon1"
