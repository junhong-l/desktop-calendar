package notification

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SoundInfo 声音信息
type SoundInfo struct {
	Name     string `json:"name"`     // 显示名称
	Path     string `json:"path"`     // 文件路径
	IsCustom bool   `json:"isCustom"` // 是否自定义声音
	IsSystem bool   `json:"isSystem"` // 是否系统内置声音
}

// GetSoundsDir 获取声音文件目录
func GetSoundsDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	soundsDir := filepath.Join(filepath.Dir(exe), "data", "sounds")

	// 确保目录存在
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		return "", err
	}

	return soundsDir, nil
}

// getWindowsSystemSounds 获取 Windows 系统内置声音
func getWindowsSystemSounds() []SoundInfo {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}
	mediaDir := filepath.Join(systemRoot, "Media")

	// Windows 常用提示音
	systemSounds := []struct {
		name string
		file string
	}{
		{"Windows 通知", "Windows Notify.wav"},
		{"Windows 提示", "Windows Notify System Generic.wav"},
		{"Windows 铃声", "Ring01.wav"},
		{"Windows 闹钟", "Alarm01.wav"},
		{"Windows 日历", "Windows Notify Calendar.wav"},
		{"Windows 邮件", "Windows Notify Email.wav"},
		{"Windows 消息", "Windows Notify Messaging.wav"},
		{"Chimes", "chimes.wav"},
		{"Chord", "chord.wav"},
		{"Ding", "ding.wav"},
		{"Notify", "notify.wav"},
		{"Tada", "tada.wav"},
		{"Windows Ding", "Windows Ding.wav"},
		{"Windows Exclamation", "Windows Exclamation.wav"},
	}

	sounds := []SoundInfo{}
	for _, s := range systemSounds {
		path := filepath.Join(mediaDir, s.file)
		if _, err := os.Stat(path); err == nil {
			sounds = append(sounds, SoundInfo{
				Name:     s.name,
				Path:     path,
				IsCustom: false,
				IsSystem: true,
			})
		}
	}

	return sounds
}

// GetAvailableSounds 获取可用的声音列表
func GetAvailableSounds() ([]SoundInfo, error) {
	sounds := []SoundInfo{}

	// 添加内置默认选项
	defaultName := "系统默认"
	if HasEmbeddedDefaultSound() {
		defaultName = "内置提示音"
	}
	sounds = append(sounds, SoundInfo{
		Name:     defaultName,
		Path:     "default",
		IsCustom: false,
		IsSystem: false,
	})

	// 添加 Windows 系统内置声音
	systemSounds := getWindowsSystemSounds()
	sounds = append(sounds, systemSounds...)

	// 获取自定义声音目录
	soundsDir, err := GetSoundsDir()
	if err != nil {
		return sounds, nil // 返回默认选项
	}

	// 扫描目录中的声音文件
	files, err := os.ReadDir(soundsDir)
	if err != nil {
		return sounds, nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".wav" {
			name := strings.TrimSuffix(file.Name(), ext)
			sounds = append(sounds, SoundInfo{
				Name:     name,
				Path:     filepath.Join(soundsDir, file.Name()),
				IsCustom: true,
				IsSystem: false,
			})
		}
	}

	return sounds, nil
}

// ImportSound 导入自定义声音
func ImportSound(srcPath string) (string, error) {
	soundsDir, err := GetSoundsDir()
	if err != nil {
		return "", fmt.Errorf("无法获取声音目录: %w", err)
	}

	// 获取文件名
	fileName := filepath.Base(srcPath)
	ext := strings.ToLower(filepath.Ext(fileName))

	// 检查文件格式 - PlaySound API 只支持 WAV 格式
	if ext != ".wav" {
		return "", fmt.Errorf("仅支持 WAV 格式音频文件，MP3/OGG/M4A 需要转换为 WAV")
	}

	// 目标路径
	destPath := filepath.Join(soundsDir, fileName)

	// 如果文件已存在，添加数字后缀
	if _, err := os.Stat(destPath); err == nil {
		base := strings.TrimSuffix(fileName, ext)
		for i := 1; ; i++ {
			newName := fmt.Sprintf("%s_%d%s", base, i, ext)
			destPath = filepath.Join(soundsDir, newName)
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break
			}
		}
	}

	// 复制文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("无法打开源文件: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("无法创建目标文件: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("复制文件失败: %w", err)
	}

	return destPath, nil
}

// DeleteSound 删除自定义声音
func DeleteSound(path string) error {
	soundsDir, err := GetSoundsDir()
	if err != nil {
		return err
	}

	// 确保只能删除 sounds 目录下的文件
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absPath, soundsDir) {
		return fmt.Errorf("无法删除非自定义声音")
	}

	return os.Remove(absPath)
}
