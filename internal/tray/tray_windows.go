//go:build windows

package tray

import (
	_ "embed"
	"runtime"
	"syscall"
	"unsafe"
)

//go:embed icon.ico
var iconData []byte

var (
	shell32               = syscall.NewLazyDLL("shell32.dll")
	user32                = syscall.NewLazyDLL("user32.dll")
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procShellNotifyIcon   = shell32.NewProc("Shell_NotifyIconW")
	procCreatePopupMenu   = user32.NewProc("CreatePopupMenu")
	procAppendMenu        = user32.NewProc("AppendMenuW")
	procTrackPopupMenu    = user32.NewProc("TrackPopupMenu")
	procDestroyMenu       = user32.NewProc("DestroyMenu")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procSetForegroundWin  = user32.NewProc("SetForegroundWindow")
	procCreateWindowEx    = user32.NewProc("CreateWindowExW")
	procDefWindowProc     = user32.NewProc("DefWindowProcW")
	procRegisterClassEx   = user32.NewProc("RegisterClassExW")
	procGetMessage        = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessage   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procLoadImage         = user32.NewProc("LoadImageW")
	procCreateIconFromRes = user32.NewProc("CreateIconFromResourceEx")
	procLookupIconIdFromD = user32.NewProc("LookupIconIdFromDirectoryEx")
	procGetModuleHandle   = kernel32.NewProc("GetModuleHandleW")
)

const (
	NIM_ADD              = 0x00000000
	NIM_MODIFY           = 0x00000001
	NIM_DELETE           = 0x00000002
	NIM_SETVERSION       = 0x00000004
	NIF_MESSAGE          = 0x00000001
	NIF_ICON             = 0x00000002
	NIF_TIP              = 0x00000004
	NOTIFYICON_VERSION_4 = 4
	WM_USER              = 0x0400
	WM_TRAYICON          = WM_USER + 1
	WM_LBUTTONDBLCLK     = 0x0203
	WM_RBUTTONUP         = 0x0205
	WM_CONTEXTMENU       = 0x007B
	WM_COMMAND           = 0x0111
	NIN_SELECT           = WM_USER + 0
	NIN_KEYSELECT        = WM_USER + 1
	MF_STRING            = 0x00000000
	MF_SEPARATOR         = 0x00000800
	TPM_RIGHTBUTTON      = 0x0002
	TPM_RETURNCMD        = 0x0100
	IDI_APPLICATION      = 32512
	IMAGE_ICON           = 1
	LR_DEFAULTSIZE       = 0x0040
	LR_SHARED            = 0x8000

	IDM_SHOW = 1001
	IDM_QUIT = 1002
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type POINT struct {
	X, Y int32
}

var (
	trayHwnd    uintptr
	trayManager *TrayManager
	trayIcon    uintptr
	nid         NOTIFYICONDATA
)

// createIconFromData 从 ico 文件数据创建图标
func createIconFromData(data []byte) uintptr {
	if len(data) < 6 {
		return 0
	}

	// ICO 文件格式：
	// 0-1: 保留，必须为0
	// 2-3: 图像类型，1=ICO
	// 4-5: 图像数量
	imageCount := int(data[4]) | int(data[5])<<8
	if imageCount == 0 {
		return 0
	}

	// 读取第一个图像目录条目（从偏移6开始，每个条目16字节）
	// 偏移12-15是图像数据在文件中的偏移
	if len(data) < 22 {
		return 0
	}

	offset := int(data[18]) | int(data[19])<<8 | int(data[20])<<16 | int(data[21])<<24
	size := int(data[14]) | int(data[15])<<8 | int(data[16])<<16 | int(data[17])<<24

	if offset+size > len(data) {
		return 0
	}

	// 从资源数据创建图标
	icon, _, _ := procCreateIconFromRes.Call(
		uintptr(unsafe.Pointer(&data[offset])),
		uintptr(size),
		1,       // TRUE = icon
		0x30000, // 版本
		0, 0,    // 使用默认尺寸
		0, // 标志
	)

	return icon
}

func startSystemTray(tm *TrayManager) {
	// 锁定 OS 线程，确保消息循环在同一线程运行
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trayManager = tm

	// 从嵌入的 ico 数据创建图标
	trayIcon = createIconFromData(iconData)
	if trayIcon == 0 {
		// 如果自定义图标加载失败，使用系统默认图标
		hInstance, _, _ := procGetModuleHandle.Call(0)
		trayIcon, _, _ = procLoadImage.Call(
			hInstance,
			uintptr(IDI_APPLICATION),
			IMAGE_ICON,
			16, 16,
			LR_SHARED,
		)
	}

	// 注册窗口类
	className := syscall.StringToUTF16Ptr("TrayWindowClass")

	wndClass := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		LpszClassName: className,
	}

	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wndClass)))

	// 创建隐藏窗口
	trayHwnd, _, _ = procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("TrayWindow"))),
		0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	)

	// 创建托盘图标
	nid = NOTIFYICONDATA{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:             trayHwnd,
		UID:              1,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAYICON,
		HIcon:            trayIcon,
		UVersion:         NOTIFYICON_VERSION_4,
	}

	tip := syscall.StringToUTF16("待办日历")
	copy(nid.SzTip[:], tip)

	procShellNotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	// 设置版本以使用新的消息格式
	procShellNotifyIcon.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&nid)))

	// 消息循环
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// 删除托盘图标
	procShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

func trayWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		// 使用 NOTIFYICON_VERSION_4 时，lParam 的低位是通知消息
		notifyMsg := uint32(lParam & 0xFFFF)
		switch notifyMsg {
		case WM_LBUTTONDBLCLK:
			// 双击打开窗口
			if trayManager != nil {
				trayManager.ShowWindow()
			}
		case WM_RBUTTONUP, WM_CONTEXTMENU:
			// 右键显示菜单
			showContextMenu(hwnd)
		case NIN_SELECT, NIN_KEYSELECT:
			// 单击或键盘选择也可以打开窗口（可选）
		}
		return 0
	case WM_COMMAND:
		menuId := int(wParam & 0xFFFF)
		switch menuId {
		case IDM_SHOW:
			if trayManager != nil {
				trayManager.ShowWindow()
			}
		case IDM_QUIT:
			if trayManager != nil {
				trayManager.Quit()
			}
			procPostQuitMessage.Call(0)
		}
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func showContextMenu(hwnd uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()

	procAppendMenu.Call(menu, MF_STRING, IDM_SHOW, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("显示主窗口"))))
	procAppendMenu.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenu.Call(menu, MF_STRING, IDM_QUIT, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("退出"))))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWin.Call(hwnd)
	procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	procDestroyMenu.Call(menu)
}

// RemoveTrayIcon 移除托盘图标
func RemoveTrayIcon() {
	procShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}
