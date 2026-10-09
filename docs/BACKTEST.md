# TradeBuddy 回测系统设计

> 历史策略验证 + 每日复盘的完整闭环
> 创建日期：2026-10-09
> 作者：Claude Opus 5.5

---

## 一、回测系统定位

**核心目标：** 验证策略有效性 + 优化参数 + 对比实盘表现

**两种回测模式：**

1. **历史回测（Backtest）** - 验证策略
   - 用途：策略验证、参数优化、风险评估
   - 时间：开发期 + 策略调整时
   - 数据：历史K线数据（可回溯至上市日）

2. **每日复盘（Review）** - 实盘评估
   - 用途：实盘交易事后分析、账本校验
   - 时间：每日15:30（收盘后）
   - 数据：当日成交 + 跟踪池状态

---

## 二、系统架构

```
┌─────────────────────────────────────────────────────┐
│             回测与复盘系统架构                        │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────────┐      ┌──────────────────┐   │
│  │  历史回测引擎     │      │  每日复盘引擎     │   │
│  │  (Backtester)    │      │  (Reviewer)      │   │
│  └──────────────────┘      └──────────────────┘   │
│          │                          │               │
│          ├─ 策略重放                ├─ 账本校验     │
│          ├─ 模拟成交                ├─ 跟踪池更新   │
│          ├─ 绩效统计                ├─ 损益计算     │
│          └─ 报告生成                └─ 报告生成     │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         策略引擎（复用）                      │  │
│  ├──────────────────────────────────────────────┤  │
│  │  Screener → Lifecycle → Tachibana → Tracker  │  │
│  └──────────────────────────────────────────────┘  │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         数据访问层（Store）                   │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 三、历史回测（Backtest）

### 3.1 核心功能

**1. 时间回放引擎**
```go
type Backtester struct {
    store       *store.Store
    screener    *screener.Screener
    lifecycle   *lifecycle.Classifier
    tachibana   *tachibana.Planner
    simulator   *Simulator
    config      *BacktestConfig
}

type BacktestConfig struct {
    StartDate   string  // 回测开始日期 "2023-01-01"
    EndDate     string  // 回测结束日期 "2023-12-31"
    InitCash    float64 // 初始资金 1000000
    Commission  float64 // 手续费率 0.0003
    Slippage    float64 // 滑点率 0.002
    MaxPositions int    // 最大持仓数 120
}
```

**2. 交易模拟器**
```go
type Simulator struct {
    cash            float64
    positions       map[string]*Position
    transactions    []*Transaction
    equity          []EquityPoint
}

// 模拟成交
func (s *Simulator) ExecuteSignal(signal *Signal, bar *DailyBar) *Transaction {
    // 计算滑点
    actualPrice := signal.Price * (1 + s.config.Slippage)
    
    // 计算手续费
    commission := actualPrice * signal.Shares * s.config.Commission
    
    // 更新现金和持仓
    s.cash -= (actualPrice * float64(signal.Shares) + commission)
    s.positions[signal.Code] = &Position{...}
    
    return &Transaction{...}
}
```

**3. 每日策略执行**
```go
func (b *Backtester) Run() (*BacktestResult, error) {
    // 遍历每个交易日
    for date := range tradingDays(b.config.StartDate, b.config.EndDate) {
        // 1. 运行初选
        screenResults := b.screener.Run(ctx, date)
        
        // 2. 运行生命周期分析
        lifecycleStates := b.lifecycle.Analyze(screenResults)
        
        // 3. 生成交易信号
        signals := b.tachibana.GenerateSignals(screenResults, lifecycleStates)
        
        // 4. 模拟成交（按开盘价/收盘价）
        for _, signal := range signals {
            if signal.Decision == "布网" {
                b.simulator.PlaceOrder(signal, date)
            }
        }
        
        // 5. 检查止损/止盈
        b.simulator.CheckExits(date)
        
        // 6. 记录每日权益
        b.recordEquity(date)
    }
    
    return b.generateReport()
}
```

### 3.2 绩效指标

```go
type BacktestResult struct {
    // 收益指标
    TotalReturn      float64  // 总收益率
    AnnualReturn     float64  // 年化收益率
    BenchmarkReturn  float64  // 基准收益率（沪深300）
    Alpha            float64  // 超额收益
    
    // 风险指标
    MaxDrawdown      float64  // 最大回撤
    Volatility       float64  // 波动率
    SharpeRatio      float64  // 夏普比率
    SortinoRatio     float64  // 索提诺比率
    
    // 交易统计
    TotalTrades      int      // 总交易次数
    WinRate          float64  // 胜率
    AvgWin           float64  // 平均盈利
    AvgLoss          float64  // 平均亏损
    ProfitFactor     float64  // 盈亏比
    
    // 持仓统计
    AvgHoldingDays   float64  // 平均持仓天数
    MaxPositions     int      // 最大持仓数
    
    // 分段统计
    MonthlyReturns   []float64
    YearlyReturns    []float64
    
    // 明细
    Transactions     []*Transaction
    EquityCurve      []EquityPoint
}
```

### 3.3 参数优化

```go
// 网格搜索优化参数
type ParamGrid struct {
    MinPctChange    []float64  // [5.0, 6.0, 7.0, 8.0]
    LookbackDays    []int      // [15, 20, 25]
    RiskPerStock    []float64  // [0.005, 0.01, 0.015]
}

func (b *Backtester) OptimizeParams(grid *ParamGrid) (*OptimizationResult, error) {
    bestSharpe := -999.0
    var bestParams *Params
    
    // 遍历参数组合
    for _, pct := range grid.MinPctChange {
        for _, lookback := range grid.LookbackDays {
            for _, risk := range grid.RiskPerStock {
                params := &Params{pct, lookback, risk}
                result := b.RunWithParams(params)
                
                if result.SharpeRatio > bestSharpe {
                    bestSharpe = result.SharpeRatio
                    bestParams = params
                }
            }
        }
    }
    
    return &OptimizationResult{bestParams, bestSharpe}, nil
}
```

---

## 四、每日复盘（Review）

### 4.1 核心功能

**1. 跟踪池状态更新**
```go
func (r *Reviewer) UpdateTracker(tradeDate string) error {
    positions := r.store.ListTrackerPositions(ctx, "active")
    
    for _, pos := range positions {
        bar := r.store.GetBarByDate(ctx, pos.Code, tradeDate)
        
        // 检查成交
        r.checkFilled(pos, bar)
        
        // 检查止损
        if bar.Low <= pos.StopPrice {
            pos.Status = "stopped"
            r.sendAlert("止损触发", pos)
        }
        
        // 检查过期
        if pos.DaysSinceAdded >= 40 {
            pos.Status = "expired"
        }
        
        r.store.UpdateTrackerPosition(ctx, pos)
    }
    
    return nil
}
```

**2. 账本校验**
```go
func (r *Reviewer) ValidateEquity(tradeDate string) error {
    // 恒等式1: 权益 = 现金 + 持仓市值
    cash := r.store.GetCash(tradeDate)
    positionValue := r.calculatePositionValue(tradeDate)
    equity := cash + positionValue
    
    // 恒等式2: 今日权益 = 昨日权益 + 今日损益 - 手续费
    prevEquity := r.store.GetEquity(prevDate)
    dailyPnL := r.calculateDailyPnL(tradeDate)
    commission := r.calculateCommission(tradeDate)
    
    expected := prevEquity + dailyPnL - commission
    
    if math.Abs(equity - expected) > 0.01 {
        return fmt.Errorf("账本不平衡: 实际权益%.2f, 预期%.2f", equity, expected)
    }
    
    return nil
}
```

**3. 损益分析**
```go
type DailyReview struct {
    TradeDate       string
    
    // 成交统计
    FilledCount     int      // 今日成交数
    StoppedCount    int      // 今日止损数
    
    // 损益
    DailyReturn     float64  // 今日收益率
    DailyPnL        float64  // 今日损益金额
    CumReturn       float64  // 累计收益率
    
    // 滑点统计
    AvgSlippage     float64  // 平均滑点
    
    // 持仓
    PositionCount   int      // 当前持仓数
    Cash            float64  // 现金
    Equity          float64  // 总权益
    
    // Top赢家/输家
    TopWinners      []*Position
    TopLosers       []*Position
}
```

---

## 五、数据库Schema扩展

### 5.1 新增表

```sql
-- 回测运行记录
CREATE TABLE backtest_runs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL,
    start_date      TEXT NOT NULL,
    end_date        TEXT NOT NULL,
    init_cash       REAL,
    params_json     TEXT,           -- 参数配置JSON
    total_return    REAL,
    sharpe_ratio    REAL,
    max_drawdown    REAL,
    win_rate        REAL,
    status          TEXT,           -- pending/running/success/failed
    created_at      TEXT DEFAULT CURRENT_TIMESTAMP,
    finished_at     TEXT
);

-- 回测交易明细
CREATE TABLE backtest_transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id          INTEGER NOT NULL,
    trade_date      TEXT NOT NULL,
    code            TEXT NOT NULL,
    direction       TEXT NOT NULL,  -- buy/sell
    price           REAL,
    quantity        INTEGER,
    commission      REAL,
    slippage        REAL,
    reason          TEXT,           -- 梯①/止损/止盈
    FOREIGN KEY (run_id) REFERENCES backtest_runs(id)
);

-- 回测权益曲线
CREATE TABLE backtest_equity (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id          INTEGER NOT NULL,
    trade_date      TEXT NOT NULL,
    cash            REAL,
    position_value  REAL,
    equity          REAL,
    return_pct      REAL,
    drawdown_pct    REAL,
    FOREIGN KEY (run_id) REFERENCES backtest_runs(id)
);
```

---

## 六、CLI命令设计

```bash
# ========================================
# 历史回测命令
# ========================================

# 运行回测
tb-cli backtest run \
    --name "2023年回测" \
    --start 2023-01-01 \
    --end 2023-12-31 \
    --capital 1000000

# 参数优化
tb-cli backtest optimize \
    --start 2022-01-01 \
    --end 2022-12-31 \
    --param min_pct_change=5,6,7,8 \
    --param lookback_days=15,20,25

# 查看回测列表
tb-cli backtest list

# 查看回测详情
tb-cli backtest show --id 1

# 对比多次回测
tb-cli backtest compare --ids 1,2,3

# 导出回测报告
tb-cli backtest export --id 1 --format excel

# ========================================
# 每日复盘命令
# ========================================

# 生成当日复盘
tb-cli review daily --date 2026-10-09

# 生成周报
tb-cli review weekly --week 2026-W41

# 生成月报
tb-cli review monthly --month 2026-10

# 查看历史复盘
tb-cli review list --from 2026-09-01 --to 2026-10-09

# 对比实盘vs回测
tb-cli review compare \
    --live-start 2026-09-01 \
    --backtest-id 1
```

---

## 七、开发计划调整

### 原计划（3期，5周）
```
第一期（2周）：核心引擎 + CLI
第二期（1周）：简化GUI
第三期（2周）：完整GUI + 打包
```

### 新计划（4期，7周）

**第一期（2周）：核心引擎 + CLI**
- Week 1: 数据层 + Screener ✅ Day 1-2已完成
- Week 2: Lifecycle + Tachibana + CLI + **每日复盘**

**第二期（2周）：回测引擎** ⭐ 新增
- Week 3: 回测引擎核心（时间回放、交易模拟）
- Week 4: 绩效统计 + 参数优化 + CLI命令

**第三期（1周）：简化GUI**
- Week 5: Wails项目 + 三视图

**第四期（2周）：完整GUI + 打包**
- Week 6-7: 仪表盘 + 回测报告可视化 + 打包

---

## 八、技术要点

### 8.1 回测 vs 实盘的差异

| 维度 | 历史回测 | 实盘 |
|------|---------|-----|
| 数据 | 历史K线（已知未来） | 实时K线（未知未来） |
| 成交 | 模拟成交（按开盘价） | 真实成交（有滑点） |
| 止损 | 理想止损（按止损价） | 实际止损（可能滑点） |
| 情绪 | 无情绪干扰 | 有心理压力 |
| 费用 | 固定费率 | 真实佣金+印花税 |

### 8.2 前视偏差（Look-Ahead Bias）避免

**错误示例：**
```go
// ❌ 用当日收盘价判断，用开盘价买入（未来信息）
if bar.Close > ma20 {
    buyAtPrice = bar.Open  // 开盘时无法知道收盘价
}
```

**正确示例：**
```go
// ✅ 用昨日收盘价判断，用今日开盘价买入
if prevBar.Close > ma20 {
    buyAtPrice = todayBar.Open  // 时间顺序正确
}
```

### 8.3 幸存者偏差（Survivorship Bias）

- ✅ 回测数据包含退市股票
- ✅ 回测数据包含ST股票
- ✅ 概率表标注样本偏差

### 8.4 性能优化

**批量计算：**
```go
// 预先计算全市场MA/ATR，避免重复计算
cache := make(map[string]*Indicators)
for _, stock := range universe {
    bars := store.GetDailyBars(stock.Code, 252)
    cache[stock.Code] = &Indicators{
        MA20:  calc.MA(bars, 20),
        ATR14: calc.ATR(bars, 14),
    }
}
```

---

## 九、验收标准

### 9.1 回测引擎验收

**功能验收：**
- [ ] 能回测指定日期区间（如2023全年）
- [ ] 绩效指标计算正确（夏普比率、最大回撤等）
- [ ] 参数优化能找到最优组合
- [ ] 权益曲线平滑连续
- [ ] 回测报告导出Excel

**性能验收：**
- [ ] 回测1年数据 < 10分钟（单线程）
- [ ] 回测5年数据 < 1小时
- [ ] 参数优化（10组合）< 2小时

**质量验收：**
- [ ] 无前视偏差（人工审查代码）
- [ ] 无幸存者偏差（数据集检查）
- [ ] 账本恒等式通过（每日权益校验）

### 9.2 每日复盘验收

**功能验收：**
- [ ] 15:30复盘能生成报告
- [ ] 账本恒等式通过
- [ ] 跟踪池状态更新正确
- [ ] 止损/止盈触发准确

**性能验收：**
- [ ] 复盘任务 < 30秒

---

## 十、风险与挑战

### 10.1 回测陷阱

1. **过度拟合** - 参数优化过度，实盘表现差
   - 解决：留出样本外数据验证
   
2. **滑点低估** - 模拟滑点与实际差异大
   - 解决：根据实盘数据动态调整滑点模型

3. **流动性忽略** - 小市值股票无法成交
   - 解决：加入成交量过滤

### 10.2 数据质量

1. **停牌数据** - 停牌期间无法交易
   - 解决：检测停牌，跳过成交

2. **复权数据** - 确保前复权一致性
   - 解决：统一使用前复权，记录复权因子

---

**准备好了吗？** 

现在有两个选择：

1. **继续Day 3-4**（数据源适配）- 按原计划推进
2. **先设计回测引擎细节** - 完善回测模块的技术方案

你想先做哪个？ 🚀
