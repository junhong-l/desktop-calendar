package utils

import (
	"time"

	"todo-calendar/internal/models"

	"github.com/gorhill/cronexpr"
)

// ParseCronExpr 解析Cron表达式并返回接下来5次执行时间
func ParseCronExpr(expr string) models.CronNextRun {
	result := models.CronNextRun{
		Expression: expr,
		NextRuns:   []time.Time{},
		IsValid:    false,
	}

	if expr == "" {
		result.Error = "Cron表达式为空"
		return result
	}

	// 解析cron表达式
	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		result.Error = "无效的Cron表达式: " + err.Error()
		return result
	}

	result.IsValid = true

	// 计算接下来5次执行时间
	now := time.Now()
	for i := 0; i < 5; i++ {
		next := cronExpr.Next(now)
		result.NextRuns = append(result.NextRuns, next)
		now = next
	}

	return result
}

// GetNextCronTime 获取下一次Cron执行时间
func GetNextCronTime(expr string) (time.Time, error) {
	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return cronExpr.Next(time.Now()), nil
}

// IsCronExprValid 检查Cron表达式是否有效
func IsCronExprValid(expr string) bool {
	if expr == "" {
		return false
	}
	_, err := cronexpr.Parse(expr)
	return err == nil
}

// CalculateEndDateByRemindCount 根据开始时间、Cron表达式和提醒次数计算结束时间
func CalculateEndDateByRemindCount(startTime time.Time, expr string, remindCount int) (time.Time, error) {
	if expr == "" {
		// 没有 cron 表达式，结束时间就是开始时间
		return startTime, nil
	}

	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}

	// 从开始时间计算，执行 remindCount 次后的时间
	current := startTime
	for i := 0; i < remindCount; i++ {
		current = cronExpr.Next(current)
	}
	return current, nil
}

// CalculateRemindCountByEndDate 根据开始时间、Cron表达式和结束时间计算提醒次数
func CalculateRemindCountByEndDate(startTime time.Time, expr string, endTime time.Time) int {
	if expr == "" {
		return 1
	}

	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		return 1
	}

	// 计算从开始时间到结束时间之间有多少次执行
	count := 0
	current := startTime
	for {
		next := cronExpr.Next(current)
		if next.After(endTime) {
			break
		}
		count++
		current = next
		// 防止无限循环
		if count > 1000 {
			break
		}
	}

	if count == 0 {
		return 1
	}
	return count
}

// GetCronDatesInRange 获取在指定日期范围内的所有cron执行日期
// 返回日期字符串集合，格式为 "2006-01-02"
func GetCronDatesInRange(expr string, todoStartTime time.Time, todoEndTime time.Time, rangeStart time.Time, rangeEnd time.Time) map[string]bool {
	dates := make(map[string]bool)

	if expr == "" {
		// 没有 cron 表达式，只返回开始日期那一天
		dateKey := todoStartTime.Format("2006-01-02")
		if !todoStartTime.Before(rangeStart) && !todoStartTime.After(rangeEnd) {
			dates[dateKey] = true
		}
		return dates
	}

	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		// 解析失败，返回开始日期
		dateKey := todoStartTime.Format("2006-01-02")
		if !todoStartTime.Before(rangeStart) && !todoStartTime.After(rangeEnd) {
			dates[dateKey] = true
		}
		return dates
	}

	// 从 todoStartTime 的前一秒开始计算，确保包含开始时间本身
	current := todoStartTime.Add(-time.Second)
	maxIterations := 1000 // 防止无限循环

	for i := 0; i < maxIterations; i++ {
		next := cronExpr.Next(current)

		// 如果下次执行时间超过了待办的结束时间或范围结束时间，停止
		if next.After(todoEndTime) || next.After(rangeEnd) {
			break
		}

		// 如果在查询范围内，添加到结果
		if !next.Before(rangeStart) {
			dateKey := next.Format("2006-01-02")
			dates[dateKey] = true
		}

		current = next
	}

	return dates
}

// GetCronScheduledTimes 获取cron表达式的执行时间列表
// 从startTime开始，计算count次执行的具体时间
func GetCronScheduledTimes(expr string, startTime time.Time, count int) []time.Time {
	var times []time.Time

	if expr == "" || count <= 0 {
		return times
	}

	cronExpr, err := cronexpr.Parse(expr)
	if err != nil {
		return times
	}

	// 从 startTime 的前一秒开始计算，确保包含开始时间本身
	current := startTime.Add(-time.Second)

	for i := 0; i < count; i++ {
		next := cronExpr.Next(current)
		// 如果返回了零值时间，停止计算
		if next.IsZero() {
			break
		}
		times = append(times, next)
		current = next
	}

	return times
}
