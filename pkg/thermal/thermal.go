package thermal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Zone 温度区域接口
type Zone interface {
	GetTemperature() (float64, error)
	Close() error
}

// ThermalZone Thermal温度区域
type ThermalZone struct {
	path string
}

// NewZone 创建新的温度区域
// zoneName 可以是具体的thermal_zone路径（如 "/sys/class/thermal/thermal_zone0"）
// 或 "auto" 自动检测
func NewZone(zoneName string) (*ThermalZone, error) {
	// 如果zoneName已经是完整路径
	if filepath.IsAbs(zoneName) {
		if _, err := os.Stat(zoneName); err != nil {
			return nil, fmt.Errorf("温度区域路径不存在: %w", err)
		}
		return &ThermalZone{path: zoneName}, nil
	}

	// 自动查找CPU温度区域
	path, err := findThermalZone(zoneName)
	if err != nil {
		return nil, fmt.Errorf("查找温度区域失败: %w", err)
	}

	return &ThermalZone{path: path}, nil
}

// GetTemperature 获取当前温度（摄氏度）
func (z *ThermalZone) GetTemperature() (float64, error) {
	tempPath := filepath.Join(z.path, "temp")

	data, err := os.ReadFile(tempPath)
	if err != nil {
		return 0, fmt.Errorf("读取温度失败: %w", err)
	}

	// 温度以毫摄氏度存储
	tempRaw, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("解析温度值失败: %w", err)
	}

	// 转换为摄氏度
	temp := float64(tempRaw) / 1000.0

	return temp, nil
}

// Close 关闭温度区域
func (z *ThermalZone) Close() error {
	// 文件系统传感器无需特殊关闭
	return nil
}

// findThermalZone 查找温度区域
func findThermalZone(zoneName string) (string, error) {
	thermalPath := "/sys/class/thermal"

	entries, err := os.ReadDir(thermalPath)
	if err != nil {
		return "", fmt.Errorf("读取thermal目录失败: %w", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("thermal目录为空，未找到任何温度区域")
	}

	// 优先选择thermal_zone0（通常是CPU）
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "thermal_zone") {
			zonePath := filepath.Join(thermalPath, entry.Name())

			// 检查是否有temp文件
			tempFile := filepath.Join(zonePath, "temp")
			if _, err := os.Stat(tempFile); err == nil {
				// 返回第一个找到的thermal_zone
				return zonePath, nil
			}
		}
	}

	return "", fmt.Errorf("未找到可用的温度区域")
}
