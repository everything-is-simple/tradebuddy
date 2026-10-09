package model

import "time"

// Instrument represents a stock/security
type Instrument struct {
	Code      string    `json:"code"`       // sh600519
	Name      string    `json:"name"`       // 贵州茅台
	Board     string    `json:"board"`      // 主板/创业板/科创板
	ListDate  string    `json:"list_date"`  // YYYY-MM-DD
	IsST      bool      `json:"is_st"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DailyBar represents a daily OHLCV bar (前复权)
type DailyBar struct {
	Code      string  `json:"code"`
	Date      string  `json:"date"` // YYYY-MM-DD
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
	Amount    float64 `json:"amount"`
	AdjFactor float64 `json:"adj_factor"` // 复权因子
	Source    string  `json:"source"`     // tdx/tencent/sina
}

// WeeklyBar represents aggregated weekly bar
type WeeklyBar struct {
	Code   string  `json:"code"`
	Year   int     `json:"year"`
	Week   int     `json:"week"` // ISO week
	Date   string  `json:"date"` // 周最后交易日
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Amount float64 `json:"amount"`
}

// MonthlyBar represents aggregated monthly bar
type MonthlyBar struct {
	Code   string  `json:"code"`
	Year   int     `json:"year"`
	Month  int     `json:"month"`
	Date   string  `json:"date"` // 月最后交易日
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Amount float64 `json:"amount"`
}
