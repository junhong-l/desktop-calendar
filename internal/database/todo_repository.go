package database

import (
	"database/sql"
	"fmt"
	"time"

	"todo-calendar/internal/models"
)

// TodoRepository 待办事项仓库
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository 创建待办仓库实例
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create 创建待办事项
func (r *TodoRepository) Create(todo *models.Todo) (int64, error) {
	query := `
		INSERT INTO todos (title, content, type, start_date, end_date, is_lunar, hide_year, cron_expr, 
			repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end, start_remind_triggered, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	// 设置默认值
	if todo.AdvanceRemind <= 0 {
		todo.AdvanceRemind = 15
	}
	if todo.RepeatCount <= 0 {
		todo.RepeatCount = 1
	}
	if todo.CurrentRepeat <= 0 {
		todo.CurrentRepeat = 1
	}
	result, err := r.db.Exec(query,
		todo.Title,
		todo.Content,
		todo.Type,
		todo.StartDate,
		todo.EndDate,
		todo.IsLunar,
		todo.HideYear,
		todo.CronExpr,
		todo.RepeatCount,
		todo.CurrentRepeat,
		todo.AdvanceRemind,
		todo.RemindAtStart,
		todo.RemindAtEnd,
		todo.StartRemindTriggered,
		now,
		now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update 更新待办事项
func (r *TodoRepository) Update(todo *models.Todo) error {
	query := `
		UPDATE todos SET
			title = ?,
			content = ?,
			type = ?,
			start_date = ?,
			end_date = ?,
			is_lunar = ?,
			hide_year = ?,
			cron_expr = ?,
			repeat_count = ?,
			current_repeat = ?,
			advance_remind = ?,
			remind_at_start = ?,
			remind_at_end = ?,
			updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query,
		todo.Title,
		todo.Content,
		todo.Type,
		todo.StartDate,
		todo.EndDate,
		todo.IsLunar,
		todo.HideYear,
		todo.CronExpr,
		todo.RepeatCount,
		todo.CurrentRepeat,
		todo.AdvanceRemind,
		todo.RemindAtStart,
		todo.RemindAtEnd,
		time.Now(),
		todo.ID,
	)
	return err
}

// Delete 删除待办事项
func (r *TodoRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

// GetByID 根据ID获取待办事项
func (r *TodoRepository) GetByID(id int64) (*models.Todo, error) {
	query := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos WHERE id = ?
	`
	todo := &models.Todo{}
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Content,
		&todo.Type,
		&todo.StartDate,
		&todo.EndDate,
		&todo.IsLunar,
		&todo.HideYear,
		&todo.CronExpr,
		&todo.RepeatCount,
		&todo.CurrentRepeat,
		&todo.AdvanceRemind,
		&todo.RemindAtStart,
		&todo.RemindAtEnd,
		&todo.StartRemindTriggered,
		&todo.IsCompleted,
		&completedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		ft := &models.FlexTime{Time: completedAt.Time}
		todo.CompletedAt = ft
	}
	return todo, nil
}

// List 获取待办列表
func (r *TodoRepository) List(filter models.TodoFilter) (*models.TodoListResult, error) {
	// 构建查询条件
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Keyword != "" {
		where += " AND title LIKE ?"
		args = append(args, "%"+filter.Keyword+"%")
	}

	if filter.Year > 0 {
		// 使用 substr 匹配年份，兼容各种日期格式
		where += " AND substr(start_date, 1, 4) = ?"
		args = append(args, fmt.Sprintf("%d", filter.Year))
	}

	if filter.Month > 0 {
		// 使用 substr 匹配月份，兼容各种日期格式
		where += " AND substr(start_date, 6, 2) = ?"
		args = append(args, fmt.Sprintf("%02d", filter.Month))
	}

	if len(filter.Types) > 0 {
		placeholders := ""
		for i, t := range filter.Types {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, t)
		}
		where += " AND type IN (" + placeholders + ")"
	}

	if filter.Completed != nil {
		where += " AND is_completed = ?"
		if *filter.Completed {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	// 获取总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM todos " + where
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// 分页
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PageSize

	// 查询数据
	query := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos ` + where + `
		ORDER BY start_date ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		var completedAt sql.NullTime
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Content,
			&todo.Type,
			&todo.StartDate,
			&todo.EndDate,
			&todo.IsLunar,
			&todo.HideYear,
			&todo.CronExpr,
			&todo.RepeatCount,
			&todo.CurrentRepeat,
			&todo.AdvanceRemind,
			&todo.RemindAtStart,
			&todo.RemindAtEnd,
			&todo.StartRemindTriggered,
			&todo.IsCompleted,
			&completedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			ft := &models.FlexTime{Time: completedAt.Time}
			todo.CompletedAt = ft
		}
		todos = append(todos, todo)
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &models.TodoListResult{
		Todos:      todos,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByDateRange 获取日期范围内的待办
func (r *TodoRepository) GetByDateRange(start, end time.Time) ([]models.Todo, error) {
	query := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos 
		WHERE (start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?)
			  OR (start_date <= ? AND end_date >= ?)
		ORDER BY start_date ASC
	`
	rows, err := r.db.Query(query, start, end, start, end, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTodos(rows)
}

// GetPendingTodos 获取待处理的待办(今天和过期)
func (r *TodoRepository) GetPendingTodos() ([]models.Todo, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	query := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos 
		WHERE is_completed = 0 AND start_date <= ?
		ORDER BY start_date ASC
	`
	rows, err := r.db.Query(query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTodos(rows)
}

// GetWeekTodos 获取本周待办
func (r *TodoRepository) GetWeekTodos() (*models.WeekTodos, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	weekEnd := weekStart.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	todos, err := r.GetByDateRange(weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	// 获取逾期未完成
	overdueQuery := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos 
		WHERE is_completed = 0 AND end_date < ?
		ORDER BY start_date ASC
	`
	rows, err := r.db.Query(overdueQuery, weekStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overdue, err := r.scanTodos(rows)
	if err != nil {
		return nil, err
	}

	return &models.WeekTodos{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
		Todos:     todos,
		Overdue:   overdue,
	}, nil
}

// MarkCompleted 标记完成
func (r *TodoRepository) MarkCompleted(id int64, completed bool) error {
	var query string
	if completed {
		query = "UPDATE todos SET is_completed = 1, completed_at = ?, start_remind_triggered = 1 WHERE id = ?"
		_, err := r.db.Exec(query, time.Now(), id)
		return err
	}
	query = "UPDATE todos SET is_completed = 0, completed_at = NULL WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// MarkStartRemindTriggered 标记开始提醒已触发
func (r *TodoRepository) MarkStartRemindTriggered(id int64) error {
	query := "UPDATE todos SET start_remind_triggered = 1 WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// GetTodayStartRemindTodos 获取今天需要开始提醒且未触发的待办
func (r *TodoRepository) GetTodayStartRemindTodos() ([]models.Todo, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	query := `
		SELECT id, title, content, type, start_date, end_date, is_lunar, hide_year, 
			   cron_expr, repeat_count, current_repeat, advance_remind, remind_at_start, remind_at_end,
			   start_remind_triggered, is_completed, completed_at, created_at, updated_at
		FROM todos 
		WHERE is_completed = 0 
		  AND remind_at_start = 1 
		  AND start_remind_triggered = 0
		  AND start_date >= ? AND start_date <= ?
		ORDER BY start_date ASC
	`
	rows, err := r.db.Query(query, todayStart, todayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTodos(rows)
}

// IncrementRepeatCount 增加循环次数
func (r *TodoRepository) IncrementRepeatCount(id int64) error {
	query := "UPDATE todos SET current_repeat = current_repeat + 1 WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// scanTodos 扫描待办列表
func (r *TodoRepository) scanTodos(rows *sql.Rows) ([]models.Todo, error) {
	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		var completedAt sql.NullTime
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Content,
			&todo.Type,
			&todo.StartDate,
			&todo.EndDate,
			&todo.IsLunar,
			&todo.HideYear,
			&todo.CronExpr,
			&todo.RepeatCount,
			&todo.CurrentRepeat,
			&todo.AdvanceRemind,
			&todo.RemindAtStart,
			&todo.RemindAtEnd,
			&todo.StartRemindTriggered,
			&todo.IsCompleted,
			&completedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			ft := &models.FlexTime{Time: completedAt.Time}
			todo.CompletedAt = ft
		}
		todos = append(todos, todo)
	}
	return todos, nil
}
