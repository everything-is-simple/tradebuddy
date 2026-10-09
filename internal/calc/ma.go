package calc

import (
	"fmt"
	"tradebuddy/internal/model"
)

// MA 计算简单移动平均线
// bars: K线数据（按时间升序）
// period: 周期（如5日、20日）
// 返回：每个时点的MA值（长度与bars相同，前period-1个值为0）
func MA(bars []*model.DailyBar, period int) ([]float64, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("bars is empty")
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive")
	}

	if period > len(bars) {
		return nil, fmt.Errorf("period %d exceeds bars length %d", period, len(bars))
	}

	result := make([]float64, len(bars))

	// 前period-1个点没有足够数据，设为0
	for i := 0; i < period-1; i++ {
		result[i] = 0
	}

	// 计算第一个MA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += bars[i].Close
	}
	result[period-1] = sum / float64(period)

	// 滑动窗口计算后续MA
	for i := period; i < len(bars); i++ {
		sum = sum - bars[i-period].Close + bars[i].Close
		result[i] = sum / float64(period)
	}

	return result, nil
}

// MALast 计算最后一个MA值
func MALast(bars []*model.DailyBar, period int) (float64, error) {
	if len(bars) < period {
		return 0, fmt.Errorf("not enough bars: need %d, got %d", period, len(bars))
	}

	sum := 0.0
	for i := len(bars) - period; i < len(bars); i++ {
		sum += bars[i].Close
	}

	return sum / float64(period), nil
}

// EMA 计算指数移动平均线
// alpha = 2 / (period + 1)
func EMA(bars []*model.DailyBar, period int) ([]float64, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("bars is empty")
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive")
	}

	result := make([]float64, len(bars))
	alpha := 2.0 / float64(period+1)

	// 第一个EMA用SMA初始化
	sum := 0.0
	for i := 0; i < period && i < len(bars); i++ {
		sum += bars[i].Close
		result[i] = 0
	}

	if len(bars) >= period {
		result[period-1] = sum / float64(period)

		// 后续使用指数平滑
		for i := period; i < len(bars); i++ {
			result[i] = alpha*bars[i].Close + (1-alpha)*result[i-1]
		}
	}

	return result, nil
}

// HighestHigh 计算指定周期内的最高价
func HighestHigh(bars []*model.DailyBar, period int) (float64, error) {
	if len(bars) < period {
		return 0, fmt.Errorf("not enough bars: need %d, got %d", period, len(bars))
	}

	high := bars[len(bars)-period].High
	for i := len(bars) - period + 1; i < len(bars); i++ {
		if bars[i].High > high {
			high = bars[i].High
		}
	}

	return high, nil
}

// LowestLow 计算指定周期内的最低价
func LowestLow(bars []*model.DailyBar, period int) (float64, error) {
	if len(bars) < period {
		return 0, fmt.Errorf("not enough bars: need %d, got %d", period, len(bars))
	}

	low := bars[len(bars)-period].Low
	for i := len(bars) - period + 1; i < len(bars); i++ {
		if bars[i].Low < low {
			low = bars[i].Low
		}
	}

	return low, nil
}

// PctChange 计算涨跌幅（百分比）
func PctChange(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return (current - previous) / previous * 100
}
