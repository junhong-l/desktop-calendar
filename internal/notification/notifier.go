package notification

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"todo-calendar/internal/database"
	"todo-calendar/internal/models"
	"todo-calendar/internal/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotifyAdvance NotificationType = "advance" // 提前提醒
	NotifyStart   NotificationType = "start"   // 到点提醒
	NotifyEnd     NotificationType = "end"     // 结束提醒
)

// Notifier 通知管理器
type Notifier struct {
	ctx          context.Context
	db           *sql.DB
	todoRepo     *database.TodoRepository
	settingsRepo *database.SettingsRepository
	ticker       *time.Ticker
	stopChan     chan struct{}
	notifiedMap  map[string]bool // 记录已通知的key: "todoID-type-date"
	notifiedLock sync.RWMutex
}

// NewNotifier 创建通知管理器
func NewNotifier(db *sql.DB) *Notifier {
	return &Notifier{
		db:           db,
		todoRepo:     database.NewTodoRepository(db),
		settingsRepo: database.NewSettingsRepository(db),
		stopChan:     make(chan struct{}),
		notifiedMap:  make(map[string]bool),
	}
}

// SetContext 设置上下文
func (n *Notifier) SetContext(ctx context.Context) {
	n.ctx = ctx
}

// getNotifyKey 生成通知key，用于避免重复通知
func (n *Notifier) getNotifyKey(todoID int64, notifyType NotificationType, date time.Time) string {
	return fmt.Sprintf("%d-%s-%s", todoID, notifyType, date.Format("2006-01-02"))
}

// hasNotified 检查是否已通知
func (n *Notifier) hasNotified(key string) bool {
	n.notifiedLock.RLock()
	defer n.notifiedLock.RUnlock()
	return n.notifiedMap[key]
}

// markNotified 标记已通知
func (n *Notifier) markNotified(key string) {
	n.notifiedLock.Lock()
	defer n.notifiedLock.Unlock()
	n.notifiedMap[key] = true
}

// StartNotificationChecker 启动通知检查器
func (n *Notifier) StartNotificationChecker() {
	n.ticker = time.NewTicker(30 * time.Second) // 每30秒检查一次

	for {
		select {
		case <-n.ticker.C:
			n.checkAndNotify()
		case <-n.stopChan:
			n.ticker.Stop()
			return
		}
	}
}

// Stop 停止通知检查器
func (n *Notifier) Stop() {
	close(n.stopChan)
}

// CheckPendingTodos 检查并通知待处理的待办(启动时调用)
func (n *Notifier) CheckPendingTodos() {
	time.Sleep(2 * time.Second) // 等待前端就绪
	n.checkAndNotify()
}

// checkAndNotify 检查并发送通知
func (n *Notifier) checkAndNotify() {
	todos, err := n.todoRepo.GetPendingTodos()
	if err != nil {
		return
	}

	// 获取设置，检查是否开启声音
	settings, err := n.settingsRepo.Get()
	playSound := true
	soundFile := ""
	if err == nil {
		playSound = settings.NotificationSound
		soundFile = settings.NotificationSoundFile
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	for _, todo := range todos {
		if todo.IsCompleted {
			continue
		}

		startTime := todo.StartDate.Time
		endTime := todo.EndDate.Time

		// 检查是否是今天的待办
		if startTime.Format("2006-01-02") != today && endTime.Format("2006-01-02") != today {
			continue
		}

		// 1. 提前提醒
		if todo.AdvanceRemind > 0 {
			advanceTime := startTime.Add(-time.Duration(todo.AdvanceRemind) * time.Minute)
			key := n.getNotifyKey(todo.ID, NotifyAdvance, now)

			if !n.hasNotified(key) && n.isTimeMatch(now, advanceTime) {
				title := fmt.Sprintf("⏰提前提醒: %s", todo.Title)
				message := fmt.Sprintf("将在 %d 分钟后开始", todo.AdvanceRemind)
				n.sendWindowsNotification(todo, title, message, playSound, soundFile, NotifyAdvance)
				n.markNotified(key)
			}
		}

		// 2. 到点提醒 (开始时间)
		if todo.RemindAtStart {
			key := n.getNotifyKey(todo.ID, NotifyStart, now)

			if !n.hasNotified(key) && n.isTimeMatch(now, startTime) {
				title := fmt.Sprintf("🔔开始: %s", todo.Title)
				message := "任务已开始"
				if todo.Content != "" {
					message = todo.Content
				}
				n.sendWindowsNotification(todo, title, message, playSound, soundFile, NotifyStart)
				n.markNotified(key)
			}
		}

		// 3. 结束提醒
		if todo.RemindAtEnd && !endTime.IsZero() {
			key := n.getNotifyKey(todo.ID, NotifyEnd, now)

			if !n.hasNotified(key) && n.isTimeMatch(now, endTime) {
				title := fmt.Sprintf("✅结束提醒: %s", todo.Title)
				message := "已到任务结束时间。"
				n.sendWindowsNotification(todo, title, message, playSound, soundFile, NotifyEnd)
				n.markNotified(key)
			}
		}
	}
}

// isTimeMatch 检查当前时间是否匹配目标时间（精确到分钟，允许30秒误差）
func (n *Notifier) isTimeMatch(now, target time.Time) bool {
	if target.IsZero() {
		return false
	}

	// 比较小时和分钟
	if now.Year() == target.Year() &&
		now.YearDay() == target.YearDay() &&
		now.Hour() == target.Hour() &&
		now.Minute() == target.Minute() {
		return true
	}
	return false
}

// sendWindowsNotification 发送 Windows Toast 通知
func (n *Notifier) sendWindowsNotification(todo models.Todo, title, message string, playSound bool, soundFile string, notifyType NotificationType) {
	// 播放声音
	if playSound {
		go func() {
			if soundFile != "" && soundFile != "default" {
				PlaySoundFileAsync(soundFile)
			} else {
				PlaySystemSound()
			}
		}()
	}

	// 启动通知弹窗进程
	go func() {
		exePath, err := os.Executable()
		if err != nil {
			return
		}

		// 获取通知类型显示名称
		var typeLabel string
		switch notifyType {
		case NotifyAdvance:
			typeLabel = "提前提醒"
		case NotifyStart:
			typeLabel = "开始提醒"
		case NotifyEnd:
			typeLabel = "结束提醒"
		default:
			typeLabel = "提醒"
		}

		// 启动通知弹窗进程
		utils.StartProcess(exePath,
			"--notify",
			"--notify-title", title,
			"--notify-message", message,
			"--notify-type", typeLabel,
			"--notify-todo", fmt.Sprintf("%d", todo.ID),
			"--notify-start", todo.StartDate.Time.Format("2006-01-02 15:04"),
			"--notify-end", todo.EndDate.Time.Format("2006-01-02 15:04"),
		)
	}()

	// 同时发送到前端（用于主窗口内的通知）
	n.sendNotification(todo)
}

// sendNotification 发送通知
func (n *Notifier) sendNotification(todo models.Todo) {
	// 构建通知数据
	notification := models.NotificationData{
		Todo:         todo,
		CurrentCount: 1,
		TotalCount:   1,
		Message:      todo.Title,
	}

	// 发送到前端
	if n.ctx != nil {
		runtime.EventsEmit(n.ctx, "todo:notification", notification)
	}
}

// GetPendingNotifications 获取待处理的通知
func (n *Notifier) GetPendingNotifications() ([]models.NotificationData, error) {
	todos, err := n.todoRepo.GetPendingTodos()
	if err != nil {
		return nil, err
	}

	notifications := []models.NotificationData{}
	for _, todo := range todos {
		notifications = append(notifications, models.NotificationData{
			Todo:         todo,
			CurrentCount: 1,
			TotalCount:   1,
			Message:      todo.Title,
		})
	}
	return notifications, nil
}

// MarkNotified 标记已通知（不再使用，保留接口兼容）
func (n *Notifier) MarkNotified(todoID int64) error {
	// 不再增加循环次数
	return nil
}
