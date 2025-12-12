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
