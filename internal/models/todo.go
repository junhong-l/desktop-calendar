package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// FlexTime 灵活时间类型，支持多种格式解析
type FlexTime struct {
	time.Time
}

// UnmarshalJSON 自定义JSON解析
func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		ft.Time = time.Time{}
		return nil
	}

	// 尝试多种格式
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var parseErr error
	for _, format := range formats {
		t, err := time.ParseInLocation(format, s, time.Local)
		if err == nil {
			ft.Time = t
			return nil
		}
		parseErr = err
	}
	return parseErr
}

// MarshalJSON 自定义JSON序列化
func (ft FlexTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// Scan 实现 sql.Scanner 接口
func (ft *FlexTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ft.Time = v
		return nil
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			t, err := time.ParseInLocation(format, v, time.Local)
			if err == nil {
				ft.Time = t
				return nil
			}
		}
		ft.Time = time.Time{}
		return nil
	default:
		ft.Time = time.Time{}
		return nil
	}
}

// Value 实现 driver.Valuer 接口
func (ft FlexTime) Value() (driver.Value, error) {
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

// TodoType 待办类型
type TodoType string

const (
	TodoTypeBirthday    TodoType = "birthday"    // 生日
	TodoTypeWork        TodoType = "work"        // 工作
	TodoTypeAnniversary TodoType = "anniversary" // 纪念日
	TodoTypeReminder    TodoType = "reminder"    // 提醒
	TodoTypeTask        TodoType = "task"        // 任务
)

// Todo 待办事项模型
type Todo struct {
	ID                   int64     `json:"id"`
	Title                string    `json:"title"`                // 标题
	Content              string    `json:"content"`              // 内容(Markdown格式)
	Type                 TodoType  `json:"type"`                 // 类型
	StartDate            FlexTime  `json:"startDate"`            // 开始日期
	EndDate              FlexTime  `json:"endDate"`              // 结束日期
	IsLunar              bool      `json:"isLunar"`              // 是否农历(生日专用)
	HideYear             bool      `json:"hideYear"`             // 隐藏年份(生日专用)
	CronExpr             string    `json:"cronExpr"`             // Crontab表达式
	RepeatCount          int       `json:"repeatCount"`          // 循环次数(0表示不限次数)
	CurrentRepeat        int       `json:"currentRepeat"`        // 当前已循环次数
	AdvanceRemind        int       `json:"advanceRemind"`        // 提前提醒(分钟)，默认15
	RemindAtStart        bool      `json:"remindAtStart"`        // 到点提醒(开始时间)
	RemindAtEnd          bool      `json:"remindAtEnd"`          // 结束提醒(结束时间)
	StartRemindTriggered bool      `json:"startRemindTriggered"` // 开始提醒是否已触发
	IsCompleted          bool      `json:"isCompleted"`          // 是否完成
	CompletedAt          *FlexTime `json:"completedAt"`          // 完成时间
	CreatedAt            FlexTime  `json:"createdAt"`            // 创建时间
	UpdatedAt            FlexTime  `json:"updatedAt"`            // 更新时间
}

// Attachment 附件模型
type Attachment struct {
	ID            int64     `json:"id"`
	TodoID        int64     `json:"todoId"`
	FileName      string    `json:"fileName"`      // 原文件名
	StoragePath   string    `json:"storagePath"`   // 存储路径
	FileSize      int64     `json:"fileSize"`      // 文件大小
	MimeType      string    `json:"mimeType"`      // MIME类型
	IsEncrypted   bool      `json:"isEncrypted"`   // 是否加密
	EncryptionKey string    `json:"encryptionKey"` // 加密密钥(存储在数据库)
	CreatedAt     time.Time `json:"createdAt"`
}

// Settings 系统设置
type Settings struct {
	ID                    int64  `json:"id"`
	EnableWidget          bool   `json:"enableWidget"`          // 启用桌面小部件
	EnableAutoStart       bool   `json:"enableAutoStart"`       // 开机自启
	MinimizeToTray        bool   `json:"minimizeToTray"`        // 最小化到托盘
	NotificationSound     bool   `json:"notificationSound"`     // 通知声音
	NotificationSoundFile string `json:"notificationSoundFile"` // 通知声音文件路径
	NotificationDuration  int    `json:"notificationDuration"`  // 通知显示时长(秒)
	WidgetPosition        string `json:"widgetPosition"`        // 小部件位置
	WidgetOpacity         int    `json:"widgetOpacity"`         // 小部件透明度
	Theme                 string `json:"theme"`                 // 主题
}

// TodoFilter 待办筛选条件
type TodoFilter struct {
	Keyword   string   `json:"keyword"`   // 搜索关键词
	Year      int      `json:"year"`      // 年份
	Month     int      `json:"month"`     // 月份
	Types     []string `json:"types"`     // 类型筛选
	Completed *bool    `json:"completed"` // 完成状态
	Page      int      `json:"page"`      // 页码
	PageSize  int      `json:"pageSize"`  // 每页数量
}

// TodoListResult 待办列表结果
type TodoListResult struct {
	Todos      []Todo `json:"todos"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// CronNextRun Cron表达式下次执行时间
type CronNextRun struct {
	Expression string      `json:"expression"`
	NextRuns   []time.Time `json:"nextRuns"`
	IsValid    bool        `json:"isValid"`
	Error      string      `json:"error,omitempty"`
}

// LunarDate 农历日期
type LunarDate struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	MonthName string `json:"monthName"` // 农历月份名称
	DayName   string `json:"dayName"`   // 农历日期名称
	YearName  string `json:"yearName"`  // 农历年份名称(干支)
	IsLeap    bool   `json:"isLeap"`    // 是否闰月
	Animal    string `json:"animal"`    // 生肖
}

// CalendarDay 日历天信息
type CalendarDay struct {
	Date           string    `json:"date"` // 使用字符串格式: "2006-01-02"
	Day            int       `json:"day"`
	WeekNumber     int       `json:"weekNumber"` // 周数
	Lunar          LunarDate `json:"lunar"`      // 农历信息
	IsToday        bool      `json:"isToday"`
	IsCurrentMonth bool      `json:"isCurrentMonth"`
	Todos          []Todo    `json:"todos"`     // 当天待办
	TodoCount      int       `json:"todoCount"` // 待办数量
}

// WeekTodos 本周待办
type WeekTodos struct {
	WeekStart string `json:"weekStart"` // 使用字符串格式: "2006-01-02"
	WeekEnd   string `json:"weekEnd"`   // 使用字符串格式: "2006-01-02"
	Todos     []Todo `json:"todos"`
	Overdue   []Todo `json:"overdue"` // 逾期未完成
}

// NotificationData 通知数据
type NotificationData struct {
	Todo         Todo   `json:"todo"`
	CurrentCount int    `json:"currentCount"` // 当前提醒次数
	TotalCount   int    `json:"totalCount"`   // 总提醒次数
	Message      string `json:"message"`
}
