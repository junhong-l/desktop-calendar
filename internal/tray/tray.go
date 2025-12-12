package tray

import (
	"context"
	"os"

	"todo-calendar/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TrayManager 托盘管理器
type TrayManager struct {
	ctx      context.Context
	isHidden bool
}

// NewTrayManager 创建托盘管理器
func NewTrayManager() *TrayManager {
	return &TrayManager{
		isHidden: false,
	}
}

// SetContext 设置上下文
func (t *TrayManager) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// StartTray 启动系统托盘
func (t *TrayManager) StartTray() {
	// 使用 Windows API 在单独文件中实现
	go startSystemTray(t)
}

// MinimizeToTray 最小化到托盘
func (t *TrayManager) MinimizeToTray() {
	if t.ctx != nil {
		runtime.WindowHide(t.ctx)
		t.isHidden = true
	}
}

// ShowWindow 显示窗口
func (t *TrayManager) ShowWindow() {
	if t.ctx != nil {
		runtime.WindowShow(t.ctx)
		runtime.WindowSetAlwaysOnTop(t.ctx, true)
		runtime.WindowSetAlwaysOnTop(t.ctx, false)
		t.isHidden = false
	}
}

// ToggleWindow 切换窗口显示状态
func (t *TrayManager) ToggleWindow() {
	if t.isHidden {
		t.ShowWindow()
	} else {
		t.MinimizeToTray()
	}
}

// Quit 退出应用
func (t *TrayManager) Quit() {
	// 先关闭小插件
	utils.CloseWindow("待办提醒")
	// 移除托盘图标
	RemoveTrayIcon()
	// 退出主程序
	if t.ctx != nil {
		runtime.Quit(t.ctx)
	}
	// 确保进程退出
	os.Exit(0)
}

// IsHidden 是否隐藏
func (t *TrayManager) IsHidden() bool {
	return t.isHidden
}
