# Fanap - CPU温度控制风扇程序

Fanap是一个用Go语言编写的Linux系统风扇控制程序，可以根据CPU温度自动调整风扇转速。支持直接运行和Docker容器化部署。

## 功能特性

- ✅ 自动检测CPU温度传感器和PWM风扇设备
- ✅ 根据温度线性调节风扇转速
- ✅ 支持自定义温度阈值和PWM范围
- ✅ 程序退出时自动恢复原始风扇控制模式
- ✅ 提供详细的调试信息
- ✅ 内置传感器检测工具（通过 `-list` 参数）
- ✅ 内置诊断工具（通过 `-check` 参数）
- ✅ **Docker容器化支持**
- ✅ **环境变量配置**
- ✅ **支持多种控制模式**（PWM、Cooling Device）
- ✅ 支持systemd服务模式

## 系统要求

### 直接运行
- Linux系统（Debian、Ubuntu、QNAP等）
- 内核支持hwmon硬件监控接口或thermal cooling device
- 需要root权限
- CPU温度传感器和PWM风扇控制器

### Docker运行
- Docker 20.10 或更高版本
- Linux主机（需要hwmon或thermal设备）
- 主机上的root权限或设备访问权限

## 快速开始

### 方式1: Docker快速启动（最简单，推荐）

#### 1. 运行快速启动脚本

```bash
chmod +x docker-quick-start.sh
./docker-quick-start.sh
```

脚本会：
- 自动构建Docker镜像
- 让你选择配置模式
- 自动启动容器
- 提供管理命令

#### 2. 查看日志

```bash
docker logs -f fanap
```

#### 3. 管理容器

```bash
# 查看状态
docker ps | grep fanap

# 停止
docker stop fanap

# 重启
docker restart fanap

# 删除
docker rm fanap
```

### 方式2: Docker手动配置

#### 1. 构建Docker镜像

```bash
./build-docker.sh
```

#### 2. 运行容器

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon:/sys/class/hwmon \
  --device=/sys/class/thermal:/sys/class/thermal \
  -e FANAP_VERBOSE=true \
  fanap:latest
```

#### 3. 查看日志

```bash
docker logs -f fanap
```

详细的Docker部署说明请查看 [DOCKER.md](DOCKER.md)

### 方式3: 直接运行（用于QNAP等）

#### 1. 构建程序

在开发环境上构建Linux版本：

```bash
./build.sh
```

#### 2. 上传到目标系统

```bash
scp build/fanap-linux-amd64 user@qnap:/tmp/
```

#### 3. 在目标系统上运行

```bash
ssh user@qnap
cd /tmp
chmod +x fanap-linux-amd64
sudo ./fanap-linux-amd64 -verbose
```

详细的直接部署说明请继续阅读下面的章节。

## 命令行参数

### 基本命令

| 参数 | 说明 |
|------|------|
| `-check` | 检查hwmon设备（诊断模式） |
| `-list` | 列出所有可用的温度传感器和PWM风扇设备 |
| `-help` | 显示帮助信息 |
| `-version` | 显示版本信息 |

### 风扇控制选项

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-interval` | 5s | 温度检查间隔 |
| `-low-temp` | 40.0 | 低温阈值（摄氏度），低于此温度使用最小PWM |
| `-high-temp` | 75.0 | 高温阈值（摄氏度），高于此温度使用最大PWM |
| `-min-pwm` | 50 | 最小PWM值（0-255） |
| `-max-pwm` | 255 | 最大PWM值（0-255） |
| `-sensor` | auto | 温度传感器路径（auto=自动检测） |
| `-pwm` | auto | PWM风扇设备路径（auto=自动检测） |
| `-verbose` | false | 详细输出模式 |

## 环境变量（Docker）

所有配置都可以通过环境变量设置：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `FANAP_INTERVAL` | 5s | 温度检查间隔 |
| `FANAP_LOW_TEMP` | 40.0 | 低温阈值（摄氏度） |
| `FANAP_HIGH_TEMP` | 75.0 | 高温阈值（摄氏度） |
| `FANAP_MIN_PWM` | 50 | 最小PWM值（0-255） |
| `FANAP_MAX_PWM` | 255 | 最大PWM值（0-255） |
| `FANAP_SENSOR` | auto | 温度传感器路径 |
| `FANAP_PWM` | auto | PWM风扇设备路径 |
| `FANAP_VERBOSE` | false | 详细日志输出 |

### 配置优先级

1. **命令行参数** (最高优先级)
2. **环境变量**
3. **默认值** (最低优先级)

## 使用示例

### Docker运行

#### 查看帮助信息

```bash
docker run --rm fanap:latest -help
```

#### 查看版本信息

```bash
docker run --rm fanap:latest -version
```

#### 诊断hwmon设备

```bash
docker run --rm \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest -check
```

#### 列出所有可用的传感器

```bash
docker run --rm \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest -list
```

#### 基本使用（自动检测）

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest
```

#### 自定义温度阈值

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  -e FANAP_LOW_TEMP=35.0 \
  -e FANAP_HIGH_TEMP=65.0 \
  fanap:latest
```

#### 快速响应模式

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  -e FANAP_INTERVAL=2s \
  -e FANAP_MIN_PWM=100 \
  fanap:latest
```

#### 静音模式

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  -e FANAP_LOW_TEMP=50.0 \
  -e FANAP_HIGH_TEMP=80.0 \
  -e FANAP_MIN_PWM=30 \
  -e FANAP_MAX_PWM=200 \
  fanap:latest
```

#### 使用docker-compose

```yaml
# 修改docker-compose.yml中的环境变量后运行
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 直接运行

#### 查看帮助信息

```bash
sudo ./fanap-linux-amd64 -help
```

#### 查看版本信息

```bash
sudo ./fanap-linux-amd64 -version
```

#### 诊断hwmon设备

```bash
sudo ./fanap-linux-amd64 -check
```

#### 基本使用（自动检测）

```bash
sudo ./fanap-linux-amd64 -verbose
```

#### 自定义温度阈值

```bash
sudo ./fanap-linux-amd64 -verbose -low-temp=45 -high-temp=80
```

#### 快速响应模式

```bash
sudo ./fanap-linux-amd64 -verbose -interval=2s -min-pwm=100
```

#### 静音模式（最低转速）

```bash
sudo ./fanap-linux-amd64 -verbose -low-temp=30 -high-temp=60 -min-pwm=30
```

#### 指定传感器和风扇

```bash
sudo ./fanap-linux-amd64 -verbose \
  -sensor /sys/class/hwmon/hwmon0/temp1_input \
  -pwm /sys/class/hwmon/hwmon0/pwm1
```

## 工作原理

1. **温度检测**：定期读取CPU温度
2. **PWM计算**：
   - 温度 ≤ 低温阈值：使用最小PWM
   - 温度 ≥ 高温阈值：使用最大PWM
   - 温度介于两者：线性插值计算PWM值
3. **风扇控制**：将计算出的PWM值写入PWM设备或Cooling Device
4. **安全退出**：程序退出时恢复原始风扇控制模式

## 支持的控制模式

### PWM控制（标准Linux系统）

- 适用于大多数Linux系统
- 使用标准的 `/sys/class/hwmon/hwmonX/pwmX` 接口
- 支持0-255的PWM值范围

### Cooling Device（QNAP等NAS设备）

- 适用于QNAP等使用thermal cooling device的设备
- 使用 `/sys/class/thermal/cooling_deviceX` 接口
- 自动将PWM值（0-255）映射到设备的冷却级别
- 支持开/关或多级控制

程序会自动检测并选择合适的控制模式。

## 故障排除

### Docker运行

#### 1. 容器无法启动

**检查设备权限**：

```bash
docker logs fanap
# 查看是否有权限错误
```

确保主机上的设备存在：

```bash
ls -la /sys/class/hwmon
ls -la /sys/class/thermal
```

**尝试特权模式**（仅用于测试，不推荐生产环境）：

```bash
docker run -d \
  --name fanap \
  --privileged \
  fanap:latest
```

#### 2. 容器启动但无法控制风扇

**检查日志**：

```bash
docker logs fanap
```

**手动运行诊断**：

```bash
docker run --rm \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest -check
```

#### 3. 权限错误

如果看到权限错误，尝试使用 `--cap-add`：

```bash
docker run -d \
  --name fanap \
  --cap-add SYS_RAWIO \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest
```

#### 4. QNAP设备

QNAP使用thermal cooling device，确保映射了thermal设备：

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/thermal:/sys/class/thermal \
  -e FANAP_VERBOSE=true \
  fanap:latest
```

如果QNAP只使用thermal设备，可以省略hwmon映射。

### 直接运行

#### 1. 找不到温度传感器或PWM风扇

**诊断步骤：**

首先运行诊断模式：

```bash
sudo ./fanap-linux-amd64 -check
```

**常见原因和解决方案：**

1. **需要root权限**
   ```bash
   sudo ./fanap-linux-amd64 -check
   ```

2. **内核模块未加载**
   ```bash
   # Intel CPU
   sudo modprobe coretemp

   # AMD CPU
   sudo modprobe k10temp
   ```

3. **hwmon目录不存在**
   - 检查BIOS设置，启用硬件监控
   - 确保系统支持hwmon

4. **温度传感器在其他位置**
   - 使用 `-check` 命令查看所有设备
   - 手动指定传感器路径

#### 2. QNAP设备

**诊断：**

```bash
sudo ./fanap-linux-amd64 -check
```

**预期输出：**

```
=== HWMon设备诊断 ===

设备: hwmon0
  名称: acpitz
  ✓ 温度1: /sys/devices/virtual/thermal/thermal_zone0/hwmon0/temp1_input ( [CPU]) [44.0°C]
  ✗ 未找到PWM风扇

设备: hwmon1
  名称: coretemp
  ✓ 温度1: /sys/devices/platform/coretemp.0/hwmon/hwmon1/temp1_input (Package id 0 [CPU]) [45.0°C]
  ...

设备: hwmon1
  ✓ 风扇: /sys/class/thermal/cooling_device4 ( [风扇]) [PWM=0]
```

如果只找到温度传感器但没有找到PWM风扇，这是QNAP正常现象。程序会使用cooling_device接口控制风扇。

**推荐配置：**

```bash
# 让风扇更早启动
sudo ./fanap-linux-amd64 -verbose -low-temp=35 -high-temp=60
```

## 温度和风扇建议

### 温度评估

对于NAS设备（QNAP TS-453D mini）：
```
✅ 优秀：< 50°C
✅ 良好：50-60°C
⚠️  偏高：60-75°C
❌ 危险：> 80°C
```

### 推荐配置

#### 平衡模式（推荐）

```bash
# 直接运行
sudo ./fanap-linux-amd64 -low-temp=40 -high-temp=75 -verbose

# Docker
docker run -d \
  -e FANAP_LOW_TEMP=40.0 \
  -e FANAP_HIGH_TEMP=75.0 \
  --device=/sys/class/thermal \
  fanap:latest
```

#### 激进散热模式

```bash
# 让风扇更早启动
sudo ./fanap-linux-amd64 -low-temp=35 -high-temp=60 -verbose

# Docker
docker run -d \
  -e FANAP_LOW_TEMP=35.0 \
  -e FANAP_HIGH_TEMP=60.0 \
  --device=/sys/class/thermal \
  fanap:latest
```

#### 静音模式

```bash
# 牺牲散热换取静音
sudo ./fanap-linux-amd64 -low-temp=50 -high-temp=80 -verbose

# Docker
docker run -d \
  -e FANAP_LOW_TEMP=50.0 \
  -e FANAP_HIGH_TEMP=80.0 \
  --device=/sys/class/thermal \
  fanap:latest
```

## 作为Systemd服务运行

### 1. 创建服务文件

复制 `systemd/fanap.service` 到系统目录：

```bash
sudo cp systemd/fanap.service /etc/systemd/system/
```

### 2. 修改服务参数（如需要）

编辑 `/etc/systemd/system/fanap.service`，修改 `ExecStart` 行的参数。

例如：

```ini
ExecStart=/usr/local/bin/fanap \
    -interval=5s \
    -low-temp=40 \
    -high-temp=75 \
    -min-pwm=50 \
    -max-pwm=255 \
    -sensor=auto \
    -pwm=auto \
    -verbose=false
```

### 3. 启动服务

```bash
sudo systemctl daemon-reload
sudo systemctl enable fanap
sudo systemctl start fanap
```

### 4. 查看状态

```bash
sudo systemctl status fanap
```

### 5. 查看日志

```bash
sudo journalctl -u fanap -f
```

### 6. 停止服务

```bash
sudo systemctl stop fanap
```

### 7. 禁用服务

```bash
sudo systemctl disable fanap
```

## 注意事项

- ⚠️ 此程序会修改风扇控制设置，使用前请了解风险
- ⚠️ 确保系统有足够的散热能力
- ⚠️ 建议先在 `-verbose` 模式下测试，确认程序工作正常
- ⚠️ 程序需要root权限或设备访问权限才能访问硬件监控接口
- ⚠️ 不同的主板和CPU支持的传感器和PWM设备不同
- ⚠️ 程序退出时会自动恢复原始风扇控制模式
- ⚠️ 如果遇到问题，先运行 `-check` 命令诊断
- ⚠️ QNAP等NAS设备使用特殊的cooling device接口，程序会自动适配
- ⚠️ Docker容器化提供了更好的隔离和管理，但需要正确映射设备

## 开发

### 构建本地版本（macOS）

```bash
make build
./build/fanap -help
```

### 构建Linux版本

```bash
make build-linux
```

### 构建Docker镜像

```bash
./build-docker.sh
```

### 格式化代码

```bash
make fmt
```

### 代码检查

```bash
make vet
```

### 运行测试

```bash
make test
```

### 清理构建文件

```bash
make clean
```

## 项目结构

```
fanap/
├── main.go                    # 主程序入口
├── go.mod                     # Go模块文件
├── Makefile                   # 构建脚本
├── build.sh                   # Linux交叉编译脚本
├── build-docker.sh            # Docker镜像构建脚本
├── docker-quick-start.sh      # Docker快速启动脚本
├── Dockerfile                 # Docker镜像定义
├── docker-compose.yml          # Docker Compose配置
├── README.md                  # 项目文档（本文件）
├── DOCKER.md                  # Docker部署详细文档
├── .gitignore                 # Git忽略文件
├── .dockerignore              # Docker构建忽略文件
├── systemd/
│   └── fanap.service          # Systemd服务配置
└── pkg/
    ├── temp/
    │   └── temp.go            # 温度传感器模块
    ├── fan/
    │   └── fan.go             # PWM风扇控制模块
    ├── thermal/
    │   └── thermal.go         # Thermal温度区域模块
    ├── cooling/
    │   └── cooling.go         # Cooling Device控制模块
    ├── controller/
    │   └── controller.go      # 控制器模块
    └── tools/
        └── hwmon.go           # 工具模块（传感器检测和诊断）
```

## 部署方式对比

| 方式 | 适用场景 | 优势 | 劣势 |
|-----|---------|------|------|
| Docker快速启动 | 通用Linux | 最简单、自动化、容器隔离 | 需要Docker |
| Docker手动配置 | 通用Linux | 灵活配置、环境变量管理 | 需要手动配置 |
| 直接运行 | QNAP NAS | 直接访问硬件、无容器开销 | 部署复杂、无容器隔离 |
| Systemd服务 | 生产环境 | 自动启动、系统集成 | 配置相对复杂 |

## 更新日志

### v1.0.0 (2025-12-31)

- 初始版本发布
- 支持CPU温度传感器自动检测（hwmon和thermal_zone）
- 支持PWM风扇控制
- 支持Cooling Device风扇控制（QNAP等NAS）
- 支持自定义温度阈值和PWM范围
- 内置传感器检测工具（-list参数）
- 内置诊断工具（-check参数）
- 支持systemd服务模式
- **支持Docker容器化部署**
- **支持环境变量配置**
- 详细的帮助信息和错误提示
- 自动检测并选择最佳的控制模式

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题或建议，请提交Issue。
