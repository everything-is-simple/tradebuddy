# TradeBuddy 开发任务清单

> Development Task List  
> Version 1.0 - 2026-10-09

---

## 📋 任务概览

**开发模式：** 三期迭代，每期独立交付  
**当前阶段：** 第一期（核心引擎 + CLI，2周）  
**开始日期：** 2026-10-09  
**预计完成：** 2026-10-23

---

## 🎯 第一期：核心引擎 + CLI（2周）

**目标：** 完成选股→分析→交易计划→复盘的自动化闭环

### Week 1: 数据层 + Screener（Day 1-7）

#### ✅ Day 0: 项目准备（已完成）

- [x] 技术验证（Demo 1-3）
- [x] 归档GLM文档
- [x] 创建项目结构
- [x] 编写核心文档（README/DESIGN/REQUIRE）
- [x] 数据库Schema设计（0001_init.sql）
- [x] 数据模型定义（model/*.go）

---

#### 📌 Day 1-2: 数据访问层（store）

**任务：** 实现SQLite数据访问层

**文件清单：**
```
internal/store/
├── store.go              # 主接口
├── instrument.go         # 股票元数据操作
├── bars.go               # K线数据操作
├── screen.go             # 初选结果操作
├── lifecycle.go          # 生命周期操作
├── tachibana.go          # 立花信号操作
├── tracker.go            # 跟踪池操作
├── transaction.go        # 成交记录操作
├── review.go             # 复盘报告操作
└── migrate.go            # 数据库迁移
```

**核心接口：**
```go
type Store interface {
    // 迁移
    Migrate() error
    
    // 股票元数据
    UpsertInstrument(ctx context.Context, inst *model.Instrument) error
    ListInstruments(ctx context.Context, filter InstrumentFilter) ([]*model.Instrument, error)
    
    // K线数据
    InsertDailyBars(ctx context.Context, bars []*model.DailyBar) error
    GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error)
    GetDailyBarsRange(ctx context.Context, code string, start, end string) ([]*model.DailyBar, error)
    
    // 筛选结果
    SaveScreenResults(ctx context.Context, tradeDate string, results []*model.ScreenResult) error
    LoadScreenResults(ctx context.Context, tradeDate string) ([]*model.ScreenResult, error)
    
    // 其他CRUD操作...
}
```

**验收标准：**
- [ ] 所有表创建成功
- [ ] 基础CRUD操作通过单元测试
- [ ] 迁移可正常执行和回滚

---

#### 📌 Day 3-4: 数据源适配（datasource）

**任务：** 实现TDX文件读取 + 腾讯API适配

**文件清单：**
```
internal/datasource/
├── tdx/
│   ├── reader.go         # 读取.day文件
│   └── reader_test.go
├── tencent/
│   ├── api.go            # 腾讯K线API
│   └── api_test.go
└── sina/
    ├── api.go            # 新浪API（备选）
    └── api_test.go
```

**TDX Reader核心逻辑：**
```go
// 读取单个.day文件
func ReadDayFile(path string) ([]*model.DailyBar, error)

// 批量读取指定代码列表
func ReadDayFiles(tdxRoot string, codes []string) (map[string][]*model.DailyBar, error)

// 32字节定长记录解析
const DayRecordSize = 32
type dayRecord struct {
    date   uint32  // YYYYMMDD
    open   uint32  // 价格×100
    high   uint32
    low    uint32
    close  uint32
    amount float32
    volume uint32
    count  uint32
}
```

**腾讯API核心逻辑：**
```go
// 获取前复权K线
func GetQFQKLine(code string, period string, n int) ([]*model.DailyBar, error)
// period: day/week/month

// 限流控制
type RateLimiter struct {
    interval time.Duration
    lastReq  time.Time
}
```

**验收标准：**
- [ ] 能正确读取 sh600519.day（茅台数据）
- [ ] 能从腾讯API获取前复权日K
- [ ] 限流策略有效（0.08-0.1s间隔）
- [ ] 数据格式与Demo 3一致

---

#### 📌 Day 5-7: Screener引擎（screener）

**任务：** 实现19:00强势股初选引擎

**文件清单：**
```
internal/screener/
├── screener.go           # 主引擎
├── filter.go             # 三条件过滤器
├── aggregator.go         # 周线/月线聚合
└── screener_test.go
```

**核心逻辑：**
```go
type Screener struct {
    store  store.Store
    config *ScreenerConfig
}

func (s *Screener) Run(ctx context.Context, tradeDate string) (*ScreenResult, error) {
    // 1. 获取全市场日线数据
    instruments := s.store.ListInstruments(...)
    
    // 2. 筛选候选股（涨幅>6%）
    candidates := s.filterByPctChange(instruments, tradeDate, 6.0)
    
    // 3. 条件B：创20日新高
    candidates = s.filterBy20DayHigh(candidates, tradeDate)
    
    // 4. 条件C：周月线趋势过滤
    candidates = s.filterByTrend(candidates, tradeDate)
    
    // 5. 保存结果
    s.store.SaveScreenResults(tradeDate, candidates)
    
    // 6. 生成Excel报告
    s.exportToExcel(candidates, tradeDate)
    
    return &ScreenResult{Count: len(candidates), FilePath: "..."}, nil
}
```

**Golden Test：**
```go
func TestScreener_2026_10_08(t *testing.T) {
    // 用10-08真实数据验证
    result := screener.Run(ctx, "2026-10-08")
    
    assert.Equal(t, 27, result.Count) // 已知产出27只
    
    // 验证部分股票代码
    codes := extractCodes(result.Stocks)
    assert.Contains(t, codes, "605133") // 世名科技
    assert.Contains(t, codes, "600825") // 新华传媒
}
```

**验收标准：**
- [ ] Golden test通过（10-08产出27只）
- [ ] 三条件过滤逻辑正确
- [ ] 周线/月线聚合算法正确
- [ ] Excel报告生成成功
- [ ] 执行时间 < 3分钟（5000只股票）

---

### Week 2: Lifecycle + Tachibana + CLI（Day 8-14）

#### 📌 Day 8-10: Lifecycle引擎（lifecycle）

**任务：** 实现19:30生命周期分析引擎

**文件清单：**
```
internal/lifecycle/
├── classifier.go         # 阶段分类器
├── swing.go              # 摆动点识别
├── probability.go        # 概率表查询
├── raw_lifespan.go       # ±25%生命视角
└── lifecycle_test.go
```

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
    volClass := lc.classifyVolume(vr13)
    
    // 5. 查概率表
    prob, ci, n := lc.probTable.Lookup(stage, volClass)
    
    // 6. ±25%生命视角
    rawView := lc.analyzeRawLifespan(weeklyBars)
    
    return &model.LifecycleState{...}, nil
}
```

**概率表加载：**
```go
// data/lifecycle_prob_table.json
{
  "2·早期_放量": {
    "survive_prob": 0.78,
    "ci_low": 0.75,
    "ci_high": 0.81,
    "sample_size": 1847
  },
  ...
}
```

**验收标准：**
- [ ] 摆动点识别算法正确（滞后4周）
- [ ] 九阶段分类逻辑正确
- [ ] 概率表查询返回正确
- [ ] ±25%生命算法正确
- [ ] Excel报告生成成功
- [ ] 执行时间 < 2分钟（50只股票）

---

#### 📌 Day 11-12: Tachibana引擎（tachibana）

**任务：** 实现20:00立花交易计划引擎

**文件清单：**
```
internal/tachibana/
├── planner.go            # 交易计划生成器
├── position.go           # 仓位计算
├── tracker.go            # 跟踪池管理
└── tachibana_test.go
```

**核心逻辑：**
```go
type TachibanaPlanner struct {
    config *TachibanaConfig
}

func (tp *TachibanaPlanner) BuildPlan(
    stock *model.Stock,
    lifecycle *model.LifecycleState,
    dailyBars []*model.DailyBar,
) (*model.TachibanaSignal, error) {
    // 1. 可交易性过滤
    if !tp.isTradeableStage(lifecycle.Stage) {
        return &model.TachibanaSignal{Decision: "排除"}, nil
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
    
    // 4. 计算梯价
    tier1 := ma20 - 0.75*atr14
    tier2 := ma20 - 1.50*atr14
    tier3 := ma20 - 2.25*atr14
    
    // 5. 计算止损和止盈
    ma60 := calc.MA(dailyBars, 60)
    stopLoss := math.Max(tier3-atr14, ma60)
    profitPrice := ma20 + 2*atr14
    
    // 6. 仓位计算
    shares := tp.calculatePosition(currentPrice, stopLoss)
    
    return &model.TachibanaSignal{...}, nil
}
```

**跟踪池状态机：**
```go
func (t *Tracker) UpdateDaily(tradeDate string) error {
    positions := t.store.LoadActivePositions()
    
    for _, pos := range positions {
        bar := t.store.GetDailyBar(pos.Code, tradeDate)
        
        // 检查成交
        t.checkFilled(pos, bar)
        
        // 检查止损
        if bar.Low <= pos.StopPrice {
            pos.Status = "stopped"
            t.sendAlert("止损触发", pos)
        }
        
        // 检查过期
        if pos.DaysSinceAdded >= 40 {
            pos.Status = "expired"
        }
        
        t.store.UpdatePosition(pos)
    }
    
    return nil
}
```

**验收标准：**
- [ ] 梯价计算正确（0.75/1.5/2.25 × ATR）
- [ ] 仓位计算正确（风险预算1%，40/30/30配比）
- [ ] 跟踪池状态机正确（成交/止损/过期）
- [ ] Excel报告生成成功（两个Sheet）
- [ ] 执行时间 < 1分钟

---

#### 📌 Day 13-14: CLI工具 + 复盘（cli + reviewer）

**任务：** 实现tb-cli命令行工具和复盘功能

**文件清单：**
```
cmd/tb-cli/
├── main.go               # CLI入口
├── cmd/
│   ├── root.go           # 根命令
│   ├── record.go         # 成交录入
│   ├── tracker.go        # 跟踪池管理
│   ├── review.go         # 复盘
│   ├── run.go            # 手动运行任务
│   └── import.go         # 数据导入

internal/reviewer/
├── reviewer.go           # 复盘引擎
└── reviewer_test.go
```

**CLI命令清单：**
```bash
# 成交录入
tb-cli record buy --code 600519 --price 1850 --qty 100 --note "梯①"
tb-cli record sell --code 600519 --price 1920 --qty 100 --note "止盈"

# 跟踪池管理
tb-cli tracker list                  # 列出跟踪池
tb-cli tracker update --date 2026-10-09  # 更新状态
tb-cli tracker show 600519           # 查看单只详情

# 复盘
tb-cli review daily --date 2026-10-09    # 生成当日复盘
tb-cli review weekly --week 2026-W41     # 周报
tb-cli review monthly --month 2026-10    # 月报

# 手动运行任务
tb-cli run screen --date 2026-10-09      # 运行初选
tb-cli run lifecycle --date 2026-10-09   # 运行生命周期
tb-cli run tachibana --date 2026-10-09   # 运行立花计划

# 数据导入
tb-cli import tdx --root H:\new_tdx64\vipdoc --codes 600519,000001
tb-cli import instruments --file data/universe.csv
```

**复盘核心逻辑：**
```go
type Reviewer struct {
    store store.Store
}

func (r *Reviewer) DailyReview(tradeDate string) (*model.Review, error) {
    // 1. 更新跟踪池
    r.updateTracker(tradeDate)
    
    // 2. 汇总成交
    txns := r.store.GetTransactionsByDate(tradeDate)
    filledCount := len(txns)
    avgSlippage := r.calculateAvgSlippage(txns)
    
    // 3. 账本校验
    equity := r.calculateEquity(tradeDate)
    prevEquity := r.store.GetEquity(prevDate)
    dailyReturn := (equity - prevEquity) / prevEquity
    
    // 4. 异常检测
    alerts := r.detectAnomalies(tradeDate)
    
    // 5. 生成报告
    review := &model.Review{
        TradeDate:    tradeDate,
        FilledCount:  filledCount,
        AvgSlippage:  avgSlippage,
        DailyReturn:  dailyReturn,
        Equity:       equity,
        ...
    }
    
    r.store.SaveReview(review)
    r.exportToExcel(review, tradeDate)
    
    return review, nil
}
```

**验收标准：**
- [ ] CLI所有命令可用
- [ ] 成交录入写入数据库
- [ ] 跟踪池更新逻辑正确
- [ ] 账本恒等式校验通过
- [ ] 复盘报告生成成功
- [ ] 命令响应时间 < 1秒

---

### 📦 第一期交付物检查清单

- [ ] **可执行文件**
  - [ ] `tb-cli.exe` (Windows)
  - [ ] 文件大小 < 20MB
  - [ ] 运行 `tb-cli --version` 正常

- [ ] **数据库**
  - [ ] `data/tradebuddy.db` 初始化成功
  - [ ] 所有表创建正确
  - [ ] 示例数据导入成功

- [ ] **配置文件**
  - [ ] `data/config.yaml` 有完整注释
  - [ ] 路径配置正确（TDX等）

- [ ] **文档**
  - [ ] README.md
  - [ ] AGENTS.md
  - [ ] docs/DESIGN.md
  - [ ] docs/REQUIRE.md
  - [ ] docs/TASK.md (本文件)
  - [ ] 用户手册（USER_GUIDE.md）

- [ ] **测试**
  - [ ] Golden test通过（10-08数据）
  - [ ] 单元测试覆盖率 ≥ 60%
  - [ ] 集成测试通过

- [ ] **报告模板**
  - [ ] reports/screen/模板.xlsx
  - [ ] reports/lifecycle/模板.xlsx
  - [ ] reports/tachibana/模板.xlsx
  - [ ] reports/review/模板.xlsx

---

## 🔄 第二期：简化GUI（1周）

**开始日期：** 2026-10-24  
**预计完成：** 2026-10-31

### 任务概览

#### Day 15-17: Wails项目初始化

- [ ] 安装Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [ ] 创建项目: `wails init -n tb-desktop -t vue`
- [ ] 配置前端（Vue 3 + Vite）
- [ ] 后端桥接（调用internal模块）

#### Day 18-19: 三视图开发

- [ ] 视图1：清单查看（表格+排序+过滤）
- [ ] 视图2：生命周期（阶段卡片+概率图）
- [ ] 视图3：交易计划（梯价表格+跟踪池）

#### Day 20-21: 手动触发任务

- [ ] 按钮：运行初选/生命周期/交易计划
- [ ] 进度条显示
- [ ] 结果刷新

### 第二期交付物

- [ ] `tb-desktop.exe`
- [ ] 用户操作手册更新

---

## 🚀 第三期：完整GUI + 打包（2周）

**开始日期：** 2026-11-01  
**预计完成：** 2026-11-15

### 任务概览

#### Week 4: 完整功能

- [ ] 仪表盘（任务状态/统计图表）
- [ ] 跟踪池管理界面
- [ ] 复盘报告查看器
- [ ] 成交录入表单

#### Week 5: 定时调度 + 打包

- [ ] 内置cron调度器
- [ ] 通知推送（桌面通知）
- [ ] 单文件打包（Windows/Linux/macOS）
- [ ] 安装程序制作（可选）

### 第三期交付物

- [ ] Windows: `tradebuddy.exe`
- [ ] Linux: `tradebuddy`
- [ ] macOS: `tradebuddy`
- [ ] 安装与配置手册

---

## 📊 进度追踪

### 当前状态（2026-10-09）

```
第一期进度：Day 0 完成
├─ ✅ 技术验证
├─ ✅ 项目初始化
├─ ✅ 核心文档
└─ 🔄 数据访问层（进行中）

预计完成日期：2026-10-23
```

### 里程碑

| 里程碑 | 预计日期 | 状态 |
|--------|----------|------|
| M1: 数据层完成 | 2026-10-11 | 🔄 进行中 |
| M2: Screener完成 | 2026-10-14 | ⏳ 待开始 |
| M3: Lifecycle完成 | 2026-10-18 | ⏳ 待开始 |
| M4: Tachibana完成 | 2026-10-21 | ⏳ 待开始 |
| M5: CLI完成 | 2026-10-23 | ⏳ 待开始 |
| M6: 第一期交付 | 2026-10-23 | ⏳ 待开始 |

---

## 🐛 已知问题与待办

### 技术债务

- [ ] Demo 2 Excel解析需要修复（数据行读取为0）
- [ ] Demo 3 TDX路径已修复但未重新验证

### 待决策事项

- [ ] 通达信数据更新策略（增量 vs 全量）
- [ ] Excel报告样式设计
- [ ] GUI配色方案

### 优化项（可后延）

- [ ] 并发抓取数据（第一期暂不实现）
- [ ] 数据缓存策略
- [ ] 日志系统完善

---

## 📝 开发规范

### Git提交规范

```
feat: 新功能
fix: 修复bug
docs: 文档更新
test: 测试相关
refactor: 重构
chore: 杂项（构建、依赖等）

示例：
feat(screener): 实现三条件过滤器
fix(store): 修复日K线重复插入问题
docs: 更新REQUIRE.md需求文档
```

### 代码审查清单

- [ ] 函数有文档注释
- [ ] 关键算法附原著引用
- [ ] 错误处理完整
- [ ] 单元测试覆盖
- [ ] 无硬编码路径

---

**下一步：** 开始Day 1-2任务 - 实现数据访问层
