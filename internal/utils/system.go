package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

const appName = "TodoCalendar"

// getStartupShortcutPath 获取启动文件夹中的快捷方式路径
func getStartupShortcutPath() (string, error) {
	// 获取用户启动文件夹路径: %APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("无法获取APPDATA环境变量")
	}
	startupFolder := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	return filepath.Join(startupFolder, appName+".lnk"), nil
}

// EnableAutoStart 启用开机自启（通过启动文件夹快捷方式）
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

	shortcutPath, err := getStartupShortcutPath()
	if err != nil {
		return err
	}

	// 使用 PowerShell 创建快捷方式，通过环境变量传递路径避免转义问题
	script := `$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut($env:SHORTCUT_PATH); $Shortcut.TargetPath = $env:TARGET_PATH; $Shortcut.WorkingDirectory = $env:WORKING_DIR; $Shortcut.Description = $env:APP_DESC; $Shortcut.Save()`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(),
		"SHORTCUT_PATH="+shortcutPath,
		"TARGET_PATH="+exePath,
		"WORKING_DIR="+filepath.Dir(exePath),
		"APP_DESC=待办日历",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("创建快捷方式失败: %w, output: %s", err, string(output))
	}

	return nil
}

// DisableAutoStart 禁用开机自启
func DisableAutoStart() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	shortcutPath, err := getStartupShortcutPath()
	if err != nil {
		return err
	}

	// 删除快捷方式文件
	err = os.Remove(shortcutPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除快捷方式失败: %w", err)
	}

	return nil
}

// IsAutoStartEnabled 检查是否已启用开机自启
func IsAutoStartEnabled() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	shortcutPath, err := getStartupShortcutPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(shortcutPath)
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
