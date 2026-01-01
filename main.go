package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fanap/pkg/controller"
	"github.com/fanap/pkg/tools"
)

const (
	Version = "1.0.2"
)

const (
	// 默认配置
	DefaultInterval   = 5 * time.Second
	DefaultLowTemp    = 40.0
	DefaultHighTemp   = 75.0
	DefaultMinPWM     = 50
	DefaultMaxPWM     = 255
	DefaultTempSensor = "auto"
	DefaultPWMDevice  = "auto"
)

var (
	showHelp    = flag.Bool("help", false, "显示帮助信息")
	showVersion = flag.Bool("version", false, "显示版本信息")
	listSensors = flag.Bool("list", false, "列出所有可用的温度传感器和PWM风扇设备")
	checkHWMon  = flag.Bool("check", false, "检查hwmon设备（诊断模式）")

	// 风扇控制参数
	interval   = flag.Duration("interval", DefaultInterval, "温度检查间隔 (如: 5s, 10s)")
	lowTemp    = flag.Float64("low-temp", DefaultLowTemp, "低温阈值（摄氏度）")
	highTemp   = flag.Float64("high-temp", DefaultHighTemp, "高温阈值（摄氏度）")
	minPWM     = flag.Int("min-pwm", DefaultMinPWM, "最小PWM值 (0-255)")
	maxPWM     = flag.Int("max-pwm", DefaultMaxPWM, "最大PWM值 (0-255)")
	tempSensor = flag.String("sensor", DefaultTempSensor, "温度传感器路径 (auto=自动检测)")
	pwmDevice  = flag.String("pwm", DefaultPWMDevice, "PWM风扇设备路径 (auto=自动检测)")
	verbose    = flag.Bool("verbose", false, "详细输出模式")
)

// getEnvDuration 从环境变量获取时间间隔
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		duration, err := time.ParseDuration(val)
		if err == nil {
			return duration
		}
		log.Printf("警告: 环境变量 %s 的值 %s 无效，使用默认值 %v\n", key, val, defaultValue)
	}
	return defaultValue
}

// getEnvFloat 从环境变量获取浮点数
func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		log.Printf("警告: 环境变量 %s 的值 %s 无效，使用默认值 %.1f\n", key, val, defaultValue)
	}
	return defaultValue
}

// getEnvInt 从环境变量获取整数
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		log.Printf("警告: 环境变量 %s 的值 %s 无效，使用默认值 %d\n", key, val, defaultValue)
	}
	return defaultValue
}

// getEnvString 从环境变量获取字符串
func getEnvString(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvBool 从环境变量获取布尔值
func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
		log.Printf("警告: 环境变量 %s 的值 %s 无效，使用默认值 %v\n", key, val, defaultValue)
	}
	return defaultValue
}

func main() {
	flag.Usage = printHelp
	flag.Parse()

	// 从环境变量读取配置（优先级：命令行 > 环境变量 > 默认值）
	if *interval == DefaultInterval {
		*interval = getEnvDuration("FANAP_INTERVAL", DefaultInterval)
	}
	if *lowTemp == DefaultLowTemp {
		*lowTemp = getEnvFloat("FANAP_LOW_TEMP", DefaultLowTemp)
	}
	if *highTemp == DefaultHighTemp {
		*highTemp = getEnvFloat("FANAP_HIGH_TEMP", DefaultHighTemp)
	}
	if *minPWM == DefaultMinPWM {
		*minPWM = getEnvInt("FANAP_MIN_PWM", DefaultMinPWM)
	}
	if *maxPWM == DefaultMaxPWM {
		*maxPWM = getEnvInt("FANAP_MAX_PWM", DefaultMaxPWM)
	}
	if *tempSensor == DefaultTempSensor {
		*tempSensor = getEnvString("FANAP_SENSOR", DefaultTempSensor)
	}
	if *pwmDevice == DefaultPWMDevice {
		*pwmDevice = getEnvString("FANAP_PWM", DefaultPWMDevice)
	}
	if !*verbose {
		*verbose = getEnvBool("FANAP_VERBOSE", false)
	}

	// 显示配置信息
	log.Println("=== Fanap 配置 ===")
	log.Printf("温度检查间隔: %v", *interval)
	log.Printf("温度阈值: %.1f°C - %.1f°C", *lowTemp, *highTemp)
	log.Printf("PWM范围: %d - %d", *minPWM, *maxPWM)
	log.Printf("温度传感器: %s", *tempSensor)
	log.Printf("PWM设备: %s", *pwmDevice)
	log.Printf("详细日志: %v", *verbose)

	// 处理特殊命令
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *listSensors {
		tools.ListHWMon()
		os.Exit(0)
	}

	if *checkHWMon {
		tools.CheckHWMon()
		os.Exit(0)
	}

	// 运行风扇控制程序
	runFanController()
}

func printVersion() {
	fmt.Printf("Fanap v%s - CPU温度控制风扇程序\n", Version)
	fmt.Println("用Go语言编写，支持Linux系统的hwmon硬件监控接口")
}

func printHelp() {
	fmt.Printf(`Fanap v%s - CPU温度控制风扇程序
=====================================

使用方法:
  fanap [选项]
  fanap -list              列出所有可用的温度传感器和PWM风扇设备
  fanap -check             检查hwmon设备（诊断模式）
  fanap -help              显示帮助信息
  fanap -version           显示版本信息

风扇控制选项:
  -interval duration        温度检查间隔 (默认: 5s)
  -low-temp float           低温阈值，低于此温度使用最小PWM (默认: 40.0)
  -high-temp float          高温阈值，高于此温度使用最大PWM (默认: 75.0)
  -min-pwm int              最小PWM值，0-255 (默认: 50)
  -max-pwm int              最大PWM值，0-255 (默认: 255)
  -sensor string            温度传感器路径 (默认: auto，自动检测)
  -pwm string               PWM风扇设备路径 (默认: auto，自动检测)
  -verbose                  详细输出模式

环境变量 (Docker):
  FANAP_INTERVAL           温度检查间隔 (如: 5s, 10s)
  FANAP_LOW_TEMP           低温阈值 (默认: 40.0)
  FANAP_HIGH_TEMP          高温阈值 (默认: 75.0)
  FANAP_MIN_PWM            最小PWM值，0-255 (默认: 50)
  FANAP_MAX_PWM            最大PWM值，0-255 (默认: 255)
  FANAP_SENSOR             温度传感器路径 (默认: auto)
  FANAP_PWM                PWM风扇设备路径 (默认: auto)
  FANAP_VERBOSE            详细输出模式 (默认: false)

配置优先级:
  1. 命令行参数
  2. 环境变量
  3. 默认值

支持的控制模式:
  - PWM控制 (标准Linux系统)
  - Cooling Device (QNAP等NAS设备)

示例:
  # 直接运行
  sudo fanap -verbose

  # 检查hwmon设备（诊断）
  sudo fanap -check

  # 列出可用的传感器和风扇
  sudo fanap -list

  # 自定义温度阈值
  sudo fanap -low-temp=35 -high-temp=65 -verbose

  # Docker运行
  docker run -d --device=/sys/class/hwmon --device=/sys/class/thermal \
             -e FANAP_VERBOSE=true fanap

工作原理:
  - 温度 <= 低温阈值: 使用最小PWM
  - 温度 >= 高温阈值: 使用最大PWM
  - 温度介于两者: 线性插值计算PWM值

注意:
  - 需要root权限或设备访问权限运行
  - 程序退出时会自动恢复原始风扇控制模式
  - 使用 -verbose 模式测试，确认程序工作正常

`, Version)
}

func runFanController() {
	// 验证参数
	if *lowTemp >= *highTemp {
		log.Fatal("错误: 低温阈值必须小于高温阈值")
	}
	if *minPWM < 0 || *minPWM > 255 {
		log.Fatal("错误: 最小PWM值必须在0-255之间")
	}
	if *maxPWM < 0 || *maxPWM > 255 {
		log.Fatal("错误: 最大PWM值必须在0-255之间")
	}
	if *minPWM >= *maxPWM {
		log.Fatal("错误: 最小PWM值必须小于最大PWM值")
	}

	log.Printf("风扇控制程序启动 v%s", Version)

	// 自动检测并创建控制器
	var ctrl *controller.TempController
	var err error

	// 如果手动指定了传感器和风扇，使用指定的配置
	if *tempSensor != "auto" || *pwmDevice != "auto" {
		log.Printf("使用手动配置: sensor=%s, pwm=%s", *tempSensor, *pwmDevice)
		ctrl, err = controller.NewControllerWithPWM(*pwmDevice, *minPWM, *maxPWM, *lowTemp, *highTemp, *interval, *verbose)
	} else {
		// 自动检测
		log.Println("自动检测温度传感器和风扇控制器")
		ctrl, err = controller.NewController(*lowTemp, *highTemp, *interval, *verbose)
	}

	if err != nil {
		log.Fatalf("初始化控制器失败: %v\n提示: 运行 'sudo fanap -check' 诊断问题", err)
	}
	defer ctrl.Stop()

	// 启动控制器
	if err := ctrl.Start(); err != nil {
		log.Fatalf("启动控制器失败: %v", err)
	}

	log.Println("风扇控制器运行中，按Ctrl+C停止...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("接收到停止信号，正在关闭...")
}
