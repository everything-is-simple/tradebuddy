-- Migration 0001: Initial Schema
-- TradeBuddy Database Schema
-- Created: 2026-10-09

-- ============================================
-- 1. 股票元数据表
-- ============================================
CREATE TABLE IF NOT EXISTS instruments (
    code        TEXT PRIMARY KEY,   -- sh600519, sz000001
    name        TEXT NOT NULL,
    board       TEXT,               -- 主板/创业板/科创板
    list_date   TEXT,               -- YYYY-MM-DD
    is_st       INTEGER DEFAULT 0,  -- 0/1
    is_active   INTEGER DEFAULT 1,  -- 0/1
    updated_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 2. 日线K线数据表（前复权）
-- ============================================
CREATE TABLE IF NOT EXISTS bars_daily (
    code        TEXT NOT NULL,
    date        TEXT NOT NULL,      -- YYYY-MM-DD
    open        REAL NOT NULL,
    high        REAL NOT NULL,
    low         REAL NOT NULL,
    close       REAL NOT NULL,
    volume      INTEGER NOT NULL,
    amount      REAL,
    adj_factor  REAL DEFAULT 1.0,   -- 复权因子
    source      TEXT,               -- tdx/tencent/sina
    PRIMARY KEY (code, date)
) WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS idx_bars_code_date ON bars_daily(code, date DESC);
CREATE INDEX IF NOT EXISTS idx_bars_date ON bars_daily(date);

-- ============================================
-- 3. 19:00 强势股初选结果
-- ============================================
CREATE TABLE IF NOT EXISTS screen_results (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT,
    pct_change  REAL NOT NULL,      -- 当日涨幅%
    close_price REAL,
    high_price  REAL,
    volume      INTEGER,
    amount      REAL,
    turnover    REAL,               -- 换手率%
    dd52        REAL,               -- 距52周高点%
    weekly_ma   REAL,               -- 周均线值
    monthly_ma  REAL,               -- 月均线值
    data_source TEXT,
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (trade_date, code)
);

CREATE INDEX IF NOT EXISTS idx_screen_date ON screen_results(trade_date DESC);

-- ============================================
-- 4. 19:30 生命周期分析
-- ============================================
CREATE TABLE IF NOT EXISTS lifecycle_states (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    stage       TEXT NOT NULL,      -- 2·早期 / 3·做头 等
    stage_desc  TEXT,
    vol_class   TEXT,               -- 放量/平量/缩量
    vr13        REAL,               -- 量能比
    pos52       REAL,               -- 52周位置
    gain_lo52   REAL,               -- 距52周低点涨幅%
    dd_hi52     REAL,               -- 距52周高点回撤%
    survive_prob REAL,              -- 8周存活概率
    ci_low      REAL,               -- 置信区间下限
    ci_high     REAL,               -- 置信区间上限
    sample_size INTEGER,            -- 样本量
    raw_view    TEXT,               -- ±25%生命JSON
    evidence    TEXT,               -- 原始证据JSON
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (trade_date, code)
);

CREATE INDEX IF NOT EXISTS idx_lifecycle_date ON lifecycle_states(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_lifecycle_stage ON lifecycle_states(stage);

-- ============================================
-- 5. 20:00 立花交易计划
-- ============================================
CREATE TABLE IF NOT EXISTS tachibana_signals (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT,
    decision    TEXT NOT NULL,      -- 布网/观察/排除
    z_score     REAL,
    band_now    TEXT,               -- 当前回档带状态
    tier1_price REAL,
    tier2_price REAL,
    tier3_price REAL,
    stop_price  REAL,
    profit_price REAL,              -- 上沿止盈价
    shares_t1   INTEGER,
    shares_t2   INTEGER,
    shares_t3   INTEGER,
    notional    REAL,               -- 预计投入金额
    risk_amount REAL,               -- 风险预算金额
    reason      TEXT,
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (trade_date, code)
);

CREATE INDEX IF NOT EXISTS idx_tachibana_date ON tachibana_signals(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_tachibana_decision ON tachibana_signals(decision);

-- ============================================
-- 6. 跟踪池持仓
-- ============================================
CREATE TABLE IF NOT EXISTS tracker_positions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT NOT NULL,
    name            TEXT,
    added_date      TEXT NOT NULL,
    frozen_at       TEXT,           -- 参数冻结时间
    tier1_price     REAL,
    tier2_price     REAL,
    tier3_price     REAL,
    stop_price      REAL,
    profit_price    REAL,
    tier1_shares    INTEGER,
    tier2_shares    INTEGER,
    tier3_shares    INTEGER,
    tier1_status    TEXT DEFAULT 'pending',  -- pending/filled
    tier2_status    TEXT DEFAULT 'pending',
    tier3_status    TEXT DEFAULT 'pending',
    tier1_filled_price REAL,       -- 实际成交价
    tier2_filled_price REAL,
    tier3_filled_price REAL,
    tier1_filled_date  TEXT,
    tier2_filled_date  TEXT,
    tier3_filled_date  TEXT,
    status          TEXT DEFAULT 'active',   -- active/stopped/expired/profit
    last_update     TEXT,
    expire_date     TEXT,
    notes           TEXT,
    UNIQUE(code, added_date)
);

CREATE INDEX IF NOT EXISTS idx_tracker_status ON tracker_positions(status);
CREATE INDEX IF NOT EXISTS idx_tracker_code ON tracker_positions(code);

-- ============================================
-- 7. 成交记录
-- ============================================
CREATE TABLE IF NOT EXISTS transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    trade_date      TEXT NOT NULL,
    code            TEXT NOT NULL,
    name            TEXT,
    direction       TEXT NOT NULL,  -- buy/sell
    planned_price   REAL,
    actual_price    REAL NOT NULL,
    quantity        INTEGER NOT NULL,
    amount          REAL,
    commission      REAL,
    slippage_pct    REAL,
    note            TEXT,           -- 梯①/梯②/止损/止盈
    source          TEXT,           -- manual/auto
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_code ON transactions(code);

-- ============================================
-- 8. 每日复盘报告
-- ============================================
CREATE TABLE IF NOT EXISTS reviews (
    trade_date      TEXT PRIMARY KEY,
    filled_count    INTEGER DEFAULT 0,
    stopped_count   INTEGER DEFAULT 0,
    avg_slippage    REAL,
    daily_return    REAL,
    equity          REAL,           -- 总权益
    cash            REAL,           -- 现金
    position_value  REAL,           -- 持仓市值
    notes           TEXT,           -- JSON详细信息
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 9. 任务执行日志（审计）
-- ============================================
CREATE TABLE IF NOT EXISTS job_runs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    job_type    TEXT NOT NULL,      -- screen/lifecycle/tachibana/review
    trade_date  TEXT,
    status      TEXT NOT NULL,      -- pending/running/success/failed
    started_at  TEXT,
    finished_at TEXT,
    duration_ms INTEGER,
    message     TEXT,
    stdout      TEXT,               -- JSON输出
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_jobs_type_date ON job_runs(job_type, trade_date DESC);

-- ============================================
-- 10. Schema版本管理
-- ============================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    applied_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

-- 记录本次迁移
INSERT INTO schema_migrations (version, name) VALUES (1, '0001_init');
