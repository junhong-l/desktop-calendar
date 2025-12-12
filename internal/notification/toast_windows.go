//go:build windows

package notification

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/go-toast/toast"
)

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	procPlaySoundW = winmm.NewProc("PlaySoundW")
)

const (
	SND_FILENAME  = 0x00020000
	SND_ASYNC     = 0x00000001
	SND_SYNC      = 0x00000000
	SND_NODEFAULT = 0x00000002
)

// ShowWindowsNotification 显示 Windows 原生 Toast 通知
func ShowWindowsNotification(title, message string, playSound bool, soundFile string) error {
	notification := toast.Notification{
		AppID:   "待办日历",
		Title:   title,
		Message: message,
	}

	// 处理声音播放
	if playSound {
		if soundFile == "" || soundFile == "default" {
			// 使用嵌入的默认声音（如果有）
			if defaultPath := GetDefaultSoundPath(); defaultPath != "" {
				notification.Audio = toast.Silent
				go PlaySoundFileAsync(defaultPath)
			} else {
				// 没有嵌入的默认声音，使用系统默认
				notification.Audio = toast.Default
			}
		} else {
			// 使用自定义声音文件
			notification.Audio = toast.Silent
			go PlaySoundFileAsync(soundFile)
		}
	} else {
		notification.Audio = toast.Silent
	}

	return notification.Push()
}

// PlaySoundFile 播放指定的声音文件（同步，用于预览）
func PlaySoundFile(soundPath string) error {
	// 处理默认声音
	if soundPath == "" || soundPath == "default" {
		if defaultPath := GetDefaultSoundPath(); defaultPath != "" {
			soundPath = defaultPath
		} else {
			return nil // 没有嵌入的默认声音
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(soundPath); os.IsNotExist(err) {
		return fmt.Errorf("sound file not found: %s", soundPath)
	}

	// 将路径转换为 UTF-16
	pathPtr, err := syscall.UTF16PtrFromString(soundPath)
	if err != nil {
		return err
	}

	// 在 goroutine 中播放，但立即返回成功（避免阻塞UI）
	go func() {
		procPlaySoundW.Call(
			uintptr(unsafe.Pointer(pathPtr)),
			0,
			SND_FILENAME|SND_SYNC|SND_NODEFAULT,
		)
	}()

	return nil
}

// PlaySoundFileAsync 异步播放指定的声音文件（用于通知）
func PlaySoundFileAsync(soundPath string) error {
	if soundPath == "" {
		return nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(soundPath); os.IsNotExist(err) {
		return err
	}

	// 将路径转换为 UTF-16
	pathPtr, err := syscall.UTF16PtrFromString(soundPath)
	if err != nil {
		return err
	}

	// 调用 Windows PlaySound API - 异步播放
	ret, _, _ := procPlaySoundW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		SND_FILENAME|SND_ASYNC|SND_NODEFAULT,
	)

	if ret == 0 {
		return syscall.GetLastError()
	}

	return nil
}

// PreviewSound 预览声音
func PreviewSound(soundPath string) error {
	if soundPath == "" || soundPath == "default" {
		// 优先使用嵌入的默认声音
		if defaultPath := GetDefaultSoundPath(); defaultPath != "" {
			return PlaySoundFile(defaultPath)
		}
		// 没有嵌入的声音，播放系统默认提示音
		return PlaySystemSound()
	}
	return PlaySoundFile(soundPath)
}

// PlaySystemSound 播放系统默认提示音
func PlaySystemSound() error {
	// 首先尝试使用嵌入的默认声音
	if defaultPath := GetDefaultSoundPath(); defaultPath != "" {
		return PlaySoundFile(defaultPath)
	}

	// 如果没有嵌入的声音，使用 Windows 系统声音
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}

	// 尝试播放常见的系统提示音
	soundPaths := []string{
		filepath.Join(systemRoot, "Media", "Windows Notify.wav"),
		filepath.Join(systemRoot, "Media", "notify.wav"),
		filepath.Join(systemRoot, "Media", "Windows Notify System Generic.wav"),
		filepath.Join(systemRoot, "Media", "chimes.wav"),
	}

	for _, path := range soundPaths {
		if _, err := os.Stat(path); err == nil {
			return PlaySoundFile(path)
		}
	}

	// 如果找不到系统声音，使用蜂鸣声
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBeep := user32.NewProc("MessageBeep")
	messageBeep.Call(0x00000040) // MB_ICONINFORMATION

	return nil
}

// PlayNotificationSound 播放通知声音（使用设置中配置的声音）
func PlayNotificationSound(soundFile string) {
	if soundFile != "" && soundFile != "default" {
		PlaySoundFile(soundFile)
	} else {
		PlaySystemSound()
	}
}
