package controller

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fanap/pkg/cooling"
	"github.com/fanap/pkg/fan"
	"github.com/fanap/pkg/temp"
	"github.com/fanap/pkg/thermal"
)

// Controller 控制器接口
type Controller interface {
	Start() error
	Stop()
}

// TempSensor 温度传感器接口
type TempSensor interface {
	GetTemperature() (float64, error)
	Close() error
}

// FanController 风扇控制器接口
type FanController interface {
	SetSpeed(speed int) error
	GetSpeed() (int, error)
	Close() error
	GetMinSpeed() int
	GetMaxSpeed() int
}

// FanControllerImpl PWM风扇控制器实现
type FanControllerImpl struct {
	fan      *fan.PWMFan
	minPWM   int
	maxPWM   int
	verbose  bool
	lastPWM  int
	mu       sync.Mutex
}

// NewFanController 创建新的PWM风扇控制器
func NewFanController(deviceName string, minPWM, maxPWM int, verbose bool) (*FanControllerImpl, error) {
	pwmFan, err := fan.NewPWMFan(deviceName, verbose)
	if err != nil {
		return nil, err
	}

	return &FanControllerImpl{
		fan:     pwmFan,
		minPWM:  minPWM,
		maxPWM:  maxPWM,
		verbose: verbose,
		lastPWM: 0,
	}, nil
}

// CoolingDeviceController 冷却设备控制器实现
type CoolingDeviceController struct {
	cooling  *cooling.CoolingDevice
	verbose  bool
	lastLevel int
	mu       sync.Mutex
}

// NewCoolingDeviceController 创建新的冷却设备控制器
func NewCoolingDeviceController(verbose bool) (*CoolingDeviceController, error) {
	coolingDevice, err := cooling.NewDevice("auto", verbose)
	if err != nil {
		return nil, err
	}

	return &CoolingDeviceController{
		cooling:  coolingDevice,
		verbose:  verbose,
		lastLevel: 0,
	}, nil
}

// SetSpeed 设置风扇速度（冷却级别）
func (cc *CoolingDeviceController) SetSpeed(speed int) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	maxLevel, err := cc.cooling.GetMaxLevel()
	if err != nil {
		return err
	}

	// 将速度（0-255）映射到冷却级别（0-maxLevel）
	level := (speed * maxLevel) / 255

	// 对于只有2级（开/关）的设备，使用阈值
	if maxLevel == 1 {
		// 如果速度高于最小PWM的50%，打开风扇，否则关闭
		if speed > 127 { // 50% of 255
			level = 1
		} else {
			level = 0
		}

		if cc.verbose {
			fmt.Printf("2级设备: speed=%d -> level=%d\n", speed, level)
		}
	}

	// 避免重复设置相同的值
	if level == cc.lastLevel {
		if cc.verbose && speed != cc.lastLevel {
			fmt.Printf("级别未变化，跳过设置: level=%d\n", level)
		}
		return nil
	}

	if err := cc.cooling.SetLevel(level); err != nil {
		return err
	}

	// 注意：详细的日志由 cooling.SetLevel 内部处理
	cc.lastLevel = level
	return nil
}

// GetSpeed 获取当前风扇速度
func (cc *CoolingDeviceController) GetSpeed() (int, error) {
	level, err := cc.cooling.GetLevel()
	if err != nil {
		return 0, err
	}

	maxLevel, err := cc.cooling.GetMaxLevel()
	if err != nil {
		return 0, err
	}

	// 将冷却级别映射回速度（0-255）
	if maxLevel == 0 {
		return 0, nil
	}
	return (level * 255) / maxLevel, nil
}

// GetMinSpeed 获取最小速度
func (cc *CoolingDeviceController) GetMinSpeed() int {
	return 0
}

// GetMaxSpeed 获取最大速度
func (cc *CoolingDeviceController) GetMaxSpeed() int {
	return 255
}

// Close 关闭冷却设备控制器
func (cc *CoolingDeviceController) Close() error {
	return cc.cooling.Close()
}

// SetSpeed 设置风扇速度
func (fc *FanControllerImpl) SetSpeed(pwm int) error {
	// 限制PWM值范围
	if pwm < fc.minPWM {
		pwm = fc.minPWM
	}
	if pwm > fc.maxPWM {
		pwm = fc.maxPWM
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// 避免重复设置相同的值
	if pwm == fc.lastPWM {
		return nil
	}

	if err := fc.fan.SetSpeed(pwm); err != nil {
		return err
	}

	fc.lastPWM = pwm
	return nil
}

// GetSpeed 获取当前风扇速度
func (fc *FanControllerImpl) GetSpeed() (int, error) {
	return fc.fan.GetSpeed()
}

// Close 关闭风扇控制器
func (fc *FanControllerImpl) Close() error {
	return fc.fan.Close()
}

// GetMinSpeed 获取最小速度
func (fc *FanControllerImpl) GetMinSpeed() int {
	return fc.minPWM
}

// GetMaxSpeed 获取最大速度
func (fc *FanControllerImpl) GetMaxSpeed() int {
	return fc.maxPWM
}

// TempController 温度控制器
type TempController struct {
	sensor   TempSensor
	fan      FanController
	lowTemp  float64
	highTemp float64
	interval time.Duration
	verbose  bool
	stopChan chan struct{}
	running  bool
}

// NewController 创建新的温度控制器（自动检测）
func NewController(lowTemp, highTemp float64, interval time.Duration, verbose bool) (*TempController, error) {
	// 尝试检测温度传感器
	sensor, err := detectSensor()
	if err != nil {
		return nil, fmt.Errorf("检测温度传感器失败: %w", err)
	}

	// 尝试检测风扇控制器
	fanCtrl, err := detectFanController(verbose)
	if err != nil {
		return nil, fmt.Errorf("检测风扇控制器失败: %w", err)
	}

	return &TempController{
		sensor:   sensor,
		fan:      fanCtrl,
		lowTemp:  lowTemp,
		highTemp: highTemp,
		interval: interval,
		verbose:  verbose,
		stopChan: make(chan struct{}),
		running:  false,
	}, nil
}

// NewControllerWithPWM 创建新的温度控制器（指定PWM设备）
func NewControllerWithPWM(pwmDevice string, minPWM, maxPWM int, lowTemp, highTemp float64, interval time.Duration, verbose bool) (*TempController, error) {
	// 使用hwmon温度传感器
	sensor, err := temp.NewSensor("auto")
	if err != nil {
		return nil, fmt.Errorf("初始化温度传感器失败: %w", err)
	}

	// 使用PWM风扇控制器
	fanCtrl, err := NewFanController(pwmDevice, minPWM, maxPWM, verbose)
	if err != nil {
		return nil, fmt.Errorf("初始化风扇控制器失败: %w", err)
	}

	return &TempController{
		sensor:   sensor,
		fan:      fanCtrl,
		lowTemp:  lowTemp,
		highTemp: highTemp,
		interval: interval,
		verbose:  verbose,
		stopChan: make(chan struct{}),
		running:  false,
	}, nil
}

// detectSensor 自动检测温度传感器
func detectSensor() (TempSensor, error) {
	// 优先尝试thermal_zone（如QNAP等设备）
	sensor, err := thermal.NewZone("auto")
	if err == nil {
		log.Println("使用thermal_zone温度传感器")
		return sensor, nil
	}

	// 回退到hwmon温度传感器
	log.Println("thermal_zone不可用，尝试hwmon温度传感器")
	hwmonSensor, err := temp.NewSensor("auto")
	if err != nil {
		return nil, fmt.Errorf("无法找到任何温度传感器")
	}

	log.Println("使用hwmon温度传感器")
	return hwmonSensor, nil
}

// detectFanController 自动检测风扇控制器
func detectFanController(verbose bool) (FanController, error) {
	// 优先尝试cooling_device（如QNAP等设备）
	fanCtrl, err := NewCoolingDeviceController(verbose)
	if err == nil {
		log.Println("使用cooling_device风扇控制器")
		return fanCtrl, nil
	}

	// 回退到PWM风扇控制器
	if verbose {
		log.Println("cooling_device不可用，尝试PWM风扇控制器")
	}
	pwmFanCtrl, err := NewFanController("auto", 50, 255, verbose)
	if err != nil {
		return nil, fmt.Errorf("无法找到任何风扇控制器")
	}

	log.Println("使用PWM风扇控制器")
	return pwmFanCtrl, nil
}

// Start 启动控制器
func (c *TempController) Start() error {
	if c.running {
		return fmt.Errorf("控制器已在运行")
	}

	c.running = true
	go c.controlLoop()

	return nil
}

// Stop 停止控制器
func (c *TempController) Stop() {
	if !c.running {
		return
	}

	close(c.stopChan)
	c.running = false
}

// controlLoop 控制循环
func (c *TempController) controlLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.adjustFanSpeed()
		}
	}
}

// adjustFanSpeed 根据温度调整风扇速度
func (c *TempController) adjustFanSpeed() {
	// 读取当前温度
	temp, err := c.sensor.GetTemperature()
	if err != nil {
		log.Printf("读取温度失败: %v\n", err)
		return
	}

	// 计算目标PWM值
	pwm := c.calculatePWM(temp)

	if c.verbose {
		currentSpeed, _ := c.fan.GetSpeed()
		fmt.Printf("温度: %.1f°C, 当前PWM: %d, 目标PWM: %d\n", temp, currentSpeed, pwm)
	} else {
		log.Printf("温度: %.1f°C, PWM: %d\n", temp, pwm)
	}

	// 设置风扇速度
	if err := c.fan.SetSpeed(pwm); err != nil {
		log.Printf("设置风扇速度失败: %v\n", err)
	}
}

// calculatePWM 根据温度计算PWM值
func (c *TempController) calculatePWM(temp float64) int {
	minPWM := c.fan.GetMinSpeed()
	maxPWM := c.fan.GetMaxSpeed()

	// 温度低于低温阈值，使用最小PWM
	if temp <= c.lowTemp {
		return minPWM
	}

	// 温度高于高温阈值，使用最大PWM
	if temp >= c.highTemp {
		return maxPWM
	}

	// 在低温和高温之间线性插值
	ratio := (temp - c.lowTemp) / (c.highTemp - c.lowTemp)
	pwm := minPWM + int(float64(maxPWM-minPWM)*ratio)

	return pwm
}
