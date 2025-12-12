package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

// InitDB 初始化数据库
func InitDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		// 获取数据目录
		dataDir, e := getDataDir()
		if e != nil {
			err = e
			return
		}

		// 确保目录存在
		if e := os.MkdirAll(dataDir, 0755); e != nil {
			err = e
			return
		}

		dbPath := filepath.Join(dataDir, "todo_calendar.db")
		db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
		if err != nil {
			return
		}

		// 设置连接池
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		// 创建表
		if e := createTables(); e != nil {
			err = e
			return
		}
	})
	return db, err
}

// getDataDir 获取数据存储目录
func getDataDir() (string, error) {
	// 优先使用程序所在目录
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "data"), nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return db
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// createTables 创建数据表
func createTables() error {
	// 创建待办事项表
	todoTable := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT DEFAULT '',
		type TEXT NOT NULL DEFAULT 'task',
		start_date DATETIME NOT NULL,
		end_date DATETIME NOT NULL,
		is_lunar INTEGER DEFAULT 0,
		hide_year INTEGER DEFAULT 0,
		cron_expr TEXT DEFAULT '',
		repeat_count INTEGER DEFAULT 1,
		current_repeat INTEGER DEFAULT 1,
		advance_remind INTEGER DEFAULT 15,
		remind_at_start INTEGER DEFAULT 1,
		remind_at_end INTEGER DEFAULT 1,
		start_remind_triggered INTEGER DEFAULT 0,
		is_completed INTEGER DEFAULT 0,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_todos_date ON todos(start_date, end_date);
	CREATE INDEX IF NOT EXISTS idx_todos_type ON todos(type);
	CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(is_completed);
	`

	// 创建附件表
	attachmentTable := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo_id INTEGER NOT NULL,
		file_name TEXT NOT NULL,
		storage_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		mime_type TEXT DEFAULT '',
		is_encrypted INTEGER DEFAULT 1,
		encryption_key TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_attachments_todo ON attachments(todo_id);
	`

	// 创建设置表
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		enable_widget INTEGER DEFAULT 1,
		enable_auto_start INTEGER DEFAULT 0,
		minimize_to_tray INTEGER DEFAULT 1,
		notification_sound INTEGER DEFAULT 1,
		notification_duration INTEGER DEFAULT 5,
		widget_position TEXT DEFAULT 'bottom-right',
		widget_opacity INTEGER DEFAULT 90,
		theme TEXT DEFAULT 'light',
		backend_url TEXT DEFAULT ''
	);
	INSERT OR IGNORE INTO settings (id) VALUES (1);
	`

	// 创建通知记录表
	notificationTable := `
	CREATE TABLE IF NOT EXISTS notification_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo_id INTEGER NOT NULL,
		notified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_notification_todo ON notification_logs(todo_id);
	`

	tables := []string{todoTable, attachmentTable, settingsTable, notificationTable}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// 迁移：添加 backend_url 字段（如果不存在）
	migrateBackendUrl := `
	ALTER TABLE settings ADD COLUMN backend_url TEXT DEFAULT '';
	`
	db.Exec(migrateBackendUrl) // 忽略错误，如果字段已存在

	// 迁移：添加 notification_sound_file 字段（如果不存在）
	migrateSoundFile := `
	ALTER TABLE settings ADD COLUMN notification_sound_file TEXT DEFAULT '';
	`
	db.Exec(migrateSoundFile) // 忽略错误，如果字段已存在

	return nil
}
