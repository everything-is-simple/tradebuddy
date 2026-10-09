package calc

import (
	"fmt"
	"time"
	"tradebuddy/internal/model"
)

// AggregateToWeekly 将日K线聚合为周K线
func AggregateToWeekly(dailyBars []*model.DailyBar) ([]*model.WeeklyBar, error) {
	if len(dailyBars) == 0 {
		return nil, fmt.Errorf("daily bars is empty")
	}

	var weeklyBars []*model.WeeklyBar
	var currentWeek *model.WeeklyBar

	for _, bar := range dailyBars {
		// 解析日期
		date, err := time.Parse("2006-01-02", bar.Date)
		if err != nil {
			continue
		}

		year, week := date.ISOWeek()

		// 新的一周，创建新的周K线
		if currentWeek == nil || currentWeek.Year != year || currentWeek.Week != week {
			if currentWeek != nil {
				weeklyBars = append(weeklyBars, currentWeek)
			}

			currentWeek = &model.WeeklyBar{
				Code:   bar.Code,
				Year:   year,
				Week:   week,
				Date:   bar.Date,
				Open:   bar.Open,
				High:   bar.High,
				Low:    bar.Low,
				Close:  bar.Close,
				Volume: bar.Volume,
				Amount: bar.Amount,
			}
		} else {
			// 同一周，更新K线数据
			currentWeek.High = max(currentWeek.High, bar.High)
			currentWeek.Low = min(currentWeek.Low, bar.Low)
			currentWeek.Close = bar.Close // 收盘价取最后一天
			currentWeek.Volume += bar.Volume
			currentWeek.Amount += bar.Amount
			currentWeek.Date = bar.Date // 日期取最后一天
		}
	}

	// 添加最后一周
	if currentWeek != nil {
		weeklyBars = append(weeklyBars, currentWeek)
	}

	return weeklyBars, nil
}

// AggregateToMonthly 将日K线聚合为月K线
func AggregateToMonthly(dailyBars []*model.DailyBar) ([]*model.MonthlyBar, error) {
	if len(dailyBars) == 0 {
		return nil, fmt.Errorf("daily bars is empty")
	}

	var monthlyBars []*model.MonthlyBar
	var currentMonth *model.MonthlyBar

	for _, bar := range dailyBars {
		// 解析日期
		date, err := time.Parse("2006-01-02", bar.Date)
		if err != nil {
			continue
		}

		year := date.Year()
		month := int(date.Month())

		// 新的一月，创建新的月K线
		if currentMonth == nil || currentMonth.Year != year || currentMonth.Month != month {
			if currentMonth != nil {
				monthlyBars = append(monthlyBars, currentMonth)
			}

			currentMonth = &model.MonthlyBar{
				Code:   bar.Code,
				Year:   year,
				Month:  month,
				Date:   bar.Date,
				Open:   bar.Open,
				High:   bar.High,
				Low:    bar.Low,
				Close:  bar.Close,
				Volume: bar.Volume,
				Amount: bar.Amount,
			}
		} else {
			// 同一月，更新K线数据
			currentMonth.High = max(currentMonth.High, bar.High)
			currentMonth.Low = min(currentMonth.Low, bar.Low)
			currentMonth.Close = bar.Close // 收盘价取最后一天
			currentMonth.Volume += bar.Volume
			currentMonth.Amount += bar.Amount
			currentMonth.Date = bar.Date // 日期取最后一天
		}
	}

	// 添加最后一月
	if currentMonth != nil {
		monthlyBars = append(monthlyBars, currentMonth)
	}

	return monthlyBars, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
