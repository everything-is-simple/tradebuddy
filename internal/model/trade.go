package model

import "time"

// ScreenResult represents 19:00 screener output
type ScreenResult struct {
	TradeDate   string    `json:"trade_date"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	PctChange   float64   `json:"pct_change"`   // 当日涨幅%
	ClosePrice  float64   `json:"close_price"`
	HighPrice   float64   `json:"high_price"`
	Volume      int64     `json:"volume"`
	Amount      float64   `json:"amount"`
	Turnover    float64   `json:"turnover"`     // 换手率%
	DD52        float64   `json:"dd52"`         // 距52周高点%
	WeeklyMA    float64   `json:"weekly_ma"`    // 周均线值
	MonthlyMA   float64   `json:"monthly_ma"`   // 月均线值
	DataSource  string    `json:"data_source"`
	CreatedAt   time.Time `json:"created_at"`
}

// LifecycleState represents 19:30 lifecycle analysis output
type LifecycleState struct {
	TradeDate    string    `json:"trade_date"`
	Code         string    `json:"code"`
	Stage        string    `json:"stage"`         // 2·早期 / 3·做头 等
	StageDesc    string    `json:"stage_desc"`
	VolClass     string    `json:"vol_class"`     // 放量/平量/缩量
	VR13         float64   `json:"vr13"`
	Pos52        float64   `json:"pos52"`         // 52周位置
	GainLo52     float64   `json:"gain_lo52"`     // 距52周低点涨幅%
	DDHI52       float64   `json:"dd_hi52"`       // 距52周高点回撤%
	SurviveProb  float64   `json:"survive_prob"`  // 8周存活概率
	CILow        float64   `json:"ci_low"`
	CIHigh       float64   `json:"ci_high"`
	SampleSize   int       `json:"sample_size"`
	RawView      string    `json:"raw_view"`      // ±25%生命JSON
	Evidence     string    `json:"evidence"`      // 原始证据JSON
	CreatedAt    time.Time `json:"created_at"`
}

// TachibanaSignal represents 20:00 Tachibana trading signal (simplified Phase 1)
type TachibanaSignal struct {
	TradeDate   string    `json:"trade_date"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`

	// 信号类型
	SignalType  string    `json:"signal_type"`  // trend_probe_entry / trend_confirmation_add / etc
	Confidence  string    `json:"confidence"`   // High / Medium / Low

	// 价格区间
	CurrentPrice  float64 `json:"current_price"`
	EntryZoneLow  float64 `json:"entry_zone_low"`
	EntryZoneHigh float64 `json:"entry_zone_high"`
	StopLoss      float64 `json:"stop_loss"`

	// 依据指标
	LifecycleGrade string  `json:"lifecycle_grade"`
	LifecycleScore float64 `json:"lifecycle_score"`
	SpanDays       int     `json:"span_days"`
	DistFrom20DH   float64 `json:"dist_from_20dh"`
	ATRNormalized  float64 `json:"atr_normalized"`

	// 提示信息
	Title       string `json:"title"`
	Description string `json:"description"`
	Risk        string `json:"risk"`
	Suggestion  string `json:"suggestion"`

	DataSource  string    `json:"data_source"`
	CreatedAt   time.Time `json:"created_at"`
}

// SignalType 常量
const (
	SignalTrendProbeEntry      = "trend_probe_entry"
	SignalTrendConfirmationAdd = "trend_confirmation_add"
	SignalDistributionReduce   = "distribution_reduce"
	SignalExitOnRhythmFailure  = "exit_on_rhythm_failure"
	SignalWaitNoAction         = "wait_no_action"
)

// Confidence 常量
const (
	ConfidenceHigh   = "High"
	ConfidenceMedium = "Medium"
	ConfidenceLow    = "Low"
)

// TrackerPosition represents a position in tracking pool
type TrackerPosition struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	AddedDate       string    `json:"added_date"`
	FrozenAt        time.Time `json:"frozen_at"`
	Tier1Price      float64   `json:"tier1_price"`
	Tier2Price      float64   `json:"tier2_price"`
	Tier3Price      float64   `json:"tier3_price"`
	StopPrice       float64   `json:"stop_price"`
	ProfitPrice     float64   `json:"profit_price"`
	Tier1Shares     int       `json:"tier1_shares"`
	Tier2Shares     int       `json:"tier2_shares"`
	Tier3Shares     int       `json:"tier3_shares"`
	Tier1Status     string    `json:"tier1_status"` // pending/filled
	Tier2Status     string    `json:"tier2_status"`
	Tier3Status     string    `json:"tier3_status"`
	Tier1FilledPrice float64  `json:"tier1_filled_price"`
	Tier2FilledPrice float64  `json:"tier2_filled_price"`
	Tier3FilledPrice float64  `json:"tier3_filled_price"`
	Tier1FilledDate  string   `json:"tier1_filled_date"`
	Tier2FilledDate  string   `json:"tier2_filled_date"`
	Tier3FilledDate  string   `json:"tier3_filled_date"`
	Status          string    `json:"status"` // active/stopped/expired/profit
	LastUpdate      time.Time `json:"last_update"`
	ExpireDate      string    `json:"expire_date"`
	Notes           string    `json:"notes"`
}

// Transaction represents an executed trade
type Transaction struct {
	ID            int64     `json:"id"`
	TradeDate     string    `json:"trade_date"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Direction     string    `json:"direction"` // buy/sell
	PlannedPrice  float64   `json:"planned_price"`
	ActualPrice   float64   `json:"actual_price"`
	Quantity      int       `json:"quantity"`
	Amount        float64   `json:"amount"`
	Commission    float64   `json:"commission"`
	SlippagePct   float64   `json:"slippage_pct"`
	Note          string    `json:"note"`
	Source        string    `json:"source"` // manual/auto
	CreatedAt     time.Time `json:"created_at"`
}

// Review represents daily review report
type Review struct {
	TradeDate     string    `json:"trade_date"`
	FilledCount   int       `json:"filled_count"`
	StoppedCount  int       `json:"stopped_count"`
	AvgSlippage   float64   `json:"avg_slippage"`
	DailyReturn   float64   `json:"daily_return"`
	Equity        float64   `json:"equity"`
	Cash          float64   `json:"cash"`
	PositionValue float64   `json:"position_value"`
	Notes         string    `json:"notes"` // JSON详细信息
	CreatedAt     time.Time `json:"created_at"`
}

// JobRun represents a task execution record
type JobRun struct {
	ID         int64     `json:"id"`
	JobType    string    `json:"job_type"` // screen/lifecycle/tachibana/review
	TradeDate  string    `json:"trade_date"`
	Status     string    `json:"status"` // pending/running/success/failed
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	DurationMS int64     `json:"duration_ms"`
	Message    string    `json:"message"`
	Stdout     string    `json:"stdout"` // JSON输出
	CreatedAt  time.Time `json:"created_at"`
}
