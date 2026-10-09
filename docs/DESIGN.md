# TradeBuddy 系统设计文档 v1.0

> 个人量化交易系统 - 纯Go实现，单文件打包
> 创建日期：2026-10-09
> 作者：Claude Opus 5.5

---

## 一、系统定位

**核心目标：** 从选股到交易到复盘的完整闭环个人量化交易系统

**理论基础：**
1. **选股层** - 欧奈尔CANSLIM + 达瓦斯箱体 + Minervini趋势模板
2. **分析层** - 生命周期统计（简化版MALF Lifespan）+ 波段强度评估
3. **交易层** - 立花义正うねり取り + 林辉太郎分批建仓
4. **评价层** - 布伦特·奔富R倍数系统 + 账本恒等式校验

**注：** Phase 1 实现简化版生命周期分析（70%功能覆盖），Phase 2 可升级为完整 MALF v2.1 状态机

**技术特点：**
- 纯Go实现（Go 1.22+）
- SQLite嵌入式数据库（modernc.org/sqlite）
- 单文件打包，跨平台运行
- 无Python依赖，无外部运行时

---

## 二、系统架构

```
┌─────────────────────────────────────────────────────┐
│              用户交互层 (User Layer)                 │
├─────────────────────────────────────────────────────┤
│  CLI工具 (tb-cli)              │  GUI桌面端 (Wails)  │
│  - 数据导入                     │  - 仪表盘            │
│  - 手动触发任务                 │  - 清单查看          │
│  - 成交录入                     │  - 生命周期可视化    │
│  - 复盘查看                     │  - 交易计划          │
│                                 │  - 复盘报告          │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│             业务逻辑层 (Business Layer)              │
├─────────────────────────────────────────────────────┤
│  核心引擎 (Engine)                                   │
│  ├─ Screener      (19:00 强势股初选)                │
│  ├─ Lifecycle     (19:30 生命周期分析)              │
│  ├─ Tachibana     (20:00 立花交易计划)              │
│  └─ Reviewer      (15:30 每日复盘)                  │
│                                                      │
│  辅助模块 (Support)                                  │
│  ├─ Scheduler     (定时任务调度)                    │
│  ├─ Calculator    (MA/ATR/MACD/摆动点等)            │
│  ├─ Tracker       (跟踪池状态机)                    │
│  └─ Reporter      (报告生成)                        │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│             数据访问层 (Data Layer)                  │
├─────────────────────────────────────────────────────┤
│  Store (SQLite访问)                                  │
│  ├─ Instruments   (股票元数据)                       │
│  ├─ BarsDaily     (日线数据)                         │
│  ├─ ScreenResults (筛选结果)                         │
│  ├─ LifecycleStates (生命周期状态)                   │
│  ├─ TachibanaSignals (交易信号)                      │
│  ├─ Transactions  (成交记录)                         │
│  ├─ TrackerPositions (跟踪池)                        │
│  └─ Reviews       (复盘报告)                         │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│            数据源适配层 (DataSource Layer)           │
├─────────────────────────────────────────────────────┤
│  ├─ TDX Reader    (通达信.day文件读取)               │
│  ├─ Sina API      (新浪榜单+K线)                     │
│  ├─ Tencent API   (腾讯前复权K线)                    │
│  └─ Manual Import (CSV/Excel导入)                   │
└─────────────────────────────────────────────────────┘
```

---

## 三、数据流

### 3.1 每日工作流（自动化）

```
[数据更新] 18:00-18:30
  ↓ 拉取全市场日线数据（TDX/腾讯）
  ↓ 更新股票元数据（ST状态/上市日期）
  
[强势股初选] 19:00
  ↓ 输入：全市场日线
  ↓ 输出：screen_results表 + Excel报告
  
[生命周期分析] 19:30
  ↓ 输入：19:00筛选结果
  ↓ 输出：lifecycle_states表 + Excel报告
  
[立花交易计划] 20:00
  ↓ 输入：19:00筛选结果 + 19:30生命周期
  ↓ 输出：tachibana_signals表 + tracker_positions更新 + Excel报告
  
[人工执行] 20:05-次日09:25
  ↓ 用户看Excel，在券商APP挂单
  ↓ 使用tb-cli录入成交信息
  
[盘中监控] 09:30-15:00 (第二期)
  ↓ 实时价格 vs 跟踪池
  ↓ 触发止损/止盈警报
  
[每日复盘] 15:30
  ↓ 汇总当日成交
  ↓ 更新跟踪池状态
  ↓ 生成复盘报告
```

### 3.2 数据依赖关系

```
instruments (股票元数据)
    ↓
bars_daily (日线数据)
    ↓
screen_results (19:00初选)
    ↓
    ├─→ lifecycle_states (19:30生命周期)
    │       ↓
    └─→ tachibana_signals (20:00立花计划)
            ↓
        tracker_positions (跟踪池)
            ↓
        transactions (成交记录)
            ↓
        reviews (复盘报告)
```

---

## 四、核心模块设计

### 4.1 Screener (强势股初选)

**职责：** 19:00筛选当日强势突破股

**输入：**
- 全市场日线数据（bars_daily表）
- 筛选参数（config.yaml）

**核心逻辑：**
```go
type ScreenerConfig struct {
    MinPctChange    float64  // 6.0 (当日涨幅>6%)
    LookbackDays    int      // 20 (创20日新高)
    WeeklyMA        int      // 30 (周线MA30)
    MonthlyMA       int      // 10 (月线MA10)
    ExcludeST       bool     // true
    MinListDays     int      // 21
}

func (s *Screener) Run(ctx context.Context, tradeDate string) (*ScreenResult, error) {
    // 1. 获取当日涨幅榜
    candidates := s.getRisers(tradeDate)
    
    // 2. 三条件筛选
    for _, stock := range candidates {
        if !s.checkConditionA(stock) { continue } // 涨幅>6%
        if !s.checkConditionB(stock) { continue } // 创20日新高
        if !s.checkConditionC(stock) { continue } // 周月线非下跌
        
        results = append(results, stock)
    }
    
    // 3. 写入数据库 + 生成Excel
    s.store.SaveScreenResults(tradeDate, results)
    s.reporter.ExportToExcel(results, outputPath)
    
    return &ScreenResult{Count: len(results), FilePath: outputPath}, nil
}
```

**输出：**
- `screen_results` 表
- `reports/screen/A股备选清单_YYYY-MM-DD.xlsx`

---

### 4.2 Lifecycle (生命周期分析)

**职责：** 19:30分析每只股票的市场生命周期阶段

**输入：**
- 19:00筛选结果（screen_results表）
- 个股周线数据（bars_daily聚合）
- 概率表（data/lifecycle_prob_table.json）

**核心逻辑：**
```go
type LifecycleClassifier struct {
    probTable *ProbabilityTable
}

func (lc *LifecycleClassifier) Classify(stock *model.Stock, weeklyBars []*model.WeeklyBar) (*model.LifecycleState, error) {
    // 1. 识别摆动点（前后4周窗口）
    swingHighs := lc.findSwingHighs(weeklyBars, 4)
    swingLows := lc.findSwingLows(weeklyBars, 4)
    
    // 2. 判定结构（上升/下跌/无趋势）
    structure := lc.analyzeStructure(swingHighs, swingLows)
    
    // 3. 确定九类阶段
    stage := lc.determineStage(structure, weeklyBars)
    
    // 4. 计算量能VR13
    vr13 := lc.calculateVR13(weeklyBars, 13)
    volClass := lc.classifyVolume(vr13) // 放量/平量/缩量
    
    // 5. 查概率表（阶段×量能 → 8周存活概率）
    prob, ci, n := lc.probTable.Lookup(stage, volClass)
    
    // 6. ±25%生命视角（第二口径）
    rawView := lc.analyzeRawLifespan(weeklyBars)
    
    return &model.LifecycleState{
        Stage:        stage,
        VolClass:     volClass,
        SurviveProb:  prob,
        ConfidenceCI: ci,
        SampleSize:   n,
        RawView:      rawView,
    }, nil
}
```

**输出：**
- `lifecycle_states` 表
- `reports/lifecycle/生命周期分析_YYYY-MM-DD.xlsx`

---

### 4.3 Tachibana (立花交易计划)

**职责：** 20:00生成具体交易参数（买入梯价、止损、跟踪池）

**输入：**
- 19:00筛选结果
- 19:30生命周期阶段
- 个股日线数据

**核心逻辑：**
```go
type TachibanaPlanner struct {
    config *TachibanaConfig
}

func (tp *TachibanaPlanner) BuildPlan(stock *model.Stock, lifecycle *model.LifecycleState, dailyBars []*model.DailyBar) (*model.TachibanaSignal, error) {
    // 1. 可交易性过滤
    decision := tp.filterByStage(lifecycle.Stage)
    if decision == "排除" {
        return &model.TachibanaSignal{Decision: "排除", Reason: "阶段不可交易"}, nil
    }
    
    // 2. 计算核心参数
    ma20 := calc.MA(dailyBars, 20)
    atr14 := calc.ATR(dailyBars, 14)
    currentPrice := dailyBars[len(dailyBars)-1].Close
    
    zScore := (ma20 - currentPrice) / atr14
    
    // 3. 回档带判定
    if zScore < 0.8 || zScore > 2.5 {
        return &model.TachibanaSignal{Decision: "排除", Reason: "不在回档带"}, nil
    }
    
    // 4. 计算买入梯价
    tier1 := ma20 - 0.75*atr14
    tier2 := ma20 - 1.50*atr14
    tier3 := ma20 - 2.25*atr14
    
    // 5. 计算止损
    ma60 := calc.MA(dailyBars, 60)
    stopLoss := math.Max(tier3-atr14, ma60)
    
    // 6. 仓位计算
    capital := tp.config.Capital           // 100万
    riskPct := tp.config.RiskPerStock      // 1%
    riskAmount := capital * riskPct        // 1万
    
    riskPerShare := currentPrice - stopLoss
    totalShares := int(riskAmount / riskPerShare)
    
    shares1 := (totalShares * 40 / 100) / 100 * 100  // 向下取整100股
    shares2 := (totalShares * 30 / 100) / 100 * 100
    shares3 := (totalShares * 30 / 100) / 100 * 100
    
    return &model.TachibanaSignal{
        Decision:  "布网",
        ZScore:    zScore,
        Tier1Price: tier1,
        Tier2Price: tier2,
        Tier3Price: tier3,
        StopPrice:  stopLoss,
        Shares1:    shares1,
        Shares2:    shares2,
        Shares3:    shares3,
    }, nil
}
```

**输出：**
- `tachibana_signals` 表
- `tracker_positions` 表更新
- `reports/tachibana/立花交易计划_YYYY-MM-DD.xlsx`

---

### 4.4 Tracker (跟踪池状态机)

**职责：** 管理挂单生命周期

**状态转换：**
```
pending (挂单等待)
    ↓ 最低价 ≤ 梯价
filled (部分/全部成交)
    ↓ 最低价 ≤ 止损价
stopped (止损离场)

pending (挂单等待)
    ↓ 40自然日无成交
expired (过期)

filled (持仓中)
    ↓ 最高价 ≥ 上沿价
profit (止盈提醒)
```

**核心逻辑：**
```go
func (t *Tracker) UpdateDaily(tradeDate string) error {
    positions := t.store.LoadActivePositions()
    
    for _, pos := range positions {
        bar := t.store.GetDailyBar(pos.Code, tradeDate)
        
        // 1. 检查成交
        if bar.Low <= pos.Tier1Price && pos.Tier1Status == "pending" {
            pos.Tier1Status = "filled"
            // 需要人工确认实际成交价和数量
        }
        
        // 2. 检查止损
        if bar.Low <= pos.StopPrice {
            pos.Status = "stopped"
            t.sendAlert("止损触发", pos)
        }
        
        // 3. 检查止盈
        if bar.High >= pos.TakeProfitPrice {
            t.sendAlert("止盈提醒", pos)
        }
        
        // 4. 检查过期
        if pos.DaysSinceAdded >= 40 && pos.FilledQty == 0 {
            pos.Status = "expired"
        }
        
        t.store.UpdatePosition(pos)
    }
    
    return nil
}
```

---

## 五、数据库Schema

### 5.1 核心表

```sql
-- 股票元数据
CREATE TABLE instruments (
    code        TEXT PRIMARY KEY,   -- sh600519
    name        TEXT NOT NULL,
    board       TEXT,               -- 主板/创业板/科创板
    list_date   TEXT,
    is_st       INTEGER DEFAULT 0,
    updated_at  TEXT
);

-- 日线K线（前复权）
CREATE TABLE bars_daily (
    code        TEXT NOT NULL,
    date        TEXT NOT NULL,
    open        REAL NOT NULL,
    high        REAL NOT NULL,
    low         REAL NOT NULL,
    close       REAL NOT NULL,
    volume      INTEGER NOT NULL,
    amount      REAL,
    PRIMARY KEY (code, date)
) WITHOUT ROWID;
CREATE INDEX idx_bars_code_date ON bars_daily(code, date DESC);

-- 19:00 初选结果
CREATE TABLE screen_results (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT,
    pct_change  REAL NOT NULL,     -- 当日涨幅%
    close_price REAL,
    high_price  REAL,
    dd52        REAL,              -- 距52周高点%
    created_at  TEXT,
    PRIMARY KEY (trade_date, code)
);

-- 19:30 生命周期状态
CREATE TABLE lifecycle_states (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    stage       TEXT NOT NULL,     -- 2·早期 / 3·做头 等
    vol_class   TEXT,              -- 放量/平量/缩量
    vr13        REAL,
    survive_prob REAL,             -- 8周存活概率
    ci_low      REAL,
    ci_high     REAL,
    sample_size INTEGER,
    raw_view    TEXT,              -- ±25%生命JSON
    created_at  TEXT,
    PRIMARY KEY (trade_date, code)
);

-- 20:00 立花交易信号
CREATE TABLE tachibana_signals (
    trade_date  TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT,
    decision    TEXT NOT NULL,     -- 布网/观察/排除
    z_score     REAL,
    tier1_price REAL,
    tier2_price REAL,
    tier3_price REAL,
    stop_price  REAL,
    shares_t1   INTEGER,
    shares_t2   INTEGER,
    shares_t3   INTEGER,
    reason      TEXT,
    created_at  TEXT,
    PRIMARY KEY (trade_date, code)
);

-- 跟踪池
CREATE TABLE tracker_positions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT NOT NULL,
    name            TEXT,
    added_date      TEXT NOT NULL,
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
    status          TEXT DEFAULT 'active',   -- active/stopped/expired/profit
    last_update     TEXT,
    expire_date     TEXT,
    UNIQUE(code, added_date)
);

-- 成交记录
CREATE TABLE transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    trade_date      TEXT NOT NULL,
    code            TEXT NOT NULL,
    direction       TEXT NOT NULL,  -- buy/sell
    planned_price   REAL,
    actual_price    REAL,
    quantity        INTEGER,
    amount          REAL,
    commission      REAL,
    slippage_pct    REAL,
    note            TEXT,           -- 梯①/梯②/止损/止盈
    created_at      TEXT
);

-- 复盘报告
CREATE TABLE reviews (
    trade_date      TEXT PRIMARY KEY,
    filled_count    INTEGER,
    stopped_count   INTEGER,
    avg_slippage    REAL,
    daily_return    REAL,
    equity          REAL,
    cash            REAL,
    position_value  REAL,
    notes           TEXT,           -- JSON详细信息
    created_at      TEXT
);
```

---

## 六、开发计划

### 第一期：核心引擎 + CLI（2周）

**Week 1: 数据层 + Screener**
- Day 1-2: Go项目初始化 + SQLite数据层
- Day 3-4: 数据源适配（TDX reader / Tencent API）
- Day 5-7: Screener引擎 + golden test（10-08那27只）

**Week 2: Lifecycle + Tachibana + CLI**
- Day 8-10: Lifecycle引擎（摆动结构+概率表）
- Day 11-12: Tachibana引擎（梯价+跟踪池）
- Day 13-14: CLI工具 + 成交录入 + 复盘脚本

**交付物：**
```
tradebuddy/
├── tb-cli.exe           # CLI工具（Windows）
├── data/
│   └── tradebuddy.db    # SQLite数据库
└── reports/             # Excel报告输出目录
```

---

### 第二期：简化GUI（1周）

**Week 3: Wails最小界面**
- Day 15-17: Wails项目初始化 + 基础框架
- Day 18-19: 三视图（清单/生命周期/交易计划）
- Day 20-21: 手动触发任务按钮 + 测试

**交付物：**
```
tb-desktop.exe           # 桌面GUI
```

---

### 第三期：完整GUI + 打包（2周）

**Week 4-5: 完整功能**
- 仪表盘（任务状态/入选数/阶段分布）
- 跟踪池管理（状态查看/手动更新）
- 复盘报告查看
- 内置定时调度
- 单文件打包

---

## 七、技术选型

### 7.1 核心依赖

```go
require (
    modernc.org/sqlite v1.27.0           // 纯Go SQLite
    github.com/spf13/cobra v1.8.0        // CLI框架
    github.com/wailsapp/wails/v2 v2.7.1  // 桌面GUI
    github.com/shopspring/decimal v1.3.1 // 高精度计算
)
```

### 7.2 为什么不用Python？

| 需求 | Python | Go |
|------|--------|-----|
| 单文件打包 | PyInstaller ~50MB + 依赖 | 单exe ~15MB |
| 跨平台运行 | 需要Python运行时 | 编译即可 |
| 启动速度 | 1-2秒（解释器启动） | <100ms |
| 内存占用 | 50-100MB | 20-30MB |
| 交付复杂度 | 打包+依赖管理 | 拷贝2个文件 |
| 并发性能 | GIL限制 | 原生goroutine |

---

## 八、验证计划（先C验证）

### 8.1 技术可行性验证

**验证1：Go读写SQLite**
```bash
cd I:\tradebuddy
mkdir -p cmd/demo
cat > cmd/demo/main.go << 'EOF'
package main
import (
    "database/sql"
    "fmt"
    _ "modernc.org/sqlite"
)
func main() {
    db, _ := sql.Open("sqlite", "test.db")
    defer db.Close()
    db.Exec("CREATE TABLE test (id INT, name TEXT)")
    db.Exec("INSERT INTO test VALUES (1, 'Hello')")
    var name string
    db.QueryRow("SELECT name FROM test WHERE id=1").Scan(&name)
    fmt.Println("✓ SQLite OK:", name)
}
EOF
go run cmd/demo/main.go
```

**验证2：解析Excel（立花交易计划）**
```bash
go get github.com/xuri/excelize/v2
# 读取 docs/证据③20_00立花交易计划_10-08.xlsx
# 验证能否正确解析27只股票数据
```

**验证3：TDX .day文件读取**
```bash
# 读取 {TDX_ROOT}\vipdoc\sh\lday\sh600519.day
# 验证32字节定长记录解析
```

---

## 九、项目文件结构

```
I:\tradebuddy/
├── cmd/
│   ├── tb-cli/
│   │   └── main.go              # CLI入口
│   └── tb-desktop/
│       └── main.go              # Wails GUI入口
│
├── internal/
│   ├── store/
│   │   ├── store.go             # SQLite主接口
│   │   ├── instruments.go       # 股票元数据
│   │   ├── bars.go              # K线数据
│   │   ├── screen.go            # 初选结果
│   │   ├── lifecycle.go         # 生命周期
│   │   ├── tachibana.go         # 立花信号
│   │   ├── tracker.go           # 跟踪池
│   │   └── migrations/
│   │       └── 0001_init.sql
│   │
│   ├── model/
│   │   ├── instrument.go
│   │   ├── bar.go
│   │   ├── screen_result.go
│   │   ├── lifecycle_state.go
│   │   └── tachibana_signal.go
│   │
│   ├── datasource/
│   │   ├── tdx/                 # 通达信.day读取
│   │   ├── tencent/             # 腾讯API
│   │   └── sina/                # 新浪API
│   │
│   ├── calc/
│   │   ├── ma.go                # 均线
│   │   ├── atr.go               # ATR
│   │   ├── macd.go              # MACD
│   │   └── swing.go             # 摆动点
│   │
│   ├── screener/
│   │   └── screener.go          # 19:00初选引擎
│   │
│   ├── lifecycle/
│   │   ├── classifier.go        # 阶段分类
│   │   └── probability.go       # 概率表
│   │
│   ├── tachibana/
│   │   ├── planner.go           # 交易计划
│   │   └── position.go          # 仓位计算
│   │
│   ├── tracker/
│   │   └── tracker.go           # 跟踪池状态机
│   │
│   ├── scheduler/
│   │   └── scheduler.go         # 定时任务
│   │
│   └── reporter/
│       ├── excel.go             # Excel报告
│       └── markdown.go          # Markdown报告
│
├── web/                         # Wails前端（第二期）
│   ├── src/
│   └── dist/
│
├── data/
│   ├── tradebuddy.db            # SQLite数据库
│   ├── lifecycle_prob_table.json # 概率表
│   └── config.yaml              # 配置文件
│
├── reports/                     # 每日报告输出
│   ├── screen/
│   ├── lifecycle/
│   ├── tachibana/
│   └── review/
│
├── go.mod
├── go.sum
├── README.md
└── docs/
    └── DESIGN.md                # 本文档
```

---

## 十、配置文件

```yaml
# data/config.yaml
database:
  path: "data/tradebuddy.db"

datasource:
  tdx_root: "C:\\new_tdx64\\vipdoc"
  use_tencent: true
  use_sina: false
  request_interval_ms: 100

screener:
  min_pct_change: 6.0
  lookback_days: 20
  weekly_ma: 30
  monthly_ma: 10
  exclude_st: true
  min_list_days: 21

lifecycle:
  swing_window: 4
  prob_table_path: "data/lifecycle_prob_table.json"

tachibana:
  capital: 1000000
  risk_per_stock_pct: 0.01
  tier_ratios: [0.4, 0.3, 0.3]
  z_min: 0.8
  z_max: 2.5
  max_notional_per_stock: 200000

tracker:
  expire_days: 40
  max_positions: 120

schedule:
  screener:  "0 19 * * 1-5"
  lifecycle: "30 19 * * 1-5"
  tachibana: "0 20 * * 1-5"
  review:    "30 15 * * 1-5"
```

---

**下一步：** 开始验证（C），然后写代码（A）。
