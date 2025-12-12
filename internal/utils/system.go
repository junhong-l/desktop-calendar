package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

const appName = "TodoCalendar"

// EnableAutoStart 启用开机自启
func EnableAutoStart() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	// 获取绝对路径
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer key.Close()

	// 用引号包裹路径以支持空格
	err = key.SetStringValue(appName, `"`+exePath+`"`)
	if err != nil {
		return fmt.Errorf("写入注册表失败: %w", err)
	}

	return nil
}

// DisableAutoStart 禁用开机自启
func DisableAutoStart() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		registry.SET_VALUE,
	)
	if err != nil {
		// 如果键不存在，视为已经禁用
		return nil
	}
	defer key.Close()

	err = key.DeleteValue(appName)
	if err != nil {
		// 如果值不存在，视为已经禁用
		return nil
	}

	return nil
}

// IsAutoStartEnabled 检查是否已启用开机自启
func IsAutoStartEnabled() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	return err == nil
}

// GetAppDataDir 获取应用数据目录
func GetAppDataDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// OpenURL 打开URL
func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// StartProcess 启动新进程
func StartProcess(exePath string, args ...string) *exec.Cmd {
	cmd := exec.Command(exePath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil
	}
	return cmd
}

// GetIPCFilePath 获取 IPC 文件路径
func GetIPCFilePath() string {
	tempDir := os.TempDir()
	return filepath.Join(tempDir, "todo-calendar-ipc.txt")
}

// WriteIPCTodoId 写入待办ID到IPC文件
func WriteIPCTodoId(todoId int64) error {
	ipcPath := GetIPCFilePath()
	return os.WriteFile(ipcPath, []byte(fmt.Sprintf("%d", todoId)), 0644)
}

// ReadIPCTodoId 读取IPC文件中的待办ID并清空
func ReadIPCTodoId() (int64, error) {
	ipcPath := GetIPCFilePath()
	data, err := os.ReadFile(ipcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	// 读取后删除文件
	os.Remove(ipcPath)

	if len(data) == 0 {
		return 0, nil
	}

	todoId, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, nil
	}

	return todoId, nil
}
