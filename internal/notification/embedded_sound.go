//go:build windows

package notification

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed sounds/*.wav
var embeddedSounds embed.FS

// GetEmbeddedSoundPath 获取嵌入的声音文件路径
// 由于 PlaySound API 需要文件路径，我们需要先将嵌入的声音提取到临时目录
func GetEmbeddedSoundPath(name string) (string, error) {
	// 尝试读取嵌入的声音
	data, err := embeddedSounds.ReadFile("sounds/" + name)
	if err != nil {
		return "", err
	}

	// 获取临时目录
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	tempDir := filepath.Join(filepath.Dir(exe), "data", "cache")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// 写入临时文件
	tempPath := filepath.Join(tempDir, name)

	// 检查文件是否已存在且大小相同
	if info, err := os.Stat(tempPath); err == nil && info.Size() == int64(len(data)) {
		return tempPath, nil
	}

	// 写入文件
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return "", err
	}

	return tempPath, nil
}

// GetDefaultSoundPath 获取默认提示音路径
func GetDefaultSoundPath() string {
	// 尝试获取嵌入的默认声音
	path, err := GetEmbeddedSoundPath("default.wav")
	if err == nil {
		return path
	}

	// 如果没有嵌入的声音，返回空（使用系统默认）
	return ""
}

// HasEmbeddedDefaultSound 检查是否有嵌入的默认声音
func HasEmbeddedDefaultSound() bool {
	_, err := embeddedSounds.ReadFile("sounds/default.wav")
	return err == nil
}
