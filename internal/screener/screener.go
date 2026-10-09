package screener

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tradebuddy/internal/calc"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/reporter"
	"tradebuddy/internal/store"
)

// Screener 强势股初选引擎
type Screener struct {
	store      *store.Store
	dataSource *datasource.Manager
	reporter   *reporter.ScreenReporter
}

// Config 筛选配置
type Config struct {
	MinPctChange   float64 // 最小涨幅（%）
	LookbackDays   int     // 回溯天数（用于计算20日高点）
	DD52Threshold  float64 // 距52周高点阈值（%）
	BatchSize      int     // 批处理大小
	EnableParallel bool    // 启用并行处理
	MaxWorkers     int     // 最大工作协程数
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MinPctChange:   6.0,   // 当日涨幅 ≥ 6%
		LookbackDays:   20,    // 20个交易日
		DD52Threshold:  25.0,  // 距52周高点 ≤ 25%
		BatchSize:      100,   // 每批100只
		EnableParallel: false, // 默认不启用并行（避免复杂性）
		MaxWorkers:     4,     // 最多4个工作协程
	}
}

// New 创建Screener实例
func New(st *store.Store, ds *datasource.Manager) *Screener {
	return &Screener{
		store:      st,
		dataSource: ds,
		reporter:   reporter.NewScreenReporter(),
	}
}

// ProgressCallback 进度回调函数
type ProgressCallback func(current, total int, elapsed time.Duration)

// Run 运行筛选（主入口）
func (s *Screener) Run(ctx context.Context, tradeDate string, cfg *Config) ([]*model.ScreenResult, error) {
	return s.RunWithProgress(ctx, tradeDate, cfg, nil)
}

// RunWithProgress 运行筛选（带进度回调）
func (s *Screener) RunWithProgress(ctx context.Context, tradeDate string, cfg *Config, progressFn ProgressCallback) ([]*model.ScreenResult, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	startTime := time.Now()

	// 1. 获取所有股票代码
	codes, err := s.dataSource.ListAvailableStocks()
	if err != nil {
		return nil, fmt.Errorf("failed to list stocks: %w", err)
	}

	fmt.Printf("开始筛选，交易日期：%s，总股票数：%d\n", tradeDate, len(codes))

	// 2. 筛选股票
	var results []*model.ScreenResult
	if cfg.EnableParallel {
		results, err = s.runParallel(ctx, codes, tradeDate, cfg, progressFn)
	} else {
		results, err = s.runSerial(ctx, codes, tradeDate, cfg, progressFn)
	}

	if err != nil {
		return nil, err
	}

	elapsed := time.Since(startTime)
	fmt.Printf("筛选完成，符合条件：%d只，耗时：%v\n", len(results), elapsed)

	// 3. 保存结果到数据库
	if len(results) > 0 {
		if err := s.store.SaveScreenResults(ctx, tradeDate, results); err != nil {
			return nil, fmt.Errorf("failed to save results: %w", err)
		}
	}

	return results, nil
}

// runSerial 串行处理
func (s *Screener) runSerial(ctx context.Context, codes []string, tradeDate string, cfg *Config, progressFn ProgressCallback) ([]*model.ScreenResult, error) {
	var results []*model.ScreenResult
	startTime := time.Now()

	for i, code := range codes {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// 进度回调
		if progressFn != nil && i%10 == 0 {
			progressFn(i, len(codes), time.Since(startTime))
		}

		// 进度提示（每100只）
		if i > 0 && i%100 == 0 {
			fmt.Printf("已处理：%d/%d (%.1f%%)\n", i, len(codes), float64(i)/float64(len(codes))*100)
		}

		// 筛选单只股票
		result, err := s.screenOne(ctx, code, tradeDate, cfg)
		if err != nil {
			continue
		}

		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// runParallel 并行处理（性能优化）
func (s *Screener) runParallel(ctx context.Context, codes []string, tradeDate string, cfg *Config, progressFn ProgressCallback) ([]*model.ScreenResult, error) {
	var (
		results   []*model.ScreenResult
		resultsMu sync.Mutex
		wg        sync.WaitGroup
		startTime = time.Now()
	)

	// 创建工作队列
	jobs := make(chan string, cfg.BatchSize)

	// 启动工作协程
	for i := 0; i < cfg.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for code := range jobs {
				result, err := s.screenOne(ctx, code, tradeDate, cfg)
				if err == nil && result != nil {
					resultsMu.Lock()
					results = append(results, result)
					resultsMu.Unlock()
				}
			}
		}()
	}

	// 分发任务
	go func() {
		for i, code := range codes {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- code:
				// 进度回调
				if progressFn != nil && i%10 == 0 {
					progressFn(i, len(codes), time.Since(startTime))
				}
			}
		}
		close(jobs)
	}()

	// 等待完成
	wg.Wait()

	return results, nil
}

// RunAndExport 运行筛选并导出Excel报告
func (s *Screener) RunAndExport(ctx context.Context, tradeDate string, cfg *Config, outputPath string) error {
	// 运行筛选
	results, err := s.Run(ctx, tradeDate, cfg)
	if err != nil {
		return fmt.Errorf("screen failed: %w", err)
	}

	// 生成报告
	if err := s.reporter.GenerateScreenReport(tradeDate, results, outputPath); err != nil {
		return fmt.Errorf("generate report failed: %w", err)
	}

	fmt.Printf("报告已生成：%s\n", outputPath)
	return nil
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
		return nil, nil
	}

	// 条件B：突破20日最高价
	if targetIdx < cfg.LookbackDays {
		return nil, fmt.Errorf("not enough history for lookback")
	}

	lookbackBars := bars[targetIdx-cfg.LookbackDays : targetIdx]
	highest20, err := calc.HighestHigh(lookbackBars, cfg.LookbackDays)
	if err != nil {
		return nil, err
	}

	if currentBar.High <= highest20 {
		return nil, nil
	}

	// 条件C：距52周高点 ≤ 25%
	weeks52 := 252
	if targetIdx < weeks52-1 {
		weeks52 = targetIdx + 1
	}

	bars52w := bars[targetIdx-weeks52+1 : targetIdx+1]
	highest52w, err := calc.HighestHigh(bars52w, len(bars52w))
	if err != nil {
		return nil, err
	}

	dd52 := calc.PctChange(currentBar.High, highest52w)
	if dd52 < -cfg.DD52Threshold {
		return nil, nil
	}

	// 聚合周线和月线
	weeklyBars, _ := calc.AggregateToWeekly(bars[:targetIdx+1])
	monthlyBars, _ := calc.AggregateToMonthly(bars[:targetIdx+1])

	var weeklyMA, monthlyMA float64
	if len(weeklyBars) >= 4 {
		weeklyMA = weeklyBars[len(weeklyBars)-1].Close
	}
	if len(monthlyBars) >= 1 {
		monthlyMA = monthlyBars[len(monthlyBars)-1].Close
	}

	// 获取股票名称
	inst, err := s.store.GetInstrument(ctx, code)
	var name string
	if err == nil {
		name = inst.Name
	} else {
		name = code
	}

	// 构建结果
	result := &model.ScreenResult{
		TradeDate:  tradeDate,
		Code:       code,
		Name:       name,
		PctChange:  pctChange,
		ClosePrice: currentBar.Close,
		HighPrice:  currentBar.High,
		Volume:     currentBar.Volume,
		Amount:     currentBar.Amount,
		Turnover:   0,
		DD52:       dd52,
		WeeklyMA:   weeklyMA,
		MonthlyMA:  monthlyMA,
		DataSource: currentBar.Source,
	}

	return result, nil
}
