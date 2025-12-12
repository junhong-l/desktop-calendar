package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"todo-calendar/internal/app"
	"todo-calendar/internal/database"
	"todo-calendar/internal/notification"
	"todo-calendar/internal/tray"
	"todo-calendar/internal/utils"
)

//go:embed all:frontend/dist
var assets embed.FS

// 全局变量，用于传递待办ID
var openTodoId int64 = 0

// 窗口模式: "main", "widget", "notification"
var windowMode string = "main"

// GetWindowMode 返回当前窗口模式（供前端调用）
func GetWindowMode() string {
	return windowMode
}

func main() {
	// 解析命令行参数
	widgetMode := flag.Bool("widget", false, "启动桌面小部件模式")
	notifyMode := flag.Bool("notify", false, "启动通知弹窗模式")
	notifyTitle := flag.String("notify-title", "", "通知标题")
	notifyMessage := flag.String("notify-message", "", "通知消息")
	notifyType := flag.String("notify-type", "提醒", "通知类型")
	notifyTodoId := flag.Int64("notify-todo", 0, "关联的待办ID")
	todoId := flag.Int64("todo", 0, "打开指定待办的详情")
	flag.Parse()

	// 保存待办ID
	openTodoId = *todoId

	// 初始化数据库
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建应用实例
	application := app.NewApp(db)

	// 根据模式启动不同的窗口
	if *widgetMode {
		runWidgetWindow(application)
	} else if *notifyMode {
		runNotificationPopup(application, *notifyTitle, *notifyMessage, *notifyType, *notifyTodoId)
	} else {
		runMainWindow(application, db)
	}
}

// runWidgetWindow 启动小部件窗口
func runWidgetWindow(application *app.App) {
	// 创建窗口模式服务
	windowModeService := app.NewWindowModeService("widget")

	err := wails.Run(&options.App{
		Title:           "待办提醒",
		Width:           340,
		Height:          500,
		MinWidth:        340,
		MinHeight:       300,
		MaxWidth:        340,
		MaxHeight:       600,
		DisableResize:   false,
		Frameless:       true,
		AlwaysOnTop:     false,
		StartHidden:     false,
		CSSDragProperty: "--wails-draggable",
		CSSDragValue:    "drag",
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 0},
		OnStartup: func(ctx context.Context) {
			application.SetContext(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			// 小部件就绪后导航到widget页面
			runtime.WindowExecJS(ctx, `window.location.hash = '#/widget'`)
			// 将窗口移动到右上角并设置到桌面底层
			go func() {
				// 等待窗口完全加载
				time.Sleep(500 * time.Millisecond)
				utils.MoveWindowToTopRight("待办提醒")
				utils.SetWindowToDesktopLevel("待办提醒")
				utils.SetWindowAsWidget("待办提醒")
			}()
		},
		Bind: []interface{}{
			application,
			windowModeService,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 true,
			IsZoomControlEnabled:              false,
			ZoomFactor:                        1.0,
			DisablePinchZoom:                  true,
			DisableFramelessWindowDecorations: false,
		},
	})

	if err != nil {
		log.Fatal("Widget Error:", err)
	}
}

// 保存通知弹窗的参数
var popupTitle, popupMessage, popupType string
var popupTodoId int64

// runNotificationPopup 启动通知弹窗窗口
func runNotificationPopup(application *app.App, title, message, notifyType string, todoId int64) {
	popupTitle = title
	popupMessage = message
	popupType = notifyType
	popupTodoId = todoId

	// 创建窗口模式服务
	windowModeService := app.NewWindowModeService("notification")

	err := wails.Run(&options.App{
		Title:         "待办通知",
		Width:         340,
		Height:        160,
		DisableResize: true,
		Frameless:     true,
		AlwaysOnTop:   true,
		StartHidden:   false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 102, G: 126, B: 234, A: 255},
		OnStartup: func(ctx context.Context) {
			application.SetContext(ctx)
			// 在启动时就设置路由
			go func() {
				time.Sleep(100 * time.Millisecond)
				runtime.WindowExecJS(ctx, `
					if (window.location.hash !== '#/notification-popup') {
						window.location.hash = '#/notification-popup';
					}
				`)
			}()
		},
		OnDomReady: func(ctx context.Context) {
			// 再次确保导航到通知弹窗页面
			runtime.WindowExecJS(ctx, `window.location.hash = '#/notification-popup'`)
			// 发送通知数据到前端
			go func() {
				time.Sleep(500 * time.Millisecond)
				// 再次确保路由正确
				runtime.WindowExecJS(ctx, `window.location.hash = '#/notification-popup'`)
				time.Sleep(200 * time.Millisecond)
				runtime.EventsEmit(ctx, "notification:show", map[string]interface{}{
					"title":   popupTitle,
					"message": popupMessage,
					"type":    popupType,
					"todoId":  popupTodoId,
				})
				// 定位到右下角并置顶
				utils.MoveWindowToBottomRight("待办通知")
				utils.SetWindowTopmost("待办通知")
			}()

			// 监听关闭事件
			runtime.EventsOn(ctx, "notification:close", func(optionalData ...interface{}) {
				runtime.Quit(ctx)
			})
		},
		Bind: []interface{}{
			application,
			windowModeService,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 true,
			IsZoomControlEnabled:              false,
			DisableFramelessWindowDecorations: true,
		},
	})

	if err != nil {
		log.Fatal("Notification Popup Error:", err)
	}
}

// runMainWindow 启动主窗口
func runMainWindow(application *app.App, db *sql.DB) {
	// 创建通知管理器
	notifier := notification.NewNotifier(db)

	// 创建托盘管理器
	trayManager := tray.NewTrayManager()

	// 创建窗口模式服务
	windowModeService := app.NewWindowModeService("main")

	// 保存上下文引用
	var appCtx context.Context

	// 创建Wails应用
	err := wails.Run(&options.App{
		Title:     "待办日历",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup: func(ctx context.Context) {
			appCtx = ctx
			application.SetContext(ctx)
			notifier.SetContext(ctx)
			trayManager.SetContext(ctx)
			// 启动系统托盘
			trayManager.StartTray()
			// 启动通知检查
			go notifier.StartNotificationChecker()
		},
		OnDomReady: func(ctx context.Context) {
			// DOM就绪后的初始化
			_ = appCtx // 避免未使用变量警告
			// 如果有待办ID参数，发送事件通知前端打开详情
			if openTodoId > 0 {
				runtime.EventsEmit(ctx, "open:todo", strconv.FormatInt(openTodoId, 10))
			}
			// 自动启动桌面小部件
			go func() {
				time.Sleep(1 * time.Second)
				application.OpenWidget()
			}()
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// 最小化到托盘而不是关闭
			if application.GetMinimizeToTray() {
				trayManager.MinimizeToTray()
				return true
			}
			return false
		},
		OnShutdown: func(ctx context.Context) {
			database.CloseDB()
		},
		Bind: []interface{}{
			application,
			notifier,
			windowModeService,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}
