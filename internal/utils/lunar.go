package utils

import (
	"time"

	"todo-calendar/internal/models"

	"github.com/6tail/lunar-go/calendar"
)

// SolarToLunar 公历转农历
func SolarToLunar(year, month, day int) models.LunarDate {
	solar := calendar.NewSolarFromYmd(year, month, day)
	lunar := solar.GetLunar()

	return models.LunarDate{
		Year:      lunar.GetYear(),
		Month:     lunar.GetMonth(),
		Day:       lunar.GetDay(),
		MonthName: lunar.GetMonthInChinese() + "月",
		DayName:   lunar.GetDayInChinese(),
		YearName:  lunar.GetYearInGanZhi(),
		IsLeap:    lunar.GetMonth() < 0,
		Animal:    lunar.GetYearShengXiao(),
	}
}

// LunarToSolar 农历转公历
func LunarToSolar(year, month, day int, isLeap bool) (time.Time, error) {
	lunarMonth := month
	if isLeap {
		lunarMonth = -month
	}
	lunar := calendar.NewLunar(year, lunarMonth, day, 0, 0, 0)
	solar := lunar.GetSolar()

	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 0, 0, 0, time.Local), nil
}

// GetSolarDateInfo 获取公历日期详细信息(用于hover显示)
func GetSolarDateInfo(year, month, day int) map[string]interface{} {
	solar := calendar.NewSolarFromYmd(year, month, day)
	lunar := solar.GetLunar()

	// 获取周数
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	_, weekNum := t.ISOWeek()

	// 获取节日 - 使用 ToList 转换
	solarFestivals := solar.GetFestivals()
	lunarFestivals := lunar.GetFestivals()
	jieQi := lunar.GetJieQi()

	festivals := []string{}
	// 遍历公历节日列表
	if solarFestivals != nil {
		for e := solarFestivals.Front(); e != nil; e = e.Next() {
			if s, ok := e.Value.(string); ok {
				festivals = append(festivals, s)
			}
		}
	}
	// 遍历农历节日列表
	if lunarFestivals != nil {
		for e := lunarFestivals.Front(); e != nil; e = e.Next() {
			if s, ok := e.Value.(string); ok {
				festivals = append(festivals, s)
			}
		}
	}
	if jieQi != "" {
		festivals = append(festivals, jieQi)
	}

	return map[string]interface{}{
		"solar": map[string]interface{}{
			"year":  year,
			"month": month,
			"day":   day,
		},
		"lunar": map[string]interface{}{
			"year":      lunar.GetYear(),
			"month":     lunar.GetMonth(),
			"day":       lunar.GetDay(),
			"monthName": lunar.GetMonthInChinese() + "月",
			"dayName":   lunar.GetDayInChinese(),
			"yearName":  lunar.GetYearInGanZhi() + "年",
			"animal":    lunar.GetYearShengXiao(),
		},
		"weekNumber": weekNum,
		"weekDay":    t.Weekday().String(),
		"festivals":  festivals,
	}
}
