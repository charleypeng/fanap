package fan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Fan 风扇接口
type Fan interface {
	SetSpeed(pwm int) error
	GetSpeed() (int, error)
	Close() error
}

// PWMFan PWM风扇控制器
type PWMFan struct {
	pwmPath      string
	enablePath   string
	originalMode int
	verbose      bool
}

// NewPWMFan 创建新的PWM风扇控制器
// deviceName 可以是具体的hwmon路径（如 "/sys/class/hwmon/hwmon0/pwm1"）
// 或设备名称
func NewPWMFan(deviceName string, verbose bool) (*PWMFan, error) {
	// 如果deviceName已经是完整路径
	if filepath.IsAbs(deviceName) {
		if _, err := os.Stat(deviceName); err != nil {
			return nil, fmt.Errorf("PWM设备路径不存在: %w", err)
		}
		return createPWMFan(deviceName, verbose)
	}

	// 自动查找PWM风扇
	pwmPath, err := findPWMDevice(deviceName)
	if err != nil {
		return nil, fmt.Errorf("查找PWM设备失败: %w", err)
	}

	return createPWMFan(pwmPath, verbose)
}

// createPWMFan 创建PWM风扇并设置为手动控制模式
func createPWMFan(pwmPath string, verbose bool) (*PWMFan, error) {
	// 构建enable路径
	enablePath := pwmPath + "_enable"

	fan := &PWMFan{
		pwmPath:    pwmPath,
		enablePath: enablePath,
		verbose:    verbose,
	}

	// 读取原始模式
	data, err := os.ReadFile(enablePath)
	if err != nil {
		return nil, fmt.Errorf("读取风扇模式失败: %w", err)
	}

	fan.originalMode, err = strconv.Atoi(string(data[:len(data)-1]))
	if err != nil {
		return nil, fmt.Errorf("解析风扇模式失败: %w", err)
	}

	if verbose {
		fmt.Printf("原始风扇模式: %d\n", fan.originalMode)
	}

	// 设置为手动控制模式 (1 = manual)
	if err := os.WriteFile(enablePath, []byte("1"), 0644); err != nil {
		return nil, fmt.Errorf("设置风扇为手动模式失败: %w", err)
	}

	if verbose {
		fmt.Println("风扇已设置为手动控制模式")
	}

	return fan, nil
}

// SetSpeed 设置风扇速度（PWM值，0-255）
func (f *PWMFan) SetSpeed(pwm int) error {
	if pwm < 0 || pwm > 255 {
		return fmt.Errorf("PWM值必须在0-255之间")
	}

	pwmStr := strconv.Itoa(pwm) + "\n"
	if err := os.WriteFile(f.pwmPath, []byte(pwmStr), 0644); err != nil {
		return fmt.Errorf("设置风扇速度失败: %w", err)
	}

	if f.verbose {
		fmt.Printf("设置风扇速度: PWM=%d\n", pwm)
	}

	return nil
}

// GetSpeed 获取当前风扇速度（PWM值）
func (f *PWMFan) GetSpeed() (int, error) {
	data, err := os.ReadFile(f.pwmPath)
	if err != nil {
		return 0, fmt.Errorf("读取风扇速度失败: %w", err)
	}

	pwm, err := strconv.Atoi(string(data[:len(data)-1]))
	if err != nil {
		return 0, fmt.Errorf("解析PWM值失败: %w", err)
	}

	return pwm, nil
}

// Close 关闭风扇控制器，恢复原始模式
func (f *PWMFan) Close() error {
	// 恢复原始模式
	modeStr := strconv.Itoa(f.originalMode) + "\n"
	if err := os.WriteFile(f.enablePath, []byte(modeStr), 0644); err != nil {
		return fmt.Errorf("恢复风扇模式失败: %w", err)
	}

	if f.verbose {
		fmt.Printf("已恢复风扇模式: %d\n", f.originalMode)
	}

	return nil
}

// findPWMDevice 查找PWM风扇设备
func findPWMDevice(deviceName string) (string, error) {
	hwmonPath := "/sys/class/hwmon"

	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		return "", fmt.Errorf("读取hwmon目录失败: %w", err)
	}

	log.Printf("查找PWM风扇设备，检查 %d 个hwmon设备", len(entries))

	var availablePWMs []string

	// 遍历所有hwmon设备
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		devicePath := filepath.Join(hwmonPath, entry.Name(), "name")
		deviceNameStr := "unknown"
		if data, err := os.ReadFile(devicePath); err == nil {
			deviceNameStr = strings.TrimSpace(string(data))
		}

		log.Printf("检查设备: %s (名称: %s)", entry.Name(), deviceNameStr)

		// 查找PWM输出
		for i := 1; i <= 10; i++ {
			pwmPath := filepath.Join(hwmonPath, entry.Name(), fmt.Sprintf("pwm%d", i))
			if _, err := os.Stat(pwmPath); err == nil {
				// 查找风扇标签
				labelPath := filepath.Join(hwmonPath, entry.Name(), fmt.Sprintf("fan%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				pwmInfo := fmt.Sprintf("%s/pwm%d (%s - %s)", entry.Name(), i, deviceNameStr, label)
				availablePWMs = append(availablePWMs, pwmInfo)
				log.Printf("  找到PWM: %s", pwmInfo)

				// 匹配风扇相关的设备名称
				if isFanDevice(deviceNameStr) {
					log.Printf("  ✓ 选择此PWM设备")
					return pwmPath, nil
				}
			}
		}
	}

	// 如果没有找到明确的风扇设备，列出所有可用PWM
	if len(availablePWMs) == 0 {
		return "", fmt.Errorf("未找到任何PWM风扇设备")
	}

	log.Printf("未找到明确的风扇设备，可用的PWM:")
	for i, pwm := range availablePWMs {
		log.Printf("  %d. %s", i+1, pwm)
	}

	return "", fmt.Errorf("未找到可用的PWM风扇设备，请使用-pwm参数指定完整路径，例如: /sys/class/hwmon/hwmon0/pwm1")
}

// isFanDevice 判断是否为风扇设备
func isFanDevice(deviceName string) bool {
	fanPatterns := []string{
		`(?i)fan`,
		`(?i)pwm`,
		`(?i)asus`,
		`(?i)nct6775`,
		`(?i)it87`,
	}

	for _, pattern := range fanPatterns {
		if regexp.MustCompile(pattern).MatchString(deviceName) {
			return true
		}
	}
	return false
}

// findPWMOutput 查找PWM输出文件
func findPWMOutput(hwmonPath string) string {
	for i := 1; i <= 10; i++ {
		pwmPath := filepath.Join(hwmonPath, fmt.Sprintf("pwm%d", i))
		if _, err := os.Stat(pwmPath); err == nil {
			return pwmPath
		}
	}
	return ""
}
