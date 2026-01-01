package temp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Sensor 温度传感器接口
type Sensor interface {
	GetTemperature() (float64, error)
	Close() error
}

// HWSensor 硬件监控传感器
type HWSensor struct {
	path string
}

// NewSensor 创建新的温度传感器
// sensorName 可以是具体的hwmon路径（如 "/sys/class/hwmon/hwmon0/temp1_input"）
// 或传感器名称
func NewSensor(sensorName string) (*HWSensor, error) {
	// 如果sensorName已经是完整路径
	if filepath.IsAbs(sensorName) {
		if _, err := os.Stat(sensorName); err != nil {
			return nil, fmt.Errorf("传感器路径不存在: %w", err)
		}
		return &HWSensor{path: sensorName}, nil
	}

	// 自动查找CPU温度传感器
	path, err := findCpuTempSensor(sensorName)
	if err != nil {
		return nil, fmt.Errorf("查找温度传感器失败: %w", err)
	}

	return &HWSensor{path: path}, nil
}

// GetTemperature 获取当前CPU温度（摄氏度）
func (s *HWSensor) GetTemperature() (float64, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return 0, fmt.Errorf("读取温度失败: %w", err)
	}

	// 温度通常以毫摄氏度存储
	tempRaw, err := strconv.Atoi(string(data[:len(data)-1])) // 去掉换行符
	if err != nil {
		return 0, fmt.Errorf("解析温度值失败: %w", err)
	}

	// 转换为摄氏度
	temp := float64(tempRaw) / 1000.0

	return temp, nil
}

// Close 关闭传感器
func (s *HWSensor) Close() error {
	// 文件系统传感器无需特殊关闭
	return nil
}

// findCpuTempSensor 查找CPU温度传感器
func findCpuTempSensor(sensorName string) (string, error) {
	// 检查hwmon目录是否存在
	hwmonPath := "/sys/class/hwmon"
	if _, err := os.Stat(hwmonPath); err != nil {
		return "", fmt.Errorf("hwmon目录不存在: %s，请确保您的系统支持硬件监控", hwmonPath)
	}

	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		return "", fmt.Errorf("读取hwmon目录失败: %w", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("hwmon目录为空，未找到任何硬件监控设备")
	}

	log.Printf("找到 %d 个hwmon设备", len(entries))

	// 收集所有可用的温度传感器
	var availableSensors []string

	// 遍历所有hwmon设备
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		devicePath := filepath.Join(hwmonPath, entry.Name(), "name")
		deviceName := "unknown"
		if _, err := os.Stat(devicePath); err == nil {
			name, err := os.ReadFile(devicePath)
			if err == nil {
				deviceName = strings.TrimSpace(string(name))
			}
		}

		log.Printf("检查设备: %s (名称: %s)", entry.Name(), deviceName)

		// 查找温度输入
		for i := 1; i <= 10; i++ {
			tempInput := filepath.Join(hwmonPath, entry.Name(), fmt.Sprintf("temp%d_input", i))
			if _, err := os.Stat(tempInput); err == nil {
				// 尝试读取温度标签
				labelPath := filepath.Join(hwmonPath, entry.Name(), fmt.Sprintf("temp%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				sensorInfo := fmt.Sprintf("%s/temp%d_input (%s - %s)", entry.Name(), i, deviceName, label)
				availableSensors = append(availableSensors, sensorInfo)
				log.Printf("  找到温度传感器: %s", sensorInfo)

				// 匹配CPU相关的设备名称
				if isCPUSensor(deviceName, label) {
					log.Printf("  ✓ 选择此传感器作为CPU温度传感器")
					return tempInput, nil
				}
			}
		}
	}

	// 如果没有找到明确的CPU传感器，列出所有可用传感器
	if len(availableSensors) == 0 {
		return "", fmt.Errorf("未找到任何温度传感器")
	}

	log.Printf("未找到明确的CPU温度传感器，可用的温度传感器:")
	for i, sensor := range availableSensors {
		log.Printf("  %d. %s", i+1, sensor)
	}

	return "", fmt.Errorf("未找到明确的CPU温度传感器，请使用-sensor参数指定完整路径，例如: /sys/class/hwmon/hwmon0/temp1_input")
}

// isCPUSensor 判断是否为CPU温度传感器
func isCPUSensor(deviceName, label string) bool {
	// 检查设备名称
	cpuDevicePatterns := []string{
		`(?i)cpu`,
		`(?i)coretemp`,
		`(?i)k10temp`,      // AMD处理器
		`(?i)nct6775`,      // 常见主板传感器
		`(?i)asus`,         // ASUS主板
		`(?i)it87`,         // IT87芯片
		`(?i)acpi`,         // ACPI接口
	}

	for _, pattern := range cpuDevicePatterns {
		if regexp.MustCompile(pattern).MatchString(deviceName) {
			return true
		}
	}

	// 检查标签
	cpuLabelPatterns := []string{
		`(?i)cpu.*temp`,
		`(?i)core.*temp`,
		`(?i)package.*id`,
		`(?i)tccd`,
	}

	for _, pattern := range cpuLabelPatterns {
		if regexp.MustCompile(pattern).MatchString(label) {
			return true
		}
	}

	return false
}

// findTempInput 查找温度输入文件
func findTempInput(hwmonPath string) string {
	for i := 1; i <= 10; i++ {
		tempInput := filepath.Join(hwmonPath, fmt.Sprintf("temp%d_input", i))
		if _, err := os.Stat(tempInput); err == nil {
			return tempInput
		}
	}
	return ""
}
