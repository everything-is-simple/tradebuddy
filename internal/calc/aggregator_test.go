package calc

import (
	"testing"
	"tradebuddy/internal/model"
)

func TestAggregateToWeekly(t *testing.T) {
	// 测试数据：一周5天的日K线
	dailyBars := []*model.DailyBar{
		{Code: "sh600519", Date: "2024-10-07", Open: 100, High: 105, Low: 99, Close: 103, Volume: 1000, Amount: 102000},  // 周一
		{Code: "sh600519", Date: "2024-10-08", Open: 103, High: 108, Low: 102, Close: 106, Volume: 1100, Amount: 116600}, // 周二
		{Code: "sh600519", Date: "2024-10-09", Open: 106, High: 110, Low: 104, Close: 108, Volume: 1200, Amount: 129600}, // 周三
		{Code: "sh600519", Date: "2024-10-10", Open: 108, High: 112, Low: 107, Close: 110, Volume: 1300, Amount: 143000}, // 周四
		{Code: "sh600519", Date: "2024-10-11", Open: 110, High: 115, Low: 109, Close: 112, Volume: 1400, Amount: 156800}, // 周五
	}

	weeklyBars, err := AggregateToWeekly(dailyBars)
	if err != nil {
		t.Fatalf("AggregateToWeekly failed: %v", err)
	}

	if len(weeklyBars) != 1 {
		t.Errorf("expected 1 weekly bar, got %d", len(weeklyBars))
	}

	week := weeklyBars[0]

	// 验证聚合结果
	if week.Open != 100 {
		t.Errorf("Open = %.2f, want 100", week.Open)
	}

	if week.High != 115 { // 周内最高
		t.Errorf("High = %.2f, want 115", week.High)
	}

	if week.Low != 99 { // 周内最低
		t.Errorf("Low = %.2f, want 99", week.Low)
	}

	if week.Close != 112 { // 周五收盘
		t.Errorf("Close = %.2f, want 112", week.Close)
	}

	expectedVolume := int64(1000 + 1100 + 1200 + 1300 + 1400)
	if week.Volume != expectedVolume {
		t.Errorf("Volume = %d, want %d", week.Volume, expectedVolume)
	}

	t.Logf("✓ Weekly bar aggregated: O=%.2f H=%.2f L=%.2f C=%.2f V=%d",
		week.Open, week.High, week.Low, week.Close, week.Volume)
}

func TestAggregateToMonthly(t *testing.T) {
	// 测试数据：一个月的部分日K线
	dailyBars := []*model.DailyBar{
		{Code: "sh600519", Date: "2024-10-01", Open: 100, High: 105, Low: 99, Close: 103, Volume: 1000},
		{Code: "sh600519", Date: "2024-10-08", Open: 103, High: 110, Low: 102, Close: 108, Volume: 1100},
		{Code: "sh600519", Date: "2024-10-15", Open: 108, High: 115, Low: 107, Close: 112, Volume: 1200},
		{Code: "sh600519", Date: "2024-10-22", Open: 112, High: 120, Low: 110, Close: 118, Volume: 1300},
		{Code: "sh600519", Date: "2024-10-29", Open: 118, High: 125, Low: 116, Close: 122, Volume: 1400},
	}

	monthlyBars, err := AggregateToMonthly(dailyBars)
	if err != nil {
		t.Fatalf("AggregateToMonthly failed: %v", err)
	}

	if len(monthlyBars) != 1 {
		t.Errorf("expected 1 monthly bar, got %d", len(monthlyBars))
	}

	month := monthlyBars[0]

	// 验证聚合结果
	if month.Open != 100 {
		t.Errorf("Open = %.2f, want 100", month.Open)
	}

	if month.High != 125 { // 月内最高
		t.Errorf("High = %.2f, want 125", month.High)
	}

	if month.Low != 99 { // 月内最低
		t.Errorf("Low = %.2f, want 99", month.Low)
	}

	if month.Close != 122 { // 最后一天收盘
		t.Errorf("Close = %.2f, want 122", month.Close)
	}

	if month.Year != 2024 || month.Month != 10 {
		t.Errorf("Year-Month = %d-%d, want 2024-10", month.Year, month.Month)
	}

	t.Logf("✓ Monthly bar aggregated: O=%.2f H=%.2f L=%.2f C=%.2f",
		month.Open, month.High, month.Low, month.Close)
}

func TestAggregateMultipleWeeks(t *testing.T) {
	// 测试数据：跨越两周
	dailyBars := []*model.DailyBar{
		{Code: "sh600519", Date: "2024-10-07", Open: 100, High: 105, Low: 99, Close: 103, Volume: 1000},  // 第1周
		{Code: "sh600519", Date: "2024-10-08", Open: 103, High: 108, Low: 102, Close: 106, Volume: 1100}, // 第1周
		{Code: "sh600519", Date: "2024-10-14", Open: 106, High: 110, Low: 104, Close: 108, Volume: 1200}, // 第2周
		{Code: "sh600519", Date: "2024-10-15", Open: 108, High: 112, Low: 107, Close: 110, Volume: 1300}, // 第2周
	}

	weeklyBars, err := AggregateToWeekly(dailyBars)
	if err != nil {
		t.Fatalf("AggregateToWeekly failed: %v", err)
	}

	if len(weeklyBars) != 2 {
		t.Errorf("expected 2 weekly bars, got %d", len(weeklyBars))
	}

	t.Logf("✓ Aggregated into %d weekly bars", len(weeklyBars))
}
