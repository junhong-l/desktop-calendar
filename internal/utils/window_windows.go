//go:build windows

package utils

import (
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procFindWindow          = user32.NewProc("FindWindowW")
	procSetParent           = user32.NewProc("SetParent")
	procFindWindowEx        = user32.NewProc("FindWindowExW")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetClassName        = user32.NewProc("GetClassNameW")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procGetWindowLong       = user32.NewProc("GetWindowLongW")
	procSetWindowLong       = user32.NewProc("SetWindowLongW")
	procSendMessageTimeout  = user32.NewProc("SendMessageTimeoutW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procPostMessage         = user32.NewProc("PostMessageW")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procMoveWindow          = user32.NewProc("MoveWindow")
)

const (
	HWND_BOTTOM      = 1
	HWND_TOP         = 0
	HWND_TOPMOST     = ^uintptr(0)     // -1
	HWND_NOTOPMOST   = ^uintptr(1) + 1 // -2
	SWP_NOSIZE       = 0x0001
	SWP_NOMOVE       = 0x0002
	SWP_NOACTIVATE   = 0x0010
	SWP_SHOWWINDOW   = 0x0040
	SWP_FRAMECHANGED = 0x0020
	GWL_EXSTYLE      = -20
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_NOACTIVATE = 0x08000000
	WS_EX_APPWINDOW  = 0x00040000
	SW_RESTORE       = 9
	SW_SHOW          = 5
	SW_HIDE          = 0
	WM_CLOSE         = 0x0010
	SM_CXSCREEN      = 0
	SM_CYSCREEN      = 1
)

// FindWorkerW 查找桌面WorkerW窗口（用于将窗口嵌入桌面）
func FindWorkerW() uintptr {
	// 查找 Progman 窗口
	progman, _, _ := procFindWindow.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Progman"))),
		0,
	)

	if progman == 0 {
		return 0
	}

	// 发送消息让 Windows 创建 WorkerW
	procSendMessageTimeout.Call(
		progman,
		0x052C, // 特殊消息
		0,
		0,
		0,
		1000,
		0,
	)

	// 查找 WorkerW 窗口
	var workerW uintptr
	callback := syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		// 获取窗口类名
		className := make([]uint16, 256)
		procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&className[0])), 256)

		if syscall.UTF16ToString(className) == "WorkerW" {
			// 检查是否有 SHELLDLL_DefView 子窗口
			shellView, _, _ := procFindWindowEx.Call(hwnd, 0,
				uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("SHELLDLL_DefView"))),
				0,
			)
			if shellView != 0 {
				// 找到包含桌面图标的 WorkerW，我们需要它后面的那个
				workerW, _, _ = procFindWindowEx.Call(0, hwnd,
					uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("WorkerW"))),
					0,
				)
			}
		}
		return 1 // 继续枚举
	})

	procEnumWindows.Call(callback, 0)

	return workerW
}

// SetWindowToDesktopLevel 将窗口设置到桌面层级（在桌面图标下方）
func SetWindowToDesktopLevel(windowTitle string) bool {
	// 查找目标窗口
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 将窗口设置到最底层
	procSetWindowPos.Call(
		hwnd,
		HWND_BOTTOM,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE,
	)

	return true
}

// SetWindowAsWidget 将窗口设置为小部件样式（不在任务栏显示，不获取焦点）
func SetWindowAsWidget(windowTitle string) bool {
	// 查找目标窗口
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 先隐藏窗口
	procShowWindow.Call(hwnd, SW_HIDE)

	// 获取当前扩展样式 (GWL_EXSTYLE = -20, 需要转换为 uintptr)
	exStyle, _, _ := procGetWindowLong.Call(hwnd, uintptr(0xFFFFFFEC)) // -20 as unsigned

	// 移除 APPWINDOW 样式，添加 TOOLWINDOW 样式（不在任务栏显示）
	newExStyle := (exStyle &^ WS_EX_APPWINDOW) | WS_EX_TOOLWINDOW
	procSetWindowLong.Call(hwnd, uintptr(0xFFFFFFEC), newExStyle)

	// 重新显示窗口
	procShowWindow.Call(hwnd, SW_SHOW)

	return true
}

// KeepWindowAtBottom 持续将窗口保持在底层
func KeepWindowAtBottom(windowTitle string) {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return
	}

	procSetWindowPos.Call(
		hwnd,
		HWND_BOTTOM,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE,
	)
}

// BringWindowToFront 将窗口置顶并激活
func BringWindowToFront(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 先显示窗口（如果是隐藏状态）
	procShowWindow.Call(hwnd, SW_SHOW)

	// 恢复窗口（如果最小化了）
	procShowWindow.Call(hwnd, SW_RESTORE)

	// 将窗口置于最前
	procSetWindowPos.Call(
		hwnd,
		HWND_TOP,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW,
	)

	// 激活窗口
	procSetForegroundWindow.Call(hwnd)

	return true
}

// IsWindowRunning 检测指定标题的窗口是否存在
func IsWindowRunning(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)
	return hwnd != 0
}

// CloseWindow 关闭指定标题的窗口
func CloseWindow(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 发送 WM_CLOSE 消息
	procPostMessage.Call(hwnd, WM_CLOSE, 0, 0)
	return true
}

// RECT 结构体
type RECT struct {
	Left, Top, Right, Bottom int32
}

// MoveWindowToTopRight 将窗口移动到屏幕右上角
func MoveWindowToTopRight(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 获取屏幕尺寸
	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)

	// 获取窗口尺寸
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	windowWidth := rect.Right - rect.Left
	windowHeight := rect.Bottom - rect.Top

	// 计算右上角位置（留10像素边距）
	x := int(screenWidth) - int(windowWidth) - 10
	y := 10

	// 移动窗口
	procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(windowWidth), uintptr(windowHeight), 1)
	return true
}

// MoveWindowToBottomRight 将窗口移动到屏幕右下角（用于通知弹窗）
func MoveWindowToBottomRight(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 获取屏幕尺寸
	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	// 获取窗口尺寸
	var rect RECT
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	windowWidth := rect.Right - rect.Left
	windowHeight := rect.Bottom - rect.Top

	// 计算右下角位置（留出任务栏空间，约60像素）
	x := int(screenWidth) - int(windowWidth) - 20
	y := int(screenHeight) - int(windowHeight) - 80

	// 移动窗口
	procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(windowWidth), uintptr(windowHeight), 1)
	return true
}

// SetWindowTopmost 将窗口设置为始终置顶
func SetWindowTopmost(windowTitle string) bool {
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowTitle))),
	)

	if hwnd == 0 {
		return false
	}

	// 设置为置顶
	ret, _, _ := procSetWindowPos.Call(
		hwnd,
		HWND_TOPMOST,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW,
	)

	return ret != 0
}
