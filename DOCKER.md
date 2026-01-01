# Fanap Docker 部署指南

Fanap支持Docker容器化部署，简化安装和配置。

## 前置要求

- Docker 20.10 或更高版本
- Linux系统（需要hwmon或thermal设备）
- 主机上的root权限或设备访问权限

## 快速开始

### 方法1: 使用Docker Compose（推荐）

1. **克隆或下载程序**

```bash
git clone <repository-url>
cd fanap
```

2. **构建镜像**

```bash
docker-compose build
```

3. **运行容器**

```bash
docker-compose up -d
```

4. **查看日志**

```bash
docker-compose logs -f
```

### 方法2: 使用Docker命令

1. **构建镜像**

```bash
docker build -t fanap:latest .
```

2. **运行容器**

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon:/sys/class/hwmon \
  --device=/sys/class/thermal:/sys/class/thermal \
  -e FANAP_VERBOSE=true \
  fanap:latest
```

## 配置选项

### 环境变量

所有配置都可以通过环境变量设置：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `FANAP_INTERVAL` | 5s | 温度检查间隔 |
| `FANAP_LOW_TEMP` | 40.0 | 低温阈值（摄氏度） |
| `FANAP_HIGH_TEMP` | 75.0 | 高温阈值（摄氏度） |
| `FANAP_MIN_PWM` | 50 | 最小PWM值 (0-255) |
| `FANAP_MAX_PWM` | 255 | 最大PWM值 (0-255) |
| `FANAP_SENSOR` | auto | 温度传感器路径 |
| `FANAP_PWM` | auto | PWM风扇设备路径 |
| `FANAP_VERBOSE` | false | 详细日志输出 |

### 配置优先级

1. **命令行参数** (最高优先级)
2. **环境变量**
3. **默认值** (最低优先级)

### 自定义配置示例

#### 通过环境变量

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  -e FANAP_INTERVAL=10s \
  -e FANAP_LOW_TEMP=35.0 \
  -e FANAP_HIGH_TEMP=65.0 \
  -e FANAP_MIN_PWM=0 \
  -e FANAP_VERBOSE=true \
  fanap:latest
```

#### 通过docker-compose.yml

编辑 `docker-compose.yml` 文件中的 `environment` 部分：

```yaml
environment:
  - FANAP_INTERVAL=10s
  - FANAP_LOW_TEMP=35.0
  - FANAP_HIGH_TEMP=65.0
  - FANAP_MIN_PWM=0
  - FANAP_MAX_PWM=255
  - FANAP_VERBOSE=true
```

然后运行：

```bash
docker-compose up -d
```

## 设备权限

### 必需的设备映射

Docker容器需要访问硬件监控设备：

```bash
--device=/sys/class/hwmon:/sys/class/hwmon
--device=/sys/class/thermal:/sys/class/thermal
```

**注意**：
- `--device` 参数提供了对硬件设备的直接访问
- 容器内的非root用户仍然需要适当的权限
- QNAP等设备通常只需要thermal设备

### 日志持久化

使用卷映射保存日志：

```bash
-v fanap-logs:/var/log/fanap
```

查看日志：

```bash
# 方法1: 使用docker logs
docker logs -f fanap

# 方法2: 直接访问持久化日志
docker exec fanap tail -f /var/log/fanap/fanap.log
```

## 常用命令

### 查看容器状态

```bash
docker ps | grep fanap
docker stats fanap
```

### 查看日志

```bash
# 实时日志
docker logs -f fanap

# 最近100行
docker logs --tail 100 fanap
```

### 进入容器

```bash
docker exec -it fanap sh
```

### 停止容器

```bash
docker stop fanap
```

### 重启容器

```bash
docker restart fanap
```

### 删除容器

```bash
docker stop fanap
docker rm fanap
```

## 生产环境部署

### 设置自动重启

```bash
docker run -d \
  --name fanap \
  --restart unless-stopped \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  -e FANAP_VERBOSE=false \
  fanap:latest
```

### 限制资源使用

在 `docker-compose.yml` 中配置：

```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 128M
    reservations:
      cpus: '0.1'
      memory: 64M
```

### 日志轮转

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 故障排除

### 容器无法启动

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

### 容器启动但无法控制风扇

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

### 权限错误

如果看到权限错误，尝试：

1. **使用特权模式** (不推荐，仅用于测试):

```bash
docker run -d \
  --name fanap \
  --privileged \
  fanap:latest
```

2. **使用--cap-add** (推荐):

```bash
docker run -d \
  --name fanap \
  --cap-add SYS_RAWIO \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest
```

### QNAP设备特有问题

QNAP NAS使用thermal cooling device，确保映射了thermal设备：

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/thermal:/sys/class/thermal \
  -e FANAP_VERBOSE=true \
  fanap:latest
```

如果QNAP只使用thermal设备，可以省略hwmon映射：

```bash
docker run -d \
  --name fanap \
  --device=/sys/class/thermal \
  fanap:latest
```

## 性能监控

### 容器资源使用

```bash
docker stats fanap
```

### 查看容器详情

```bash
docker inspect fanap
```

## 安全建议

1. **最小化权限**：只映射必要的设备
2. **资源限制**：限制CPU和内存使用
3. **日志管理**：配置日志轮转避免磁盘满
4. **定期更新**：保持镜像更新到最新版本
5. **监控日志**：定期检查日志文件大小

## 更新和维护

### 更新容器

```bash
# 停止并删除旧容器
docker stop fanap
docker rm fanap

# 构建新镜像
docker build -t fanap:latest .

# 运行新容器
docker run -d \
  --name fanap \
  --device=/sys/class/hwmon \
  --device=/sys/class/thermal \
  fanap:latest
```

### 使用docker-compose更新

```bash
# 重新构建并启动
docker-compose up -d --build
```

## 卸载

```bash
# 停止并删除容器
docker stop fanap
docker rm fanap

# 删除镜像（可选）
docker rmi fanap:latest

# 删除卷（可选）
docker volume rm fanap-logs
```
