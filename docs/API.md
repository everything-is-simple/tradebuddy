# TradeBuddy API 规范文档

> API Specification  
> Version 1.0 - 2026-10-09

---

## API 概览

**类型：** 内部API（非HTTP REST，Go函数调用）  
**风格：** 领域驱动设计（DDD）  
**错误处理：** Go标准error返回值  
**并发：** context.Context传递取消信号

---

## 模块架构

```
                    ┌─────────────┐
                    │   tb-cli    │ (CLI入口)
                    │ tb-desktop  │ (GUI入口)
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │Screener │      │ Lifecycle │     │ Tachibana │
   └────┬────┘      └─────┬─────┘     └─────┬─────┘
        │                 │                  │
        └─────────────────┼──────────────────┘
                          │
                    ┌─────▼─────┐
                    │   Store   │ (数据访问层)
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │  SQLite   │
                    └───────────┘
```

---

## 1. Store (数据访问层)

### 1.1 接口定义

```go
package store

import (
    "context"
    "tradebuddy/internal/model"
)

type Store interface {
    // 数据库管理
    Migrate() error
    Close() error
    
    // Instruments（股票元数据）
    UpsertInstrument(ctx context.Context, inst *model.Instrument) error
    GetInstrument(ctx context.Context, code string) (*model.Instrument, error)
    ListInstruments(ctx context.Context, filter InstrumentFilter) ([]*model.Instrument, error)
    
    // Bars（K线数据）
    InsertDailyBars(ctx context.Context, bars []*model.DailyBar) error
    GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error)
    GetDailyBarsRange(ctx context.Context, code string, startDate, endDate string) ([]*model.DailyBar, error)
    
    // ScreenResults（筛选结果）
    SaveScreenResults(ctx context.Context, tradeDate string, results []*model.ScreenResult) error
    LoadScreenResults(ctx context.Context, tradeDate string) ([]*model.ScreenResult, error)
    
    // LifecycleStates（生命周期）
    SaveLifecycleStates(ctx context.Context, states []*model.LifecycleState) error
    LoadLifecycleStates(ctx context.Context, tradeDate string) ([]*model.LifecycleState, error)
    
    // TachibanaSignals（立花信号）
    SaveTachibanaSignals(ctx context.Context, signals []*model.TachibanaSignal) error
    LoadTachibanaSignals(ctx context.Context, tradeDate string) ([]*model.TachibanaSignal, error)
    
    // TrackerPositions（跟踪池）
    InsertTrackerPosition(ctx context.Context, pos *model.TrackerPosition) error
    UpdateTrackerPosition(ctx context.Context, pos *model.TrackerPosition) error
    LoadActivePositions(ctx context.Context) ([]*model.TrackerPosition, error)
    LoadPositionsByCode(ctx context.Context, code string) ([]*model.TrackerPosition, error)
    
    // Transactions（成交记录）
    InsertTransaction(ctx context.Context, txn *model.Transaction) error
    GetTransactionsByDate(ctx context.Context, tradeDate string) ([]*model.Transaction, error)
    GetTransactionsByCode(ctx context.Context, code string) ([]*model.Transaction, error)
    
    // Reviews（复盘报告）
    SaveReview(ctx context.Context, review *model.Review) error
    GetReview(ctx context.Context, tradeDate string) (*model.Review, error)
    
    // JobRuns（任务执行记录）
    InsertJobRun(ctx context.Context, job *model.JobRun) error
    UpdateJobRun(ctx context.Context, job *model.JobRun) error
}
```

### 1.2 过滤器定义

```go
type InstrumentFilter struct {
    Board      string   // "主板", "创业板", "科创板"
    ExcludeST  bool     // 排除ST股票
    IsActive   bool     // 只返回活跃股票
    Codes      []string // 指定代码列表
}
```

### 1.3 错误定义

```go
var (
    ErrNotFound         = errors.New("record not found")
    ErrDuplicateKey     = errors.New("duplicate key")
    ErrInvalidInput     = errors.New("invalid input")
    ErrDBConnection     = errors.New("database connection error")
    ErrTransactionFailed = errors.New("transaction failed")
)
```

---

## 2. Screener (强势股初选)

### 2.1 接口定义

```go
package screener

import (
    "context"
    "tradebuddy/internal/model"
)

type Screener interface {
    // Run 执行19:00初选
    Run(ctx context.Context, tradeDate string) (*ScreenResult, error)
}

type ScreenResult struct {
    TradeDate  string
    Count      int
    Stocks     []*model.ScreenResult
    FilePath   string // Excel报告路径
    Duration   time.Duration
}

type ScreenerConfig struct {
    MinPctChange   float64 // 6.0（涨幅下限）
    LookbackDays   int     // 20（创新高回溯期）
    WeeklyMA       int     // 30（周均线周期）
    MonthlyMA      int     // 10（月均线周期）
    ExcludeST      bool    // true（排除ST）
    MinListDays    int     // 21（最小上市天数）
}
```

### 2.2 核心方法

```go
// NewScreener 创建Screener实例
func NewScreener(store store.Store, config *ScreenerConfig) Screener

// filterByPctChange 条件A：涨幅>6%
func (s *screener) filterByPctChange(candidates []*model.Instrument, tradeDate string) ([]*model.ScreenResult, error)

// filterBy20DayHigh 条件B：创20日新高
func (s *screener) filterBy20DayHigh(candidates []*model.ScreenResult) ([]*model.ScreenResult, error)

// filterByTrend 条件C：周月线趋势过滤
func (s *screener) filterByTrend(candidates []*model.ScreenResult) ([]*model.ScreenResult, error)

// exportToExcel 生成Excel报告
func (s *screener) exportToExcel(results []*model.ScreenResult, tradeDate string) (string, error)
```

---

## 3. Lifecycle (生命周期分析)

### 3.1 接口定义

```go
package lifecycle

import (
    "context"
    "tradebuddy/internal/model"
)

type Classifier interface {
    // Classify 分析单只股票的生命周期
    Classify(ctx context.Context, stock *model.ScreenResult, weeklyBars []*model.WeeklyBar) (*model.LifecycleState, error)
    
    // BatchClassify 批量分析（并发）
    BatchClassify(ctx context.Context, stocks []*model.ScreenResult) ([]*model.LifecycleState, error)
}

type LifecycleConfig struct {
    SwingWindow    int    // 4（摆动点窗口）
    ProbTablePath  string // 概率表JSON路径
}

// ProbabilityTable 概率表查询接口
type ProbabilityTable interface {
    Lookup(stage string, volClass string) (*ProbResult, error)
}

type ProbResult struct {
    SurviveProb float64 // 8周存活概率
    CILow       float64 // 置信区间下限
    CIHigh      float64 // 置信区间上限
    SampleSize  int     // 样本量
}
```

### 3.2 核心方法

```go
// NewClassifier 创建分类器
func NewClassifier(store store.Store, config *LifecycleConfig) (Classifier, error)

// findSwingHighs 识别摆动高点
func (c *classifier) findSwingHighs(bars []*model.WeeklyBar, window int) []int

// findSwingLows 识别摆动低点
func (c *classifier) findSwingLows(bars []*model.WeeklyBar, window int) []int

// analyzeStructure 判定结构（上升/下跌/无趋势）
func (c *classifier) analyzeStructure(highs, lows []int, bars []*model.WeeklyBar) string

// determineStage 确定九类阶段
func (c *classifier) determineStage(structure string, bars []*model.WeeklyBar) string

// calculateVR13 计算量能比
func (c *classifier) calculateVR13(bars []*model.WeeklyBar, period int) float64

// classifyVolume VR13分类
func (c *classifier) classifyVolume(vr13 float64) string

// analyzeRawLifespan ±25%生命视角
func (c *classifier) analyzeRawLifespan(bars []*model.WeeklyBar) string
```

---

## 4. Tachibana (立花交易计划)

### 4.1 接口定义

```go
package tachibana

import (
    "context"
    "tradebuddy/internal/model"
)

type Planner interface {
    // BuildPlan 生成单只股票的交易计划
    BuildPlan(ctx context.Context, 
        stock *model.ScreenResult,
        lifecycle *model.LifecycleState,
        dailyBars []*model.DailyBar) (*model.TachibanaSignal, error)
    
    // BatchBuildPlans 批量生成计划
    BatchBuildPlans(ctx context.Context, 
        stocks []*model.ScreenResult,
        lifecycles []*model.LifecycleState) ([]*model.TachibanaSignal, error)
}

type TachibanaConfig struct {
    Capital             float64   // 1000000（参考资金）
    RiskPerStockPct     float64   // 0.01（单股风险1%）
    TierRatios          []float64 // [0.4, 0.3, 0.3]（梯档配比）
    ZMin                float64   // 0.8（回档带下限）
    ZMax                float64   // 2.5（回档带上限）
    MaxNotionalPerStock float64   // 200000（单股市值上限）
}
```

### 4.2 核心方法

```go
// NewPlanner 创建交易计划器
func NewPlanner(store store.Store, tracker Tracker, config *TachibanaConfig) Planner

// isTradeableStage 可交易性过滤
func (p *planner) isTradeableStage(stage string) bool

// calculateZScore 计算回档深度
func (p *planner) calculateZScore(ma20, currentPrice, atr14 float64) float64

// calculateTierPrices 计算梯价
func (p *planner) calculateTierPrices(ma20, atr14 float64) (tier1, tier2, tier3 float64)

// calculateStopPrice 计算止损价
func (p *planner) calculateStopPrice(tier3, atr14, ma60 float64) float64

// calculatePosition 仓位计算
func (p *planner) calculatePosition(currentPrice, stopPrice float64) (shares1, shares2, shares3 int)
```

---

## 5. Tracker (跟踪池管理)

### 5.1 接口定义

```go
package tracker

import (
    "context"
    "tradebuddy/internal/model"
)

type Tracker interface {
    // AddPosition 添加到跟踪池
    AddPosition(ctx context.Context, signal *model.TachibanaSignal) error
    
    // UpdateDaily 每日更新跟踪池状态
    UpdateDaily(ctx context.Context, tradeDate string) error
    
    // CheckFilled 检查成交
    CheckFilled(ctx context.Context, pos *model.TrackerPosition, bar *model.DailyBar) (bool, error)
    
    // CheckStopped 检查止损
    CheckStopped(ctx context.Context, pos *model.TrackerPosition, bar *model.DailyBar) (bool, error)
    
    // CheckExpired 检查过期
    CheckExpired(ctx context.Context, pos *model.TrackerPosition) (bool, error)
    
    // GetActiveCount 获取活跃持仓数
    GetActiveCount(ctx context.Context) (int, error)
}

type TrackerConfig struct {
    ExpireDays   int // 40（挂单过期天数）
    MaxPositions int // 120（跟踪池上限）
}
```

### 5.2 状态机

```go
// Position状态转换
const (
    StatusActive  = "active"   // 活跃跟踪
    StatusStopped = "stopped"  // 已止损
    StatusExpired = "expired"  // 已过期
    StatusProfit  = "profit"   // 已止盈
)

// Tier状态
const (
    TierPending = "pending" // 挂单等待
    TierFilled  = "filled"  // 已成交
)
```

---

## 6. Reviewer (复盘分析)

### 6.1 接口定义

```go
package reviewer

import (
    "context"
    "tradebuddy/internal/model"
)

type Reviewer interface {
    // DailyReview 每日复盘
    DailyReview(ctx context.Context, tradeDate string) (*model.Review, error)
    
    // WeeklyReview 周报
    WeeklyReview(ctx context.Context, year int, week int) (*WeeklyReport, error)
    
    // MonthlyReview 月报
    MonthlyReview(ctx context.Context, year int, month int) (*MonthlyReport, error)
}

type WeeklyReport struct {
    Year           int
    Week           int
    WinRate        float64 // 胜率
    AvgRMultiple   float64 // 平均R倍数
    ProfitLossRatio float64 // 盈亏比
    MaxDrawdown    float64 // 最大回撤%
    Notes          string
}

type MonthlyReport struct {
    Year            int
    Month           int
    TotalReturn     float64 // 总收益率%
    SignalAccuracy  float64 // 信号准确率
    ProbCalibError  float64 // 概率校准误差pp
    Notes           string
}
```

---

## 7. Calc (计算工具库)

### 7.1 技术指标

```go
package calc

// MA 简单移动平均
func MA(prices []float64, period int) float64

// EMA 指数移动平均
func EMA(prices []float64, period int) float64

// ATR 真实波幅平均
func ATR(bars []*model.DailyBar, period int) float64

// MACD 指标
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) (dif, dea, macd float64)

// VR 量比
func VR(bars []*model.DailyBar, period int) float64

// WilsonConfidenceInterval Wilson评分置信区间
func WilsonConfidenceInterval(successes, trials int, confidenceLevel float64) (low, high float64)
```

---

## 8. DataSource (数据源适配)

### 8.1 TDX Reader

```go
package tdx

// ReadDayFile 读取单个.day文件
func ReadDayFile(path string) ([]*model.DailyBar, error)

// ReadDayFiles 批量读取
func ReadDayFiles(tdxRoot string, codes []string) (map[string][]*model.DailyBar, error)

// ParseDayRecord 解析32字节记录
func ParseDayRecord(data []byte) (*model.DailyBar, error)
```

### 8.2 Tencent API

```go
package tencent

// GetQFQKLine 获取前复权K线
func GetQFQKLine(code string, period string, n int) ([]*model.DailyBar, error)

// GetWeeklyKLine 获取周K线
func GetWeeklyKLine(code string, n int) ([]*model.WeeklyBar, error)

// GetMonthlyKLine 获取月K线
func GetMonthlyKLine(code string, n int) ([]*model.MonthlyBar, error)
```

---

## 9. CLI命令接口

### 9.1 命令清单

```bash
# 数据导入
tb-cli import tdx --root H:\new_tdx64\vipdoc --codes 600519,000001
tb-cli import instruments --file data/universe.csv

# 手动运行任务
tb-cli run screen --date 2026-10-09
tb-cli run lifecycle --date 2026-10-09
tb-cli run tachibana --date 2026-10-09
tb-cli run review --date 2026-10-09

# 成交录入
tb-cli record buy --code 600519 --price 1850 --qty 100 --note "梯①"
tb-cli record sell --code 600519 --price 1920 --qty 100 --note "止盈"

# 跟踪池管理
tb-cli tracker list
tb-cli tracker update --date 2026-10-09
tb-cli tracker show 600519

# 复盘查看
tb-cli review daily --date 2026-10-09
tb-cli review weekly --week 2026-W41
tb-cli review monthly --month 2026-10

# 数据库管理
tb-cli migrate up
tb-cli migrate down
tb-cli backup --output backups/
tb-cli restore --file backups/tradebuddy_20261001.db
```

---

## 10. 错误处理规范

### 10.1 错误类型

```go
type Error struct {
    Code    string // "DB001", "SCR002"
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}
```

### 10.2 错误码定义

```
DB001 - 数据库连接失败
DB002 - 查询执行失败
DB003 - 记录未找到
DB004 - 唯一约束违反

SCR001 - 数据源不可用
SCR002 - 数据解析失败
SCR003 - 筛选条件无效

LCF001 - 概率表加载失败
LCF002 - 摆动点识别失败
LCF003 - 数据不足（<52周）

TCH001 - 回档带判定失败
TCH002 - 仓位计算溢出
TCH003 - 跟踪池已满

TRK001 - 持仓不存在
TRK002 - 状态转换非法
TRK003 - 重复添加

REV001 - 复盘数据缺失
REV002 - 账本恒等式失败
```

---

## 11. 并发控制

### 11.1 Context使用

```go
// 所有长时间运行的操作都接受context
func (s *Screener) Run(ctx context.Context, tradeDate string) (*ScreenResult, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // 继续执行
    }
    
    // 传递给下游
    results, err := s.store.LoadScreenResults(ctx, tradeDate)
    // ...
}
```

### 11.2 数据库连接池

```go
type Store struct {
    db *sql.DB // 内置连接池
}

// 配置建议
db.SetMaxOpenConns(25)   // 最大连接数
db.SetMaxIdleConns(5)    // 空闲连接数
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 12. 测试接口

### 12.1 Mock接口

```go
// MockStore 用于测试
type MockStore struct {
    mock.Mock
}

func (m *MockStore) GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error) {
    args := m.Called(ctx, code, limit)
    return args.Get(0).([]*model.DailyBar), args.Error(1)
}
```

### 12.2 测试辅助函数

```go
// NewTestStore 创建内存SQLite用于测试
func NewTestStore(t *testing.T) *Store

// LoadGoldenData 加载Golden Test数据
func LoadGoldenData(t *testing.T, date string) *GoldenDataset

// AssertScreenResult 断言初选结果
func AssertScreenResult(t *testing.T, got, want *ScreenResult)
```

---

**下一步：** 创建TRD（技术实现文档）和ADR（架构决策记录）
