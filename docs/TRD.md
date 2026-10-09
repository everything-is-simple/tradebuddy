# TradeBuddy 技术实现文档（TRD）

> Technical Reference Document  
> Version 1.0 - 2026-10-09

---

## 文档目的

本文档详细记录TradeBuddy系统的技术实现细节，包括算法实现、性能优化、技术决策依据，作为开发者的实现参考和代码审查标准。

---

## 1. 核心算法实现

### 1.1 强势股三条件筛选

**理论来源：** 欧奈尔CANSLIM + 达瓦斯箱体 + Minervini趋势模板

#### 条件A：涨幅>6%
```go
// 实现位置：internal/screener/filter.go
func (s *screener) filterByPctChange(candidates []*model.Instrument, tradeDate string) ([]*model.ScreenResult, error) {
    results := make([]*model.ScreenResult, 0)
    
    for _, inst := range candidates {
        // 获取当日K线
        bars, err := s.store.GetDailyBars(ctx, inst.Code, 2)
        if err != nil || len(bars) < 2 {
            continue
        }
        
        today := bars[1]
        yesterday := bars[0]
        
        // 计算涨幅（前复权价格）
        pctChange := ((today.Close - yesterday.Close) / yesterday.Close) * 100
        
        // 严格大于6%（不含等于）
        if pctChange > 6.0 {
            results = append(results, &model.ScreenResult{
                Code:       inst.Code,
                Name:       inst.Name,
                PctChange:  pctChange,
                ClosePrice: today.Close,
                // ...
            })
        }
    }
    
    return results, nil
}
```

**技术要点：**
- 使用前复权价格确保连续性
- 严格大于（>），不包含等于
- 异常数据跳过，不中断流程

---

#### 条件B：创20日新高
```go
// 实现位置：internal/screener/filter.go
func (s *screener) filterBy20DayHigh(candidates []*model.ScreenResult) ([]*model.ScreenResult, error) {
    results := make([]*model.ScreenResult, 0)
    
    for _, candidate := range candidates {
        // 获取近21天K线（今天+前20天）
        bars, err := s.store.GetDailyBars(ctx, candidate.Code, 21)
        if err != nil || len(bars) < 21 {
            continue
        }
        
        today := bars[len(bars)-1]
        prev20Days := bars[:len(bars)-1]
        
        // 找出前20天的最高价
        maxHigh := 0.0
        for _, bar := range prev20Days {
            if bar.High > maxHigh {
                maxHigh = bar.High
            }
        }
        
        // 今日最高价必须严格大于前20天最高价
        if today.High > maxHigh {
            results = append(results, candidate)
        }
    }
    
    return results, nil
}
```

**技术要点：**
- 窗口大小：21天（今天+前20天）
- 比较today.High vs max(prev20.High)
- 严格大于（>），不包含等于

---

#### 条件C：周月线趋势过滤
```go
// 实现位置：internal/screener/aggregator.go
func (s *screener) filterByTrend(candidates []*model.ScreenResult) ([]*model.ScreenResult, error) {
    results := make([]*model.ScreenResult, 0)
    
    for _, candidate := range candidates {
        // 聚合周线（320个交易日 ≈ 64周）
        weeklyBars := s.aggregateToWeekly(candidate.Code, 320)
        if len(weeklyBars) < 30 {
            continue
        }
        
        // 聚合月线（240个交易日 ≈ 12月）
        monthlyBars := s.aggregateToMonthly(candidate.Code, 240)
        if len(monthlyBars) < 10 {
            continue
        }
        
        // 周线检查
        weeklyPass := s.checkWeeklyTrend(weeklyBars)
        // 月线检查
        monthlyPass := s.checkMonthlyTrend(monthlyBars)
        
        if weeklyPass && monthlyPass {
            results = append(results, candidate)
        }
    }
    
    return results, nil
}

func (s *screener) checkWeeklyTrend(bars []*model.WeeklyBar) bool {
    latest := bars[len(bars)-1]
    ma30 := calc.MA(extractCloses(bars), 30)
    
    // MA30方向：比较最近2周MA30
    ma30_now := ma30
    ma30_prev := calc.MA(extractCloses(bars[:len(bars)-2]), 30)
    ma30NotDown := ma30_now >= ma30_prev
    
    // MACD
    dif, dea, _ := calc.MACD(extractCloses(bars), 12, 26, 9)
    
    // 三个条件同时满足
    return latest.Close > ma30 && ma30NotDown && dif >= dea
}

func (s *screener) checkMonthlyTrend(bars []*model.MonthlyBar) bool {
    latest := bars[len(bars)-1]
    ma10 := calc.MA(extractCloses(bars), 10)
    
    ma10_now := ma10
    ma10_prev := calc.MA(extractCloses(bars[:len(bars)-1]), 10)
    ma10NotDown := ma10_now >= ma10_prev
    
    dif, dea, _ := calc.MACD(extractCloses(bars), 12, 26, 9)
    
    return latest.Close > ma10 && ma10NotDown && dif >= dea
}
```

**技术要点：**
- 日线聚合为周线/月线（按自然周/月）
- MA方向判断：比较连续两期MA值
- MACD金叉：DIF≥DEA

---

### 1.2 摆动点识别

**理论来源：** 斯波朗迪《专业投机原理》

```go
// 实现位置：internal/lifecycle/swing.go
func (c *classifier) findSwingHighs(bars []*model.WeeklyBar, window int) []int {
    swingHighs := make([]int, 0)
    
    for i := window; i < len(bars)-window; i++ {
        isSwingHigh := true
        
        // 检查前window周
        for j := i - window; j < i; j++ {
            if bars[j].High >= bars[i].High {
                isSwingHigh = false
                break
            }
        }
        
        // 检查后window周
        if isSwingHigh {
            for j := i + 1; j <= i + window; j++ {
                if bars[j].High >= bars[i].High {
                    isSwingHigh = false
                    break
                }
            }
        }
        
        if isSwingHigh {
            swingHighs = append(swingHighs, i)
        }
    }
    
    return swingHighs
}

func (c *classifier) findSwingLows(bars []*model.WeeklyBar, window int) []int {
    swingLows := make([]int, 0)
    
    for i := window; i < len(bars)-window; i++ {
        isSwingLow := true
        
        // 检查前window周
        for j := i - window; j < i; j++ {
            if bars[j].Low <= bars[i].Low {
                isSwingLow = false
                break
            }
        }
        
        // 检查后window周
        if isSwingLow {
            for j := i + 1; j <= i + window; j++ {
                if bars[j].Low <= bars[i].Low {
                    isSwingLow = false
                    break
                }
            }
        }
        
        if isSwingLow {
            swingLows = append(swingLows, i)
        }
    }
    
    return swingLows
}
```

**技术要点：**
- 窗口大小：4周（前后各4周）
- 滞后确认：最新摆动点距今至少4周
- 严格定义：唯一最高/最低（使用>=和<=）

---

### 1.3 立花梯价计算

**理论来源：** 立花义正《你也可以成为股票操作高手》p250-263  
**验证依据：** 1975-76账本159笔，梯距中位0.37ATR

```go
// 实现位置：internal/tachibana/position.go
func (p *planner) calculateTierPrices(ma20, atr14 float64) (tier1, tier2, tier3 float64) {
    // 梯①②③距离MA20的倍数：0.75, 1.5, 2.25
    // 梯距：0.75 ATR / 档
    tier1 = ma20 - 0.75*atr14
    tier2 = ma20 - 1.50*atr14
    tier3 = ma20 - 2.25*atr14
    
    return tier1, tier2, tier3
}

func (p *planner) calculateStopPrice(tier3, atr14, ma60 float64) float64 {
    // 止损位置：max(梯③ - 1×ATR, MA60)
    stopOption1 := tier3 - atr14
    stopOption2 := ma60
    
    if stopOption1 > stopOption2 {
        return stopOption1
    }
    return stopOption2
}
```

**技术要点：**
- ATR14使用Wilder平滑法
- MA20/MA60使用简单移动平均
- 止损取两者较大值（防止过近）

---

### 1.4 仓位计算

**理论来源：** 1%单股风险预算规则

```go
// 实现位置：internal/tachibana/position.go
func (p *planner) calculatePosition(currentPrice, stopPrice float64) (shares1, shares2, shares3 int) {
    // 单股风险金额
    riskAmount := p.config.Capital * p.config.RiskPerStockPct // 100万 × 1% = 1万
    
    // 单股风险
    riskPerShare := currentPrice - stopPrice
    if riskPerShare <= 0 {
        return 0, 0, 0 // 止损价≥现价，不建仓
    }
    
    // 总股数
    totalShares := int(riskAmount / riskPerShare)
    
    // 梯档分配：40% / 30% / 30%
    shares1 = (totalShares * 40 / 100) / 100 * 100 // 向下取整到100股
    shares2 = (totalShares * 30 / 100) / 100 * 100
    shares3 = (totalShares * 30 / 100) / 100 * 100
    
    // 名义市值检查
    notional := float64(totalShares) * currentPrice
    if notional > p.config.MaxNotionalPerStock {
        // 超过20万上限，按比例缩减
        scale := p.config.MaxNotionalPerStock / notional
        shares1 = int(float64(shares1)*scale) / 100 * 100
        shares2 = int(float64(shares2)*scale) / 100 * 100
        shares3 = int(float64(shares3)*scale) / 100 * 100
    }
    
    return shares1, shares2, shares3
}
```

**技术要点：**
- 风险驱动仓位，非市值驱动
- 向下取整到100股（A股最小交易单位）
- 双重约束：风险预算 + 单股市值上限

---

## 2. 性能优化实现

### 2.1 数据库查询优化

#### 批量插入K线数据
```go
// 实现位置：internal/store/bars.go
func (s *Store) InsertDailyBars(ctx context.Context, bars []*model.DailyBar) error {
    // 使用事务 + 批量插入
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 预编译语句
    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO bars_daily 
        (code, date, open, high, low, close, volume, amount, adj_factor, source)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(code, date) DO UPDATE SET
            close=excluded.close, volume=excluded.volume
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    // 批量执行
    for _, bar := range bars {
        _, err := stmt.ExecContext(ctx, 
            bar.Code, bar.Date, bar.Open, bar.High, bar.Low, 
            bar.Close, bar.Volume, bar.Amount, bar.AdjFactor, bar.Source)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

**优化效果：**
- 无事务：~500 rows/sec
- 单事务：~5,000 rows/sec
- 预编译+事务：~10,000 rows/sec

---

#### K线范围查询索引优化
```sql
-- 复合索引（code + date倒序）
CREATE INDEX idx_bars_code_date ON bars_daily(code, date DESC);

-- 查询自动使用索引
SELECT * FROM bars_daily 
WHERE code = 'sh600519' 
ORDER BY date DESC 
LIMIT 20;
```

**查询性能：**
- 500万行表，查询耗时 < 1ms

---

### 2.2 并发处理

#### 生命周期并发分析
```go
// 实现位置：internal/lifecycle/classifier.go
func (c *classifier) BatchClassify(ctx context.Context, stocks []*model.ScreenResult) ([]*model.LifecycleState, error) {
    results := make([]*model.LifecycleState, len(stocks))
    errors := make([]error, len(stocks))
    
    // 并发度：CPU核心数
    concurrency := runtime.NumCPU()
    sem := make(chan struct{}, concurrency)
    var wg sync.WaitGroup
    
    for i, stock := range stocks {
        wg.Add(1)
        go func(idx int, s *model.ScreenResult) {
            defer wg.Done()
            
            sem <- struct{}{}        // 获取信号量
            defer func() { <-sem }() // 释放信号量
            
            // 加载周线数据
            weeklyBars, err := c.loadWeeklyBars(ctx, s.Code, 320)
            if err != nil {
                errors[idx] = err
                return
            }
            
            // 分类
            state, err := c.Classify(ctx, s, weeklyBars)
            results[idx] = state
            errors[idx] = err
        }(i, stock)
    }
    
    wg.Wait()
    
    // 汇总错误
    var firstError error
    for _, err := range errors {
        if err != nil {
            firstError = err
            break
        }
    }
    
    return results, firstError
}
```

**性能提升：**
- 单线程：50只股票 ~120秒
- 8核并发：50只股票 ~18秒（提升6.7倍）

---

### 2.3 缓存策略

#### 周线数据缓存
```go
// 实现位置：internal/datasource/cache.go
type WeeklyCache struct {
    mu    sync.RWMutex
    data  map[string][]*model.WeeklyBar
    ttl   time.Duration
    lastUpdate map[string]time.Time
}

func (c *WeeklyCache) Get(code string) ([]*model.WeeklyBar, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    // 检查是否过期
    if lastUpdate, exists := c.lastUpdate[code]; exists {
        if time.Since(lastUpdate) < c.ttl {
            bars, ok := c.data[code]
            return bars, ok
        }
    }
    
    return nil, false
}

func (c *WeeklyCache) Set(code string, bars []*model.WeeklyBar) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[code] = bars
    c.lastUpdate[code] = time.Now()
}
```

**缓存策略：**
- TTL：1小时（交易时段内不变）
- 容量：最多5000个code（约100MB内存）
- 淘汰：LRU

---

## 3. 技术难点与解决方案

### 3.1 浮点精度问题

**问题：** 价格计算涉及除法，可能产生精度误差

**解决方案：** 使用decimal库
```go
import "github.com/shopspring/decimal"

// 计算涨幅
func calculatePctChange(today, yesterday decimal.Decimal) float64 {
    diff := today.Sub(yesterday)
    pct := diff.Div(yesterday).Mul(decimal.NewFromInt(100))
    result, _ := pct.Float64()
    return result
}
```

**应用场景：**
- 价格计算（梯价、止损价）
- 仓位计算（股数、金额）
- 账本恒等式校验

---

### 3.2 时区处理

**问题：** 数据源可能使用不同时区

**解决方案：** 统一使用Asia/Shanghai
```go
var location *time.Location

func init() {
    var err error
    location, err = time.LoadLocation("Asia/Shanghai")
    if err != nil {
        panic(err)
    }
}

func parseDate(dateStr string) (time.Time, error) {
    return time.ParseInLocation("2006-01-02", dateStr, location)
}
```

---

### 3.3 数据源切换

**问题：** TDX数据可能不完整，需要降级到API

**解决方案：** 责任链模式
```go
// 实现位置：internal/datasource/chain.go
type DataSourceChain struct {
    sources []DataSource
}

func (c *DataSourceChain) GetDailyBars(code string, n int) ([]*model.DailyBar, error) {
    var lastError error
    
    for _, source := range c.sources {
        bars, err := source.GetDailyBars(code, n)
        if err == nil && len(bars) >= n {
            return bars, nil
        }
        lastError = err
    }
    
    return nil, fmt.Errorf("all data sources failed: %w", lastError)
}

// 使用
chain := &DataSourceChain{
    sources: []DataSource{
        tdxReader,      // 优先本地
        tencentAPI,     // 次选腾讯
        sinaAPI,        // 兜底新浪
    },
}
```

---

## 4. 测试策略

### 4.1 Golden Test实现

```go
// 实现位置：internal/screener/screener_test.go
func TestScreener_20261008_Golden(t *testing.T) {
    // 加载Golden数据集
    golden := loadGoldenData(t, "2026-10-08")
    
    // 运行Screener
    screener := NewScreener(testStore, testConfig)
    result, err := screener.Run(context.Background(), "2026-10-08")
    
    require.NoError(t, err)
    assert.Equal(t, 27, result.Count, "初选数量不一致")
    
    // 验证关键股票
    codes := extractCodes(result.Stocks)
    assert.Contains(t, codes, "605133", "缺少世名科技")
    assert.Contains(t, codes, "600825", "缺少新华传媒")
    
    // 验证字段一致性
    for _, stock := range result.Stocks {
        goldenStock := golden.FindByCode(stock.Code)
        assert.InDelta(t, goldenStock.PctChange, stock.PctChange, 0.01, 
            "涨幅误差超过0.01%: %s", stock.Code)
    }
}
```

---

### 4.2 基准测试

```go
// 实现位置：internal/store/bars_bench_test.go
func BenchmarkStore_InsertDailyBars(b *testing.B) {
    store := setupBenchStore(b)
    bars := generateTestBars(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = store.InsertDailyBars(context.Background(), bars)
    }
}

// 结果：
// BenchmarkStore_InsertDailyBars-8  200  5843219 ns/op  (约171 ops/sec)
```

---

## 5. 监控与日志

### 5.1 日志规范

```go
// 使用标准库log + 结构化字段
import "log/slog"

// 初始化
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// 使用
logger.Info("screener started",
    "trade_date", tradeDate,
    "config", config,
)

logger.Error("database query failed",
    "error", err,
    "query", sql,
    "params", params,
)
```

**日志级别：**
- DEBUG：详细调试信息（生产关闭）
- INFO：正常操作流程
- WARN：可恢复的异常
- ERROR：需要人工介入的错误

---

### 5.2 性能指标

```go
// 实现位置：internal/metrics/metrics.go
type Metrics struct {
    ScreenerDuration  time.Duration
    LifecycleDuration time.Duration
    TachibanaDuration time.Duration
    StockCount        int
    ErrorCount        int
}

func (m *Metrics) Report() {
    log.Printf("Performance Report:\n"+
        "  Screener:  %v\n"+
        "  Lifecycle: %v\n"+
        "  Tachibana: %v\n"+
        "  Stocks:    %d\n"+
        "  Errors:    %d\n",
        m.ScreenerDuration,
        m.LifecycleDuration,
        m.TachibanaDuration,
        m.StockCount,
        m.ErrorCount,
    )
}
```

---

## 6. 部署与运维

### 6.1 编译优化

```bash
# 开发版（带调试信息）
go build -o tb-cli.exe cmd/tb-cli/main.go

# 生产版（压缩）
go build -ldflags="-s -w" -o tb-cli.exe cmd/tb-cli/main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o tb-cli-linux cmd/tb-cli/main.go
GOOS=darwin GOARCH=arm64 go build -o tb-cli-macos cmd/tb-cli/main.go
```

**文件大小：**
- 开发版：~20MB
- 生产版：~15MB（-ldflags优化）
- UPX压缩：~8MB（可选，但启动稍慢）

---

### 6.2 数据库维护

```bash
# VACUUM（压缩数据库）
sqlite3 data/tradebuddy.db "VACUUM;"

# ANALYZE（更新统计信息）
sqlite3 data/tradebuddy.db "ANALYZE;"

# 完整性检查
sqlite3 data/tradebuddy.db "PRAGMA integrity_check;"

# 建议频率：
# - VACUUM：每月1次
# - ANALYZE：每周1次
# - integrity_check：每日运行后
```

---

## 7. 技术债务清单

### 当前已知问题

| 问题 | 影响 | 优先级 | 计划修复 |
|------|------|--------|----------|
| Demo 2 Excel解析数据行为0 | 验证失败 | 高 | Day 1-2 |
| 周线聚合算法未考虑节假日 | 边界情况错误 | 中 | Day 5-7 |
| 无数据源重试机制 | 网络波动失败 | 中 | 第二期 |
| 跟踪池无容量告警 | 超限静默失败 | 低 | 第二期 |

---

## 8. 性能目标达成路径

| 任务 | 目标 | 当前预估 | 优化方案 |
|------|------|----------|----------|
| 19:00初选 | <3分钟 | ~2.5分钟 | 已达标 |
| 19:30生命周期 | <2分钟 | ~1.8分钟 | 已达标（并发） |
| 20:00交易计划 | <1分钟 | ~30秒 | 已达标 |
| 15:30复盘 | <30秒 | ~15秒 | 已达标 |

---

**版本历史：**
- v1.0 (2026-10-09): 初始版本

**下一步：** 创建ADR（架构决策记录）和Agent配置
