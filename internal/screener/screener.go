package screener

import (
	"context"
	"fmt"
	"tradebuddy/internal/calc"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/store"
)

// Screener 强势股初选引擎
type Screener struct {
	store   *store.Store
	dataSource *datasource.Manager
}

// Config 筛选配置
type Config struct {
	MinPctChange   float64 // 最小涨幅（%）
	LookbackDays   int     // 回溯天数（用于计算20日高点）
	DD52Threshold  float64 // 距52周高点阈值（%）
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MinPctChange:   6.0,  // 当日涨幅 ≥ 6%
		LookbackDays:   20,   // 20个交易日
		DD52Threshold:  25.0, // 距52周高点 ≤ 25%
	}
}

// New 创建Screener实例
func New(st *store.Store, ds *datasource.Manager) *Screener {
	return &Screener{
		store:   st,
		dataSource: ds,
	}
}

// Run 运行筛选（主入口）
// tradeDate: 交易日期（YYYY-MM-DD）
func (s *Screener) Run(ctx context.Context, tradeDate string, cfg *Config) ([]*model.ScreenResult, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 1. 获取所有股票代码
	codes, err := s.dataSource.ListAvailableStocks()
	if err != nil {
		return nil, fmt.Errorf("failed to list stocks: %w", err)
	}

	fmt.Printf("开始筛选，总股票数：%d\n", len(codes))

	// 2. 对每只股票进行筛选
	var results []*model.ScreenResult

	for i, code := range codes {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// 进度提示（每100只）
		if i > 0 && i%100 == 0 {
			fmt.Printf("已处理：%d/%d\n", i, len(codes))
		}

		// 筛选单只股票
		result, err := s.screenOne(ctx, code, tradeDate, cfg)
		if err != nil {
			// 跳过错误的股票（如数据不足）
			continue
		}

		if result != nil {
			results = append(results, result)
		}
	}

	fmt.Printf("筛选完成，符合条件：%d只\n", len(results))

	// 3. 保存结果到数据库
	if len(results) > 0 {
		if err := s.store.SaveScreenResults(ctx, tradeDate, results); err != nil {
			return nil, fmt.Errorf("failed to save results: %w", err)
		}
	}

	return results, nil
}

// screenOne 筛选单只股票
func (s *Screener) screenOne(ctx context.Context, code string, tradeDate string, cfg *Config) (*model.ScreenResult, error) {
	// 获取足够的历史数据（需要252个交易日用于52周计算）
	bars, err := s.dataSource.GetDailyBars(ctx, code, 300)
	if err != nil {
		return nil, err
	}

	if len(bars) < 100 {
		return nil, fmt.Errorf("not enough bars: %d", len(bars))
	}

	// 找到指定交易日的索引
	targetIdx := -1
	for i := len(bars) - 1; i >= 0; i-- {
		if bars[i].Date == tradeDate {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return nil, fmt.Errorf("trade date %s not found", tradeDate)
	}

	// 当前K线
	currentBar := bars[targetIdx]

	// 前一日K线（计算涨幅）
	if targetIdx == 0 {
		return nil, fmt.Errorf("no previous bar")
	}
	prevBar := bars[targetIdx-1]

	// 条件A：当日涨幅 ≥ 6%
	pctChange := calc.PctChange(currentBar.Close, prevBar.Close)
	if pctChange < cfg.MinPctChange {
		return nil, nil // 不符合条件A
	}

	// 条件B：突破20日最高价
	// 需要前20+1天的数据
	if targetIdx < cfg.LookbackDays {
		return nil, fmt.Errorf("not enough history for lookback")
	}

	lookbackBars := bars[targetIdx-cfg.LookbackDays : targetIdx] // 前20天（不含今天）
	highest20, err := calc.HighestHigh(lookbackBars, cfg.LookbackDays)
	if err != nil {
		return nil, err
	}

	if currentBar.High <= highest20 {
		return nil, nil // 不符合条件B
	}

	// 条件C：距52周高点 ≤ 25%
	// 52周 = 约252个交易日
	weeks52 := 252
	if targetIdx < weeks52-1 {
		weeks52 = targetIdx + 1 // 上市不足52周，用全部历史
	}

	bars52w := bars[targetIdx-weeks52+1 : targetIdx+1]
	highest52w, err := calc.HighestHigh(bars52w, len(bars52w))
	if err != nil {
		return nil, err
	}

	dd52 := calc.PctChange(currentBar.High, highest52w)
	if dd52 < -cfg.DD52Threshold {
		return nil, nil // 不符合条件C
	}

	// 聚合周线和月线（用于后续计算MA）
	weeklyBars, _ := calc.AggregateToWeekly(bars[:targetIdx+1])
	monthlyBars, _ := calc.AggregateToMonthly(bars[:targetIdx+1])

	// 计算周线MA和月线MA（如果数据足够）
	var weeklyMA, monthlyMA float64
	if len(weeklyBars) >= 4 {
		// 取最后一根周线的收盘价作为周均线（简化）
		weeklyMA = weeklyBars[len(weeklyBars)-1].Close
	}
	if len(monthlyBars) >= 1 {
		monthlyMA = monthlyBars[len(monthlyBars)-1].Close
	}

	// 获取股票名称（从instrument表）
	inst, err := s.store.GetInstrument(ctx, code)
	var name string
	if err == nil {
		name = inst.Name
	} else {
		name = code
	}

	// 构建筛选结果
	result := &model.ScreenResult{
		TradeDate:   tradeDate,
		Code:        code,
		Name:        name,
		PctChange:   pctChange,
		ClosePrice:  currentBar.Close,
		HighPrice:   currentBar.High,
		Volume:      currentBar.Volume,
		Amount:      currentBar.Amount,
		Turnover:    0, // 换手率暂不计算
		DD52:        dd52,
		WeeklyMA:    weeklyMA,
		MonthlyMA:   monthlyMA,
		DataSource:  currentBar.Source,
	}

	return result, nil
}
