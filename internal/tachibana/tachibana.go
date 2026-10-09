package tachibana

import (
	"context"
	"fmt"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/store"
)

// Tachibana 立花提示引擎
type Tachibana struct {
	store      *store.Store
	dataSource *datasource.Manager
}

// New 创建Tachibana实例
func New(st *store.Store, ds *datasource.Manager) *Tachibana {
	return &Tachibana{
		store:      st,
		dataSource: ds,
	}
}

// Run 运行立花提示分析
func (t *Tachibana) Run(ctx context.Context, tradeDate string, lifecycleMetrics []*model.LifecycleMetrics) ([]*model.TachibanaSignal, error) {
	fmt.Printf("开始立花提示分析，交易日期：%s，股票数：%d\n", tradeDate, len(lifecycleMetrics))

	var signals []*model.TachibanaSignal

	for i, m := range lifecycleMetrics {
		if i > 0 && i%20 == 0 {
			fmt.Printf("已处理：%d/%d\n", i, len(lifecycleMetrics))
		}

		signal := t.analyzeOne(ctx, m, tradeDate)
		if signal != nil {
			signals = append(signals, signal)
		}
	}

	// 保存到数据库
	if len(signals) > 0 {
		if err := t.store.SaveTachibanaSignals(ctx, tradeDate, signals); err != nil {
			return nil, fmt.Errorf("failed to save signals: %w", err)
		}
	}

	fmt.Printf("立花提示分析完成，信号数：%d\n", len(signals))

	return signals, nil
}

// analyzeOne 分析单只股票
func (t *Tachibana) analyzeOne(ctx context.Context, m *model.LifecycleMetrics, tradeDate string) *model.TachibanaSignal {
	signal := &model.TachibanaSignal{
		Code:           m.Code,
		Name:           m.Name,
		TradeDate:      tradeDate,
		LifecycleGrade: m.Grade,
		LifecycleScore: m.LifecycleScore,
		SpanDays:       m.SpanDays,
		DistFrom20DH:   m.DistFrom20DH,
		ATRNormalized:  m.ATRNormalized,
		DataSource:     m.DataSource,
	}

	// 决策优先级：
	// 1. 节奏失败 -> 2. 分批减仓 -> 3. 同向加码 -> 4. 试探建仓 -> 5. 等待观望

	// 1. 检查节奏失败
	if t.checkExitOnRhythmFailure(m) {
		t.fillExitOnRhythmFailure(signal, m)
		return signal
	}

	// 2. 检查分批减仓
	if t.checkDistributionReduce(m) {
		t.fillDistributionReduce(signal, m)
		return signal
	}

	// 3. 检查同向加码
	if t.checkTrendConfirmationAdd(m) {
		t.fillTrendConfirmationAdd(signal, m)
		return signal
	}

	// 4. 检查试探建仓
	if t.checkTrendProbeEntry(m) {
		t.fillTrendProbeEntry(signal, m)
		return signal
	}

	// 5. 默认等待观望
	t.fillWaitNoAction(signal, m)
	return signal
}

// ============================================
// 决策规则检查
// ============================================

// checkTrendProbeEntry 检查试探建仓条件
func (t *Tachibana) checkTrendProbeEntry(m *model.LifecycleMetrics) bool {
	// 条件：
	// 1. Grade = A 或 B
	// 2. 距20日高点 -10% 到 0%（轻度回踩）
	// 3. ATR标准化 > 2.0
	// 4. 波段天数 < 20

	if m.Grade != "A" && m.Grade != "B" {
		return false
	}

	if m.DistFrom20DH < -10.0 || m.DistFrom20DH > 0 {
		return false
	}

	if m.ATRNormalized < 2.0 {
		return false
	}

	if m.SpanDays >= 20 {
		return false
	}

	return true
}

// checkTrendConfirmationAdd 检查同向加码条件
func (t *Tachibana) checkTrendConfirmationAdd(m *model.LifecycleMetrics) bool {
	// 条件：
	// 1. Grade = A
	// 2. 距20日高点 > 0（创新高）
	// 3. ATR Rank > 0.7
	// 4. 波段天数 5-30

	if m.Grade != "A" {
		return false
	}

	if m.DistFrom20DH <= 0 {
		return false
	}

	if m.ATRRank < 0.7 {
		return false
	}

	if m.SpanDays < 5 || m.SpanDays > 30 {
		return false
	}

	return true
}

// checkDistributionReduce 检查分批减仓条件
func (t *Tachibana) checkDistributionReduce(m *model.LifecycleMetrics) bool {
	// 条件：
	// 1. 波段天数 > 30
	// 2. 距52周高点 > -5%
	// 3. ATR Rank < 0.3
	// 4. 价格幅度 > 40%

	if m.SpanDays <= 30 {
		return false
	}

	if m.DistFrom52WH < -5.0 {
		return false
	}

	if m.ATRRank >= 0.3 {
		return false
	}

	if m.PriceRangePct <= 40.0 {
		return false
	}

	return true
}

// checkExitOnRhythmFailure 检查节奏失败条件
func (t *Tachibana) checkExitOnRhythmFailure(m *model.LifecycleMetrics) bool {
	// 条件：
	// 1. Grade = C 或 D
	// 2. 距20日高点 < -15%（深度回踩）
	// 3. ATR Rank < 0.3 或 Span Rank < 0.3

	if m.Grade != "C" && m.Grade != "D" {
		return false
	}

	if m.DistFrom20DH >= -15.0 {
		return false
	}

	if m.ATRRank >= 0.3 && m.SpanRank >= 0.3 {
		return false
	}

	return true
}

// ============================================
// 填充信号信息
// ============================================

// fillTrendProbeEntry 填充试探建仓信号
func (t *Tachibana) fillTrendProbeEntry(signal *model.TachibanaSignal, m *model.LifecycleMetrics) {
	signal.SignalType = model.SignalTrendProbeEntry
	signal.Title = "【试探建仓机会】"

	// 计算买入区间（基于ATR）
	atr14 := m.PriceRange / m.ATRNormalized
	signal.EntryZoneLow = m.ClosePrice - 1.5*atr14
	signal.EntryZoneHigh = m.ClosePrice - 0.5*atr14
	signal.StopLoss = m.ClosePrice - 2.0*atr14
	signal.CurrentPrice = m.ClosePrice

	// 置信度
	if m.Grade == "A" {
		signal.Confidence = model.ConfidenceHigh
	} else {
		signal.Confidence = model.ConfidenceMedium
	}

	// 描述
	signal.Description = fmt.Sprintf("波段处于健康推进状态（评级%s，%d天），当前轻度回踩%.2f%%，可考虑小仓试探",
		m.Grade, m.SpanDays, m.DistFrom20DH)

	// 建议
	signal.Suggestion = "小仓试探，不追高。回踩到买入区间时分批建仓。设置止损，保持纪律。"

	// 风险
	signal.Risk = "不要重仓，分批建仓。破止损坚决退出。"
}

// fillTrendConfirmationAdd 填充同向加码信号
func (t *Tachibana) fillTrendConfirmationAdd(signal *model.TachibanaSignal, m *model.LifecycleMetrics) {
	signal.SignalType = model.SignalTrendConfirmationAdd
	signal.Title = "【同向加码机会】"
	signal.Confidence = model.ConfidenceHigh

	// 计算加码区间
	atr14 := m.PriceRange / m.ATRNormalized
	signal.EntryZoneLow = m.ClosePrice + 0.5*atr14
	signal.EntryZoneHigh = m.ClosePrice + 1.5*atr14
	signal.StopLoss = m.ClosePrice - 2.0*atr14
	signal.CurrentPrice = m.ClosePrice

	signal.Description = fmt.Sprintf("波段持续推进（%d天），创近期新高+%.2f%%，节奏健康，可考虑加码",
		m.SpanDays, m.DistFrom20DH)

	signal.Suggestion = "等待突破确认后加码。保持分批纪律，设止损。"

	signal.Risk = "保持分批纪律。破止损坚决退出。"
}

// fillDistributionReduce 填充分批减仓信号
func (t *Tachibana) fillDistributionReduce(signal *model.TachibanaSignal, m *model.LifecycleMetrics) {
	signal.SignalType = model.SignalDistributionReduce
	signal.Title = "【分批减仓提示】"
	signal.Confidence = model.ConfidenceMedium
	signal.CurrentPrice = m.ClosePrice

	signal.Description = fmt.Sprintf("波段持续%d天，涨幅%.2f%%，接近52周高点（%.2f%%），波动衰减",
		m.SpanDays, m.PriceRangePct, m.DistFrom52WH)

	signal.Suggestion = "分3-5次逐步减仓，保留观察仓。"

	signal.Risk = "不要一次性清仓。行情如果继续，可保留少量仓位。"
}

// fillExitOnRhythmFailure 填充节奏失败信号
func (t *Tachibana) fillExitOnRhythmFailure(signal *model.TachibanaSignal, m *model.LifecycleMetrics) {
	signal.SignalType = model.SignalExitOnRhythmFailure
	signal.Title = "【节奏失败警示】"
	signal.Confidence = model.ConfidenceHigh
	signal.CurrentPrice = m.ClosePrice

	signal.Description = fmt.Sprintf("波段推进力度减弱（评级%s），价格深度回踩%.2f%%，结构疲弱",
		m.Grade, m.DistFrom20DH)

	signal.Suggestion = "考虑退出或大幅减仓。"

	signal.Risk = "不要死扛，承认失败。及时止损。"
}

// fillWaitNoAction 填充等待观望信号
func (t *Tachibana) fillWaitNoAction(signal *model.TachibanaSignal, m *model.LifecycleMetrics) {
	signal.SignalType = model.SignalWaitNoAction
	signal.Title = "【等待观望】"
	signal.Confidence = model.ConfidenceLow
	signal.CurrentPrice = m.ClosePrice

	var reason string
	if m.Grade == "N/A" {
		reason = "样本不足，结构不清晰"
	} else if m.SpanDays < 3 {
		reason = fmt.Sprintf("波段太新（%d天）", m.SpanDays)
	} else {
		reason = "结构位置不够明确"
	}

	signal.Description = fmt.Sprintf("当前%s，不强迫市场给出机会", reason)

	signal.Suggestion = "继续观察，不急于行动。耐心等待更好的时机。"

	signal.Risk = "不要勉强交易。"
}
