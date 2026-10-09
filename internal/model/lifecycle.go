package model

// LifecycleMetrics 生命周期指标
type LifecycleMetrics struct {
	Code      string // 股票代码
	Name      string // 股票名称
	TradeDate string // 交易日期（YYYY-MM-DD）

	// 时间维度
	SpanDays      int // 波段持续天数（从突破20日高到现在）
	DaysSince20DH int // 距离突破20日高点的天数

	// 价格维度
	PriceRange    float64 // 价格幅度（当前价 - 低点价）
	PriceRangePct float64 // 价格幅度百分比
	ATRNormalized float64 // ATR标准化幅度（price_range / atr14）

	// 结构位置
	DistFrom20DH float64 // 距20日高点（%）
	DistFrom52WH float64 // 距52周高点（%）

	// 排名指标（百分位，0-1之间）
	SpanRank  float64 // 持续时间排名
	RangeRank float64 // 价格幅度排名
	ATRRank   float64 // ATR标准化排名

	// 综合评分
	LifecycleScore float64 // 生命周期综合评分（0-100）
	Grade          string  // 评级（A/B/C/D）

	// 元数据
	DataSource string // 数据来源
}
