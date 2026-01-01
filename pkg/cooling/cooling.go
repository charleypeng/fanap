package cooling

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Device 冷却设备接口
type Device interface {
	SetLevel(level int) error
	GetLevel() (int, error)
	GetMaxLevel() (int, error)
	Close() error
}

// CoolingDevice 冷却设备
type CoolingDevice struct {
	devicePath string
	typePath   string
	maxState   int
	curState   int
	verbose    bool
}

// NewDevice 创建新的冷却设备
// deviceName 可以是具体的cooling_device路径（如 "/sys/class/thermal/cooling_device4"）
// 或 "auto" 自动检测风扇设备
func NewDevice(deviceName string, verbose bool) (*CoolingDevice, error) {
	var devicePath string

	// 如果deviceName已经是完整路径
	if filepath.IsAbs(deviceName) {
		if _, err := os.Stat(deviceName); err != nil {
			return nil, fmt.Errorf("冷却设备路径不存在: %w", err)
		}
		devicePath = deviceName
	} else {
		// 自动查找风扇类型的cooling device
		var err error
		devicePath, err = findCoolingDevice(verbose)
		if err != nil {
			return nil, fmt.Errorf("查找冷却设备失败: %w", err)
		}
	}

	// 读取设备类型
	typePath := filepath.Join(devicePath, "type")
	deviceType, err := os.ReadFile(typePath)
	if err != nil {
		return nil, fmt.Errorf("读取设备类型失败: %w", err)
	}

	// 检查是否为Fan类型
	if !strings.Contains(strings.ToLower(string(deviceType)), "fan") {
		return nil, fmt.Errorf("设备不是风扇类型: %s", strings.TrimSpace(string(deviceType)))
	}

	// 读取最大状态
	maxStatePath := filepath.Join(devicePath, "max_state")
	maxStateData, err := os.ReadFile(maxStatePath)
	if err != nil {
		return nil, fmt.Errorf("读取最大状态失败: %w", err)
	}

	maxState, err := strconv.Atoi(strings.TrimSpace(string(maxStateData)))
	if err != nil {
		return nil, fmt.Errorf("解析最大状态失败: %w", err)
	}

	if verbose {
		fmt.Printf("冷却设备路径: %s\n", devicePath)
		fmt.Printf("设备类型: %s\n", strings.TrimSpace(string(deviceType)))
		fmt.Printf("最大状态: %d\n", maxState)
	}

	// 读取当前状态
	curStatePath := filepath.Join(devicePath, "cur_state")
	curStateData, err := os.ReadFile(curStatePath)
	curState := 0
	if err == nil {
		curState, _ = strconv.Atoi(strings.TrimSpace(string(curStateData)))
	}

	return &CoolingDevice{
		devicePath: devicePath,
		typePath:   typePath,
		maxState:   maxState,
		curState:   curState,
		verbose:    verbose,
	}, nil
}

// SetLevel 设置冷却级别
func (d *CoolingDevice) SetLevel(level int) error {
	if level < 0 {
		level = 0
	}
	if level > d.maxState {
		level = d.maxState
	}

	curStatePath := filepath.Join(d.devicePath, "cur_state")
	levelStr := strconv.Itoa(level) + "\n"

	if err := os.WriteFile(curStatePath, []byte(levelStr), 0644); err != nil {
		return fmt.Errorf("设置冷却级别失败: %w", err)
	}

	if d.verbose {
		fmt.Printf("设置风扇级别: %d/%d\n", level, d.maxState)
	}

	d.curState = level
	return nil
}

// GetLevel 获取当前冷却级别
func (d *CoolingDevice) GetLevel() (int, error) {
	curStatePath := filepath.Join(d.devicePath, "cur_state")

	data, err := os.ReadFile(curStatePath)
	if err != nil {
		return 0, fmt.Errorf("读取冷却级别失败: %w", err)
	}

	level, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("解析冷却级别失败: %w", err)
	}

	return level, nil
}

// GetMaxLevel 获取最大冷却级别
func (d *CoolingDevice) GetMaxLevel() (int, error) {
	return d.maxState, nil
}

// Close 关闭冷却设备
func (d *CoolingDevice) Close() error {
	// 冷却设备无需特殊关闭
	return nil
}

// findCoolingDevice 查找风扇类型的冷却设备
func findCoolingDevice(verbose bool) (string, error) {
	thermalPath := "/sys/class/thermal"

	entries, err := os.ReadDir(thermalPath)
	if err != nil {
		return "", fmt.Errorf("读取thermal目录失败: %w", err)
	}

	// 查找所有cooling_device
	var coolingDevices []string

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cooling_device") {
			coolingDevices = append(coolingDevices, entry.Name())
		}
	}

	if len(coolingDevices) == 0 {
		return "", fmt.Errorf("未找到任何冷却设备")
	}

	if verbose {
		fmt.Printf("找到 %d 个冷却设备\n", len(coolingDevices))
	}

	// 优先查找Fan类型的设备
	for _, deviceName := range coolingDevices {
		devicePath := filepath.Join(thermalPath, deviceName)

		typePath := filepath.Join(devicePath, "type")
		typeData, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}

		deviceType := strings.ToLower(strings.TrimSpace(string(typeData)))

		if verbose {
			fmt.Printf("检查 %s: 类型=%s\n", deviceName, deviceType)
		}

		// 查找Fan类型的设备
		if strings.Contains(deviceType, "fan") {
			if verbose {
				fmt.Printf("找到风扇设备: %s\n", devicePath)
			}
			return devicePath, nil
		}
	}

	// 如果没有找到Fan类型，返回第一个冷却设备
	if len(coolingDevices) > 0 {
		if verbose {
			fmt.Printf("警告: 未找到Fan类型的设备，使用第一个冷却设备: %s\n",
				filepath.Join(thermalPath, coolingDevices[0]))
		}
		return filepath.Join(thermalPath, coolingDevices[0]), nil
	}

	return "", fmt.Errorf("未找到可用的风扇设备")
}
