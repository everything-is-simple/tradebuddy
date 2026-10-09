-- Add tachibana_signals table
CREATE TABLE IF NOT EXISTS tachibana_signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trade_date TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT,

    -- 信号类型
    signal_type TEXT NOT NULL,
    confidence TEXT,

    -- 价格区间
    current_price REAL,
    entry_zone_low REAL,
    entry_zone_high REAL,
    stop_loss REAL,

    -- 依据指标
    lifecycle_grade TEXT,
    lifecycle_score REAL,
    span_days INTEGER,
    dist_from_20dh REAL,
    atr_normalized REAL,

    -- 提示信息
    title TEXT,
    description TEXT,
    risk TEXT,
    suggestion TEXT,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trade_date, code)
);

CREATE INDEX idx_tachibana_trade_date ON tachibana_signals(trade_date);
CREATE INDEX idx_tachibana_signal_type ON tachibana_signals(signal_type);
CREATE INDEX idx_tachibana_confidence ON tachibana_signals(confidence);
