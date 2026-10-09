package calc

import (
	"fmt"
	"math"
	"tradebuddy/internal/model"
)

// ATR 计算平均真实波幅（Average True Range）
// bars: K线数据（按时间升序）
// period: 周期（通常14）
// 返回：每个时点的ATR值
func ATR(bars []*model.DailyBar, period int) ([]float64, error) {
	if len(bars) < 2 {
		return nil, fmt.Errorf("need at least 2 bars for ATR")
	}

	if period <= 0 {
		return nil, fmt.Errorf("period must be positive")
	}

	// 计算真实波幅TR
	tr := make([]float64, len(bars))
	tr[0] = bars[0].High - bars[0].Low // 第一根K线的TR

	for i := 1; i < len(bars); i++ {
		// TR = max(H-L, |H-Cp|, |L-Cp|)
		// H: 当日最高价, L: 当日最低价, Cp: 前日收盘价
		hl := bars[i].High - bars[i].Low
		hcp := math.Abs(bars[i].High - bars[i-1].Close)
		lcp := math.Abs(bars[i].Low - bars[i-1].Close)

		tr[i] = max(hl, max(hcp, lcp))
	}

	// 计算ATR（使用Wilder平滑法，类似EMA）
	atr := make([]float64, len(bars))

	// 前period-1个点不足以计算ATR
	for i := 0; i < period-1; i++ {
		atr[i] = 0
	}

	// 第一个ATR是前period个TR的平均值
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr[period-1] = sum / float64(period)

	// 后续ATR使用Wilder平滑：ATR(t) = (ATR(t-1) * (period-1) + TR(t)) / period
	for i := period; i < len(bars); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return atr, nil
}

// ATRLast 计算最后一个ATR值
func ATRLast(bars []*model.DailyBar, period int) (float64, error) {
	if len(bars) < period+1 {
		return 0, fmt.Errorf("not enough bars: need %d, got %d", period+1, len(bars))
	}

	atr, err := ATR(bars, period)
	if err != nil {
		return 0, err
	}

	return atr[len(atr)-1], nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
