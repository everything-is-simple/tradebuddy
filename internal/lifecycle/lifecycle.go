package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"tradebuddy/internal/calc"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/store"
)

// Lifecycle 生命周期分析引擎
type Lifecycle struct {
	store      *store.Store
	dataSource *datasource.Manager
}

// Config 配置
type Config struct {
	MinHistoryDays int     // 最少历史天数
	MinSampleSize  int     // 最小样本量
	SpanWeight     float64 // 持续时间权重
	RangeWeight    float64 // 价格幅度权重
	ATRWeight      float64 // ATR权重
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MinHistoryDays: 60,   // 至少60天历史
		MinSampleSize:  30,   // 最少30个样本
		SpanWeight:     0.2,  // 持续时间权重20%
		RangeWeight:    0.4,  // 价格幅度权重40%
		ATRWeight:      0.4,  // ATR权重40%
	}
}

// New 创建Lifecycle实例
func New(st *store.Store, ds *datasource.Manager) *Lifecycle {
	return &Lifecycle{
		store:      st,
		dataSource: ds,
	}
}

// Run 运行生命周期分析
func (lc *Lifecycle) Run(ctx context.Context, tradeDate string, screenResults []*model.ScreenResult, cfg *Config) ([]*model.LifecycleMetrics, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	fmt.Printf("开始生命周期分析，交易日期：%s，股票数：%d\n", tradeDate, len(screenResults))

	// 1. 计算每只股票的生命周期指标
	var metrics []*model.LifecycleMetrics

	for i, sr := range screenResults {
		if i > 0 && i%10 == 0 {
			fmt.Printf("已处理：%d/%d\n", i, len(screenResults))
		}

		m, err := lc.analyzeOne(ctx, sr, tradeDate, cfg)
		if err != nil {
			// 跳过错误的股票
			continue
		}

		if m != nil {
			metrics = append(metrics, m)
		}
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no valid lifecycle metrics")
	}

	// 2. 计算市场排名
	lc.calculateRanks(metrics, cfg)

	// 3. 计算综合评分
	lc.calculateScores(metrics, cfg)

	// 4. 保存到数据库
	if err := lc.store.SaveLifecycleMetrics(ctx, tradeDate, metrics); err != nil {
		return nil, fmt.Errorf("failed to save metrics: %w", err)
	}

	fmt.Printf("生命周期分析完成，有效股票：%d只\n", len(metrics))

	return metrics, nil
}

// analyzeOne 分析单只股票
func (lc *Lifecycle) analyzeOne(ctx context.Context, sr *model.ScreenResult, tradeDate string, cfg *Config) (*model.LifecycleMetrics, error) {
	// 获取历史K线数据
	bars, err := lc.dataSource.GetDailyBars(ctx, sr.Code, cfg.MinHistoryDays+50)
	if err != nil {
		return nil, err
	}

	if len(bars) < cfg.MinHistoryDays {
		return nil, fmt.Errorf("insufficient history: %d bars", len(bars))
	}

	// 找到交易日索引
	targetIdx := -1
	for i := len(bars) - 1; i >= 0; i-- {
		if bars[i].Date == tradeDate {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return nil, fmt.Errorf("trade date not found: %s", tradeDate)
	}

	// 计算指标
	m := &model.LifecycleMetrics{
		Code:       sr.Code,
		Name:       sr.Name,
		TradeDate:  tradeDate,
		DataSource: sr.DataSource,
	}

	// 1. 计算波段持续天数
	spanDays, breakoutIdx, err := lc.calcSpanDays(bars, targetIdx)
	if err != nil {
		return nil, err
	}
	m.SpanDays = spanDays
	m.DaysSince20DH = spanDays

	// 2. 计算价格幅度
	priceRange, priceRangePct := lc.calcPriceRange(bars, breakoutIdx, targetIdx)
	m.PriceRange = priceRange
	m.PriceRangePct = priceRangePct

	// 3. 计算ATR标准化
	atrNormalized, err := lc.calcATRNormalized(bars, targetIdx, priceRange)
	if err != nil {
		return nil, err
	}
	m.ATRNormalized = atrNormalized

	// 4. 计算结构位置
	m.DistFrom20DH = lc.calcDistFrom20DH(bars, breakoutIdx, targetIdx)
	m.DistFrom52WH = sr.DD52 // 从 Screener 结果中获取

	return m, nil
}

// calcSpanDays 计算波段持续天数
func (lc *Lifecycle) calcSpanDays(bars []*model.DailyBar, targetIdx int) (int, int, error) {
	// 向前寻找突破20日高点的那一天
	for i := targetIdx; i >= 20; i-- {
		// 当前bar
		currentBar := bars[i]

		// 前20天的最高价
		lookbackBars := bars[i-20 : i]
		highest20, err := calc.HighestHigh(lookbackBars, 20)
		if err != nil {
			continue
		}

		// 如果当前高点突破了前20天最高价
		if currentBar.High > highest20 {
			// 找到突破点
			spanDays := targetIdx - i
			return spanDays, i, nil
		}
	}

	// 如果没找到突破点，返回0
	return 0, targetIdx, fmt.Errorf("breakout not found")
}

// calcPriceRange 计算价格幅度
func (lc *Lifecycle) calcPriceRange(bars []*model.DailyBar, breakoutIdx, targetIdx int) (float64, float64) {
	// 突破前20天的最低价
	lookbackBars := bars[breakoutIdx-20 : breakoutIdx]
	lowPrice, _ := calc.LowestLow(lookbackBars, 20)

	// 当前价格
	currentPrice := bars[targetIdx].Close

	// 价格幅度
	priceRange := currentPrice - lowPrice
	priceRangePct := (priceRange / lowPrice) * 100

	return priceRange, priceRangePct
}

// calcATRNormalized 计算ATR标准化
func (lc *Lifecycle) calcATRNormalized(bars []*model.DailyBar, targetIdx int, priceRange float64) (float64, error) {
	// 计算14日ATR
	if targetIdx < 14 {
		return 0, fmt.Errorf("not enough bars for ATR")
	}

	atrBars := bars[:targetIdx+1]
	atr14, err := calc.ATRLast(atrBars, 14)
	if err != nil {
		return 0, err
	}

	if atr14 == 0 {
		return 0, fmt.Errorf("ATR is zero")
	}

	return priceRange / atr14, nil
}

// calcDistFrom20DH 计算距20日高点的距离
func (lc *Lifecycle) calcDistFrom20DH(bars []*model.DailyBar, breakoutIdx, targetIdx int) float64 {
	// 突破时的价格
	breakoutPrice := bars[breakoutIdx].High

	// 当前价格
	currentPrice := bars[targetIdx].Close

	// 距离（%）
	return ((currentPrice - breakoutPrice) / breakoutPrice) * 100
}

// calculateRanks 计算市场排名
func (lc *Lifecycle) calculateRanks(metrics []*model.LifecycleMetrics, cfg *Config) {
	if len(metrics) < cfg.MinSampleSize {
		// 样本不足，全部设为 -1
		for _, m := range metrics {
			m.SpanRank = -1.0
			m.RangeRank = -1.0
			m.ATRRank = -1.0
		}
		return
	}

	// 收集样本
	spanSample := make([]float64, len(metrics))
	rangeSample := make([]float64, len(metrics))
	atrSample := make([]float64, len(metrics))

	for i, m := range metrics {
		spanSample[i] = float64(m.SpanDays)
		rangeSample[i] = m.PriceRangePct
		atrSample[i] = m.ATRNormalized
	}

	// 计算每只股票的排名
	for _, m := range metrics {
		m.SpanRank = percentileRank(float64(m.SpanDays), spanSample)
		m.RangeRank = percentileRank(m.PriceRangePct, rangeSample)
		m.ATRRank = percentileRank(m.ATRNormalized, atrSample)
	}
}

// calculateScores 计算综合评分
func (lc *Lifecycle) calculateScores(metrics []*model.LifecycleMetrics, cfg *Config) {
	for _, m := range metrics {
		if m.SpanRank < 0 || m.RangeRank < 0 || m.ATRRank < 0 {
			// 排名无效
			m.LifecycleScore = -1
			m.Grade = "N/A"
			continue
		}

		// 加权平均
		score := cfg.SpanWeight*m.SpanRank*100 +
			cfg.RangeWeight*m.RangeRank*100 +
			cfg.ATRWeight*m.ATRRank*100

		m.LifecycleScore = score

		// 评级
		if score >= 80 {
			m.Grade = "A"
		} else if score >= 60 {
			m.Grade = "B"
		} else if score >= 40 {
			m.Grade = "C"
		} else {
			m.Grade = "D"
		}
	}

	// 按评分排序
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].LifecycleScore > metrics[j].LifecycleScore
	})
}

// percentileRank 计算百分位排名
func percentileRank(value float64, sample []float64) float64 {
	if len(sample) == 0 {
		return -1.0
	}

	count := 0
	for _, v := range sample {
		if v < value {
			count++
		}
	}

	return float64(count) / float64(len(sample))
}
