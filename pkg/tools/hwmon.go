package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckHWMon 检查hwmon设备（诊断模式）
func CheckHWMon() {
	hwmonPath := "/sys/class/hwmon"

	fmt.Println("=== HWMon设备诊断 ===")
	fmt.Println()

	// 1. 检查hwmon目录
	fmt.Println("1. 检查hwmon目录:")
	if _, err := os.Stat(hwmonPath); err != nil {
		fmt.Printf("   ✗ hwmon目录不存在: %s\n", hwmonPath)
		fmt.Println("   建议: 确保内核支持hwmon，或加载相关内核模块")
		return
	}
	fmt.Printf("   ✓ hwmon目录存在\n")

	// 2. 读取hwmon目录
	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		fmt.Printf("   ✗ 读取hwmon目录失败: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("   ✗ hwmon目录为空")
		fmt.Println("   建议: 加载内核模块")
		fmt.Println("   - Intel CPU: sudo modprobe coretemp")
		fmt.Println("   - AMD CPU: sudo modprobe k10temp")
		return
	}

	fmt.Printf("   ✓ 找到 %d 个hwmon设备\n", len(entries))
	fmt.Println()

	// 3. 检查每个设备
	fmt.Println("2. 检查每个hwmon设备:")
	deviceCount := 0
	sensorCount := 0
	pwmCount := 0

	for _, entry := range entries {
		// hwmon设备可能是符号链接，需要检查实际路径
		fullPath := filepath.Join(hwmonPath, entry.Name())

		// 检查是否为目录（包括符号链接指向的目录）
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		// 如果是符号链接，获取目标信息
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(fullPath)
			if err != nil {
				continue
			}
			// 解析相对路径
			if !filepath.IsAbs(target) {
				fullPath = filepath.Join(hwmonPath, target)
			} else {
				fullPath = target
			}
			info, err = os.Stat(fullPath)
			if err != nil || !info.IsDir() {
				continue
			}
		} else if !info.IsDir() {
			// 不是目录也不是符号链接，跳过
			continue
		}

		deviceCount++
		deviceDir := fullPath

		// 读取设备名称
		namePath := filepath.Join(deviceDir, "name")
		deviceName := "unknown"
		if data, err := os.ReadFile(namePath); err == nil {
			deviceName = strings.TrimSpace(string(data))
		}

		fmt.Printf("   设备 %d: %s\n", deviceCount, entry.Name())
		fmt.Printf("     名称: %s\n", deviceName)

		// 检查温度传感器
		deviceSensorCount := 0
		for i := 1; i <= 10; i++ {
			tempInput := filepath.Join(deviceDir, fmt.Sprintf("temp%d_input", i))
			if _, err := os.Stat(tempInput); err == nil {
				deviceSensorCount++
				sensorCount++

				// 读取温度值
				data, err := os.ReadFile(tempInput)
				temp := "N/A"
				if err == nil {
					tempRaw := strings.TrimSpace(string(data))
					if tempVal, err := parseTemp(tempRaw); err == nil {
						temp = fmt.Sprintf("%.1f°C", tempVal)
					}
				}

				// 读取标签
				labelPath := filepath.Join(deviceDir, fmt.Sprintf("temp%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				// 判断是否为CPU传感器
				isCPU := isCPUSensor(deviceName, label)
				cpuMark := ""
				if isCPU {
					cpuMark = " [CPU]"
				}

				fmt.Printf("     ✓ 温度%d: %s (%s%s) [%s]\n", i, tempInput, label, cpuMark, temp)
			}
		}

		if deviceSensorCount == 0 {
			fmt.Printf("     ✗ 未找到温度传感器\n")
		}

		// 检查PWM风扇
		devicePWMCount := 0
		for i := 1; i <= 10; i++ {
			pwmPath := filepath.Join(deviceDir, fmt.Sprintf("pwm%d", i))
			if _, err := os.Stat(pwmPath); err == nil {
				devicePWMCount++
				pwmCount++

				// 读取PWM值
				data, err := os.ReadFile(pwmPath)
				pwm := "N/A"
				if err == nil {
					pwmRaw := strings.TrimSpace(string(data))
					if pwmVal, err := parseInt(pwmRaw); err == nil {
						pwm = fmt.Sprintf("%d", pwmVal)
					}
				}

				// 读取标签
				labelPath := filepath.Join(deviceDir, fmt.Sprintf("fan%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				fmt.Printf("     ✓ PWM%d: %s (%s) [PWM=%s]\n", i, pwmPath, label, pwm)
			}
		}

		if devicePWMCount == 0 {
			fmt.Printf("     ✗ 未找到PWM风扇\n")
		}

		fmt.Println()
	}

	// 4. 总结
	fmt.Println("3. 总结:")
	fmt.Printf("   设备数量: %d\n", deviceCount)
	fmt.Printf("   温度传感器: %d\n", sensorCount)
	fmt.Printf("   PWM风扇: %d\n", pwmCount)
	fmt.Println()

	// 5. 建议
	fmt.Println("4. 建议:")
	if sensorCount == 0 && pwmCount == 0 {
		fmt.Println("   ✗ 未找到任何温度传感器或PWM风扇")
		fmt.Println("   可能原因:")
		fmt.Println("   1. 内核模块未加载")
		fmt.Println("      - Intel CPU: sudo modprobe coretemp")
		fmt.Println("      - AMD CPU: sudo modprobe k10temp")
		fmt.Println("   2. 设备需要特定的内核驱动程序")
		fmt.Println("   3. BIOS中未启用硬件监控")
	} else if sensorCount > 0 && pwmCount == 0 {
		fmt.Println("   ⚠ 找到温度传感器但未找到PWM风扇")
		fmt.Println("   可能原因:")
		fmt.Println("   1. 系统使用4针风扇（无PWM控制）")
		fmt.Println("   2. 风扇由BIOS或主板独立控制")
		fmt.Println("   3. 需要特定的驱动程序")
		fmt.Println()
		fmt.Println("   建议命令:")
		fmt.Println("   sudo fanap -sensor <温度传感器路径> -pwm auto -verbose")
	} else if sensorCount == 0 && pwmCount > 0 {
		fmt.Println("   ⚠ 找到PWM风扇但未找到温度传感器")
		fmt.Println("   可能原因:")
		fmt.Println("   1. 温度传感器在其他位置")
		fmt.Println("   2. 需要加载温度传感器驱动")
		fmt.Println()
		fmt.Println("   建议命令:")
		fmt.Println("   sudo fanap -sensor auto -pwm <PWM路径> -verbose")
	} else {
		fmt.Println("   ✓ 系统配置正常，可以使用自动检测模式")
		fmt.Println()
		fmt.Println("   建议命令:")
		fmt.Println("   sudo fanap -verbose")
	}

	fmt.Println()

	// 6. 检查权限
	fmt.Println("5. 权限检查:")
	uid := os.Getuid()
	if uid == 0 {
		fmt.Println("   ✓ 当前用户: root")
	} else {
		fmt.Printf("   ✗ 当前用户: UID=%d (非root)\n", uid)
		fmt.Println("   建议使用 sudo 运行程序")
	}
}

// ListHWMon 列出所有可用的硬件监控设备
func ListHWMon() {
	hwmonPath := "/sys/class/hwmon"

	if _, err := os.Stat(hwmonPath); err != nil {
		fmt.Printf("错误: hwmon目录不存在: %s\n", hwmonPath)
		fmt.Println("\n可能的原因:")
		fmt.Println("1. 需要root权限运行此程序")
		fmt.Println("2. hwmon硬件监控接口不支持")
		os.Exit(1)
	}

	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		fmt.Printf("错误: 读取hwmon目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 可用的硬件监控设备 ===\n")

	if len(entries) == 0 {
		fmt.Println("警告: hwmon目录为空")
		fmt.Println("\n建议:")
		fmt.Println("1. 确保使用root权限运行")
		fmt.Println("2. 加载内核模块:")
		fmt.Println("   - Intel CPU: sudo modprobe coretemp")
		fmt.Println("   - AMD CPU: sudo modprobe k10temp")
		fmt.Println("3. 检查BIOS设置")
		return
	}

	fmt.Printf("找到 %d 个hwmon设备\n\n", len(entries))

	sensorCount := 0
	pwmCount := 0

	for _, entry := range entries {
		// hwmon设备可能是符号链接，需要检查实际路径
		fullPath := filepath.Join(hwmonPath, entry.Name())

		// 检查是否为目录（包括符号链接指向的目录）
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		// 如果是符号链接，获取目标信息
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(fullPath)
			if err != nil {
				continue
			}
			// 解析相对路径
			if !filepath.IsAbs(target) {
				fullPath = filepath.Join(hwmonPath, target)
			} else {
				fullPath = target
			}
		}

		devicePath := filepath.Join(fullPath, "name")
		deviceName := "unknown"
		if data, err := os.ReadFile(devicePath); err == nil {
			deviceName = strings.TrimSpace(string(data))
		} else {
			// name文件可能不存在，这很正常
			deviceName = "unknown"
		}

		fmt.Printf("设备: %s\n", entry.Name())
		fmt.Printf("  名称: %s\n", deviceName)

		// 列出温度传感器
		for i := 1; i <= 10; i++ {
			tempInput := filepath.Join(fullPath, fmt.Sprintf("temp%d_input", i))
			if _, err := os.Stat(tempInput); err == nil {
				labelPath := filepath.Join(fullPath, fmt.Sprintf("temp%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				// 读取当前温度
				data, err := os.ReadFile(tempInput)
				temp := "N/A"
				if err == nil {
					tempRaw := strings.TrimSpace(string(data))
					if tempVal, err := parseTemp(tempRaw); err == nil {
						temp = fmt.Sprintf("%.1f°C", tempVal)
					}
				}

				// 判断是否为CPU传感器
				isCPU := isCPUSensor(deviceName, label)
				cpuMark := ""
				if isCPU {
					cpuMark = " [CPU]"
				}

				fmt.Printf("  温度%d: %s (%s%s) [%s]\n", i, tempInput, label, cpuMark, temp)
				sensorCount++
			}
		}

		// 列出PWM风扇
		for i := 1; i <= 10; i++ {
			pwmPath := filepath.Join(fullPath, fmt.Sprintf("pwm%d", i))
			if _, err := os.Stat(pwmPath); err == nil {
				labelPath := filepath.Join(fullPath, fmt.Sprintf("fan%d_label", i))
				label := ""
				if data, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(data))
				}

				// 读取当前PWM值
				data, err := os.ReadFile(pwmPath)
				pwm := "N/A"
				if err == nil {
					pwmRaw := strings.TrimSpace(string(data))
					if pwmVal, err := parseInt(pwmRaw); err == nil {
						pwm = fmt.Sprintf("%d", pwmVal)
					}
				}

				// 判断是否为风扇设备
				isFan := isFanDevice(deviceName)
				fanMark := ""
				if isFan {
					fanMark = " [风扇]"
				}

				fmt.Printf("  风扇%d: %s (%s%s) [PWM=%s]\n", i, pwmPath, label, fanMark, pwm)
				pwmCount++
			}
		}

		fmt.Println()
	}

	if sensorCount == 0 && pwmCount == 0 {
		fmt.Println("警告: 未找到任何温度传感器或PWM风扇")
		fmt.Println("\n可能的原因:")
		fmt.Println("1. 需要root权限运行此程序")
		fmt.Println("2. 内核模块未加载")
		fmt.Println("   - Intel CPU: sudo modprobe coretemp")
		fmt.Println("   - AMD CPU: sudo modprobe k10temp")
		fmt.Println("3. 系统不支持hwmon硬件监控")
		fmt.Println("4. 设备需要特定的内核驱动程序")
	} else {
		fmt.Printf("共找到 %d 个温度传感器，%d 个PWM风扇\n", sensorCount, pwmCount)
		fmt.Println("\n使用方法:")
		if sensorCount > 0 {
			fmt.Println("  # 使用自动检测（推荐）")
			fmt.Println("  sudo fanap -verbose")
			fmt.Println()
			fmt.Println("  # 手动指定传感器")
			fmt.Println("  sudo fanap -sensor <传感器路径> -pwm <风扇路径> -verbose")
			fmt.Println()
		} else if pwmCount > 0 {
			fmt.Println("  找到PWM风扇但未找到温度传感器")
			fmt.Println("  手动指定传感器:")
			fmt.Println("  sudo fanap -sensor <传感器路径> -pwm <风扇路径> -verbose")
			fmt.Println()
		}
		fmt.Println("示例:")
		if sensorCount > 0 && pwmCount > 0 {
			fmt.Println("  sudo fanap -sensor /sys/class/hwmon/hwmon0/temp1_input \\")
			fmt.Println("              -pwm /sys/class/hwmon/hwmon0/pwm1 \\")
			fmt.Println("              -verbose")
		} else if sensorCount > 0 {
			fmt.Println("  sudo fanap -sensor /sys/class/hwmon/hwmon0/temp1_input \\")
			fmt.Println("              -pwm auto \\")
			fmt.Println("              -verbose")
		}
	}
}

// isCPUSensor 判断是否为CPU温度传感器
func isCPUSensor(deviceName, label string) bool {
	// 检查设备名称
	cpuDevicePatterns := []string{
		`(?i)cpu`,
		`(?i)coretemp`,
		`(?i)k10temp`,   // AMD处理器
		`(?i)nct6775`,   // 常见主板传感器
		`(?i)asus`,      // ASUS主板
		`(?i)it87`,      // IT87芯片
		`(?i)acpi`,      // ACPI接口
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

func parseTemp(s string) (float64, error) {
	val := strings.TrimSpace(s)
	var temp int
	_, err := fmt.Sscanf(val, "%d", &temp)
	if err != nil {
		return 0, err
	}
	return float64(temp) / 1000.0, nil
}

func parseInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}
