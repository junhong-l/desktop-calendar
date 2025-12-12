package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"todo-calendar/internal/database"
	"todo-calendar/internal/models"
	"todo-calendar/internal/notification"
	"todo-calendar/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	db             *sql.DB
	todoRepo       *database.TodoRepository
	settingsRepo   *database.SettingsRepository
	attachmentRepo *database.AttachmentRepository
}

// NewApp creates app instance
func NewApp(db *sql.DB) *App {
	return &App{
		db:             db,
		todoRepo:       database.NewTodoRepository(db),
		settingsRepo:   database.NewSettingsRepository(db),
		attachmentRepo: database.NewAttachmentRepository(db),
	}
}

// SetContext sets context
func (a *App) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// GetMinimizeToTray gets minimize to tray setting
func (a *App) GetMinimizeToTray() bool {
	settings, err := a.settingsRepo.Get()
	if err != nil {
		return true
	}
	return settings.MinimizeToTray
}

// ==================== Todo API ====================

// CreateTodo creates todo
func (a *App) CreateTodo(todo models.Todo) (int64, error) {
	if todo.Title == "" {
		return 0, fmt.Errorf("title cannot be empty")
	}
	if todo.AdvanceRemind <= 0 {
		todo.AdvanceRemind = 15
	}
	return a.todoRepo.Create(&todo)
}

// UpdateTodo updates todo
func (a *App) UpdateTodo(todo models.Todo) error {
	if todo.ID <= 0 {
		return fmt.Errorf("invalid todo ID")
	}
	return a.todoRepo.Update(&todo)
}

// DeleteTodo deletes todo
func (a *App) DeleteTodo(id int64) error {
	if err := a.attachmentRepo.DeleteByTodoID(id); err != nil {
		return err
	}
	return a.todoRepo.Delete(id)
}

// GetTodo gets single todo
func (a *App) GetTodo(id int64) (*models.Todo, error) {
	return a.todoRepo.GetByID(id)
}

// GetTodoList gets todo list
func (a *App) GetTodoList(filter models.TodoFilter) (*models.TodoListResult, error) {
	return a.todoRepo.List(filter)
}

// GetPendingTodos gets pending todos
func (a *App) GetPendingTodos() ([]models.Todo, error) {
	return a.todoRepo.GetPendingTodos()
}

// GetTodayStartRemindTodos 获取今天需要开始提醒的待办（软件启动时调用）
func (a *App) GetTodayStartRemindTodos() ([]models.Todo, error) {
	return a.todoRepo.GetTodayStartRemindTodos()
}

// MarkStartRemindTriggered 标记开始提醒已触发
func (a *App) MarkStartRemindTriggered(id int64) error {
	return a.todoRepo.MarkStartRemindTriggered(id)
}

// GetWeekTodos gets week todos
func (a *App) GetWeekTodos() (*models.WeekTodos, error) {
	return a.todoRepo.GetWeekTodos()
}

// MarkTodoCompleted marks todo completed
func (a *App) MarkTodoCompleted(id int64, completed bool) error {
	return a.todoRepo.MarkCompleted(id, completed)
}

// GetTodosByDate gets todos by date
func (a *App) GetTodosByDate(dateStr string) ([]models.Todo, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format")
	}
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24*time.Hour - time.Second)
	return a.todoRepo.GetByDateRange(start, end)
}

// GetTodosByMonth gets todos by month
func (a *App) GetTodosByMonth(year, month int) ([]models.Todo, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	return a.todoRepo.GetByDateRange(start, end)
}

// ==================== Calendar API ====================

// GetCalendarMonth gets calendar month view data
func (a *App) GetCalendarMonth(year, month int) ([]models.CalendarDay, error) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)

	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startDate := firstDay.AddDate(0, 0, -(weekday - 1))

	endWeekday := int(lastDay.Weekday())
	if endWeekday == 0 {
		endWeekday = 7
	}
	endDate := lastDay.AddDate(0, 0, 7-endWeekday)

	todos, err := a.todoRepo.GetByDateRange(startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	todoMap := make(map[string][]models.Todo)
	for _, todo := range todos {
		current := todo.StartDate.Time
		for !current.After(todo.EndDate.Time) && !current.After(endDate) {
			if !current.Before(startDate) {
				dateKey := current.Format("2006-01-02")
				todoMap[dateKey] = append(todoMap[dateKey], todo)
			}
			current = current.AddDate(0, 0, 1)
		}
	}

	days := []models.CalendarDay{}
	today := time.Now()
	current := startDate

	for !current.After(endDate) {
		dateKey := current.Format("2006-01-02")
		lunarDate := utils.SolarToLunar(current.Year(), int(current.Month()), current.Day())
		_, weekNum := current.ISOWeek()

		day := models.CalendarDay{
			Date:           dateKey,
			Day:            current.Day(),
			WeekNumber:     weekNum,
			Lunar:          lunarDate,
			IsToday:        current.Year() == today.Year() && current.YearDay() == today.YearDay(),
			IsCurrentMonth: current.Month() == time.Month(month),
			Todos:          todoMap[dateKey],
			TodoCount:      len(todoMap[dateKey]),
		}
		days = append(days, day)
		current = current.AddDate(0, 0, 1)
	}

	return days, nil
}

// GetLunarDate gets lunar date info
func (a *App) GetLunarDate(year, month, day int) models.LunarDate {
	return utils.SolarToLunar(year, month, day)
}

// ConvertLunarToSolar converts lunar to solar
func (a *App) ConvertLunarToSolar(year, month, day int, isLeap bool) (time.Time, error) {
	return utils.LunarToSolar(year, month, day, isLeap)
}

// ==================== Cron API ====================

// ParseCronExpression parses cron expression
func (a *App) ParseCronExpression(expr string) models.CronNextRun {
	return utils.ParseCronExpr(expr)
}

// CalculateEndDate calculates end date
func (a *App) CalculateEndDate(startDateStr string, cronExpr string, remindCount int) (time.Time, error) {
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		startDate, err = time.Parse("2006-01-02T15:04:05", startDateStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid start time format")
		}
	}
	return utils.CalculateEndDateByRemindCount(startDate, cronExpr, remindCount)
}

// CalculateRemindCount calculates remind count
func (a *App) CalculateRemindCount(startDateStr string, cronExpr string, endDateStr string) int {
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		startDate, _ = time.Parse("2006-01-02T15:04:05", startDateStr)
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		endDate, _ = time.Parse("2006-01-02T15:04:05", endDateStr)
	}
	return utils.CalculateRemindCountByEndDate(startDate, cronExpr, endDate)
}

// ==================== Attachment API ====================

// UploadAttachment uploads attachment
func (a *App) UploadAttachment(todoID int64, fileName string, dataBase64 string, mimeType string) (*models.Attachment, error) {
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid file data")
	}
	return a.attachmentRepo.EncryptAndSaveFile(todoID, fileName, data, mimeType)
}

// GetAttachment gets decrypted attachment
func (a *App) GetAttachment(id int64) (string, error) {
	data, err := a.attachmentRepo.DecryptFile(id)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// GetAttachmentInfo gets attachment info
func (a *App) GetAttachmentInfo(id int64) (*models.Attachment, error) {
	return a.attachmentRepo.GetByID(id)
}

// GetTodoAttachments gets todo attachments
func (a *App) GetTodoAttachments(todoID int64) ([]models.Attachment, error) {
	return a.attachmentRepo.GetByTodoID(todoID)
}

// DeleteAttachment deletes attachment
func (a *App) DeleteAttachment(id int64) error {
	return a.attachmentRepo.Delete(id)
}

// DownloadAttachment downloads attachment to user selected location
func (a *App) DownloadAttachment(id int64) (bool, error) {
	// Get attachment info
	attachment, err := a.attachmentRepo.GetByID(id)
	if err != nil {
		return false, err
	}

	// Decrypt file data
	data, err := a.attachmentRepo.DecryptFile(id)
	if err != nil {
		return false, err
	}

	// Open save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: attachment.FileName,
		Title:           "保存附件",
	})
	if err != nil {
		return false, err
	}

	if savePath == "" {
		return false, nil // User cancelled
	}

	// Write file
	err = os.WriteFile(savePath, data, 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ==================== Settings API ====================

// GetSettings gets settings
func (a *App) GetSettings() (*models.Settings, error) {
	return a.settingsRepo.Get()
}

// UpdateSettings updates settings
func (a *App) UpdateSettings(settings models.Settings) error {
	if settings.EnableAutoStart {
		if err := utils.EnableAutoStart(); err != nil {
			return fmt.Errorf("failed to enable auto start: %w", err)
		}
	} else {
		if err := utils.DisableAutoStart(); err != nil {
			return fmt.Errorf("failed to disable auto start: %w", err)
		}
	}
	return a.settingsRepo.Update(&settings)
}

// ==================== Todo Types API ====================

// GetTodoTypes returns all todo types
func (a *App) GetTodoTypes() []map[string]string {
	return []map[string]string{
		{"value": "birthday", "label": "生日", "icon": "\U0001F382", "color": "#FF6B6B"},
		{"value": "work", "label": "工作", "icon": "\U0001F4BC", "color": "#4ECDC4"},
		{"value": "anniversary", "label": "纪念日", "icon": "\U0001F49D", "color": "#FF69B4"},
		{"value": "reminder", "label": "提醒", "icon": "\u23F0", "color": "#FFD93D"},
		{"value": "task", "label": "任务", "icon": "\u2705", "color": "#6BCB77"},
	}
}

// OpenWidget 打开桌面小部件窗口
func (a *App) OpenWidget() error {
	// 检查小部件是否已经在运行
	if utils.IsWindowRunning("待办提醒") {
		return fmt.Errorf("widget is already running")
	}

	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// 使用 --widget 参数启动新进程
	cmd := utils.StartProcess(exePath, "--widget")
	if cmd == nil {
		return fmt.Errorf("failed to start widget process")
	}

	return nil
}

// CloseWidget 关闭桌面小部件窗口
func (a *App) CloseWidget() error {
	if !utils.CloseWindow("待办提醒") {
		return fmt.Errorf("widget is not running")
	}
	return nil
}

// IsWidgetRunning 检测小部件是否在运行
func (a *App) IsWidgetRunning() bool {
	return utils.IsWindowRunning("待办提醒")
}

// OpenMainWindowWithTodo 打开主窗口并显示指定待办的详情
// 通过IPC文件传递待办ID给主窗口
func (a *App) OpenMainWindowWithTodo(todoId int64) error {
	err := utils.WriteIPCTodoId(todoId)
	if err != nil {
		return err
	}
	// 将主窗口置顶
	utils.BringWindowToFront("待办日历")
	return nil
}

// CheckIPCTodo 检查IPC文件中是否有待办ID需要打开
func (a *App) CheckIPCTodo() (int64, error) {
	return utils.ReadIPCTodoId()
}

// ==================== Sound API ====================

// SoundInfo 声音信息（导出给前端）
type SoundInfo = notification.SoundInfo

// GetAvailableSounds 获取可用的声音列表
func (a *App) GetAvailableSounds() ([]SoundInfo, error) {
	return notification.GetAvailableSounds()
}

// PreviewSound 预览声音
func (a *App) PreviewSound(soundPath string) error {
	return notification.PreviewSound(soundPath)
}

// ImportSound 导入自定义声音（打开文件选择对话框）
func (a *App) ImportSound() (string, error) {
	// 打开文件选择对话框
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择声音文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "WAV 音频文件 (*.wav)",
				Pattern:     "*.wav",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", nil // 用户取消选择
	}

	// 导入声音文件
	return notification.ImportSound(filePath)
}

// DeleteSound 删除自定义声音
func (a *App) DeleteSound(soundPath string) error {
	return notification.DeleteSound(soundPath)
}
