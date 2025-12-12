package database

import (
	"database/sql"

	"todo-calendar/internal/models"
)

// SettingsRepository 设置仓库
type SettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository 创建设置仓库实例
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get 获取设置
func (r *SettingsRepository) Get() (*models.Settings, error) {
	query := `
		SELECT id, enable_widget, enable_auto_start, minimize_to_tray, 
			   notification_sound, notification_duration, widget_position, 
			   widget_opacity, theme, COALESCE(notification_sound_file, '')
		FROM settings WHERE id = 1
	`
	settings := &models.Settings{}
	err := r.db.QueryRow(query).Scan(
		&settings.ID,
		&settings.EnableWidget,
		&settings.EnableAutoStart,
		&settings.MinimizeToTray,
		&settings.NotificationSound,
		&settings.NotificationDuration,
		&settings.WidgetPosition,
		&settings.WidgetOpacity,
		&settings.Theme,
		&settings.NotificationSoundFile,
	)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// Update 更新设置
func (r *SettingsRepository) Update(settings *models.Settings) error {
	query := `
		UPDATE settings SET
			enable_widget = ?,
			enable_auto_start = ?,
			minimize_to_tray = ?,
			notification_sound = ?,
			notification_duration = ?,
			widget_position = ?,
			widget_opacity = ?,
			theme = ?,
			notification_sound_file = ?
		WHERE id = 1
	`
	_, err := r.db.Exec(query,
		settings.EnableWidget,
		settings.EnableAutoStart,
		settings.MinimizeToTray,
		settings.NotificationSound,
		settings.NotificationDuration,
		settings.WidgetPosition,
		settings.WidgetOpacity,
		settings.Theme,
		settings.NotificationSoundFile,
	)
	return err
}
