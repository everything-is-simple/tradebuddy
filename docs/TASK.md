# TradeBuddy 开发任务清单 v2.0

> Development Task List（含回测引擎）
> Version 2.0 - 2026-10-09
> **重要变更：** 新增第二期回测引擎开发（2周）

---

## 📋 任务概览

**开发模式：** 四期迭代，每期独立交付  
**当前阶段：** 第一期（核心引擎 + CLI，2周）  
**开始日期：** 2026-10-09  
**总预计完成：** 2026-11-27（7周）

---

## 🎯 第一期：核心引擎 + CLI + 每日复盘（2周）

**目标：** 完成选股→分析→交易计划→**每日复盘**的自动化闭环

### Week 1: 数据层 + Screener（Day 1-7）

#### ✅ Day 0: 项目准备（已完成）
- [x] 技术验证（Demo 1-3）
- [x] 归档GLM文档
- [x] 创建项目结构
- [x] 编写核心文档（README/DESIGN/REQUIRE/TASK/BACKTEST）
- [x] 数据库Schema设计（0001_init.sql）
- [x] 数据模型定义（model/*.go）

#### ✅ Day 1-2: 数据访问层（store）- 已完成
- [x] 实现SQLite数据访问层
- [x] 10个数据访问模块
- [x] 完整单元测试（9个测试全部通过）
- [x] 已推送到GitHub

#### ✅ Day 3-4: 数据源适配（datasource）- 已完成

**任务：** 实现TDX文件读取 + 腾讯API适配 + 数据源管理器

**文件清单：**
```
internal/datasource/
├── manager.go            # 数据源管理器 ✅
├── manager_test.go       # 管理器测试 ✅
├── tdx/
│   ├── reader.go         # 读取.day文件 ✅
│   └── reader_test.go    # TDX测试 ✅
└── tencent/
    ├── api.go            # 腾讯前复权K线API ✅
    └── api_test.go       # API测试 ✅
```

**核心任务：**
- [x] TDX .day文件32字节定长记录解析 ✅
- [x] 腾讯API前复权K线获取（框架完成）✅
- [x] 限流控制（0.1s间隔）✅
- [x] 单元测试（使用sh600519.day验证）✅
- [x] 数据源管理器（TDX优先，API备份）✅

**验收标准：**
- [x] 能正确读取 sh600519.day（茅台数据）✅
- [x] 成功读取11722个.day文件 ✅
- [x] 限流策略有效 ✅
- [x] 数据格式正确（价格、成交量）✅
- [x] 测试全部通过（11/11）✅

**完成日期：** 2026-10-09  
**Git提交：** e18f731

#### ✅ Day 5-7: Screener引擎（screener）- 已完成

**任务：** 实现19:00强势股初选引擎

**开始时间：** 2026-10-09  
**完成时间：** 2026-10-10

**文件清单：**
```
internal/screener/
├── screener.go           # 主引擎
├── filter.go             # 三条件过滤器
├── aggregator.go         # 周线/月线聚合
└── screener_test.go

internal/calc/
├── ma.go                 # 均线计算
├── atr.go                # ATR计算
└── aggregator.go         # K线聚合
```

**Golden Test：**
```go
func TestScreener_2026_10_08(t *testing.T) {
    // 用10-08真实数据验证
    result := screener.Run(ctx, "2026-10-08")
    assert.Equal(t, 27, result.Count) // 已知产出27只
    assert.Contains(t, codes, "605133") // 世名科技
}
```

**验收标准：**
- [ ] Golden test通过（10-08产出27只）
- [ ] 三条件过滤逻辑正确
- [ ] 周线/月线聚合算法正确
- [ ] Excel报告生成成功
- [ ] 执行时间 < 3分钟（5000只股票）

---

### Week 2: Lifecycle + Tachibana + CLI + Review（Day 8-14）

#### ✅ Day 11-14: Tachibana立花提示引擎（tachibana）- 已完成

**任务：** 实现20:00立花交易提示引擎（简化版 Phase 1）

**开始时间：** 2026-10-10  
**完成时间：** 2026-10-10

**设计文档：** docs/TACHIBANA-DESIGN.md

**实现范围：**
- ✅ 5种交易依据分类
- ✅ 决策规则引擎
- ✅ 价格区间计算（ATR基础）
- ✅ 基于Lifecycle指标判断
- ✅ Excel报告生成
- ✅ 数据库迁移

**核心任务：**
- [x] 设计文档 ✅
- [x] 数据模型 ✅
- [x] 决策引擎 ✅
- [x] 价格区间计算 ✅
- [x] 数据库层 ✅
- [x] 数据库迁移 ✅
- [x] Excel报告生成 ✅
- [x] 单元测试 ✅

**验收标准：**
- [x] 5种信号类型识别 ✅
- [x] 决策规则正确 ✅
- [x] 价格区间计算准确 ✅
- [x] Excel报告生成成功 ✅
- [x] 单元测试通过 ✅

**完成日期：** 2026-10-10  
**Git提交：** 待提交

---

**任务：** 实现20:00立花交易提示引擎（简化版 Phase 1）

**开始时间：** 2026-10-10

**设计文档：** docs/TACHIBANA-DESIGN.md

**实现范围：**
- ✅ 5种交易依据分类
- ✅ 决策规则引擎
- ✅ 价格区间计算（ATR基础）
- ✅ 基于Lifecycle指标判断
- ⏳ Excel报告生成
- ⏳ 数据库迁移

**核心任务：**
- [x] 设计文档 ✅
- [x] 数据模型 ✅
- [x] 决策引擎 ✅
- [x] 价格区间计算 ✅
- [x] 单元测试框架 ✅
- [ ] 数据库迁移 ⏳
- [ ] Excel报告生成 ⏳
- [ ] 集成测试 ⏳

**验收标准：**
- [x] 5种信号类型识别 ✅
- [x] 决策规则正确 ✅
- [x] 价格区间计算准确 ✅
- [ ] Excel报告生成成功 ⏳
- [ ] 单元测试通过 ⏳
- [ ] 集成测试通过 ⏳

**进度：** 60%完成

---

#### ✅ Day 8-10: Lifecycle引擎（lifecycle）- 已完成

**任务：** 实现19:30生命周期分析引擎（Phase 1 简化版）

**开始时间：** 2026-10-10  
**完成时间：** 2026-10-10

**设计文档：** docs/LIFECYCLE-DESIGN.md

**实现范围：**
- ✅ 基于 MALF v2.1 Lifespan 层简化实现
- ✅ 核心指标：span_days, price_range, atr_normalized
- ✅ 市场排名：percentile_rank
- ✅ 综合评分：lifecycle_score + grade
- ✅ Excel报告生成
- ❌ 不实现完整状态机（Phase 2）

**核心任务：**
- [x] 实现生命周期指标计算 ✅
- [x] 实现市场排名算法 ✅
- [x] 实现综合评分系统 ✅
- [x] Excel报告生成 ✅
- [x] 单元测试 ✅
- [ ] 集成测试 ⚠️ 需完整数据

**验收标准：**
- [x] 指标计算正确 ✅
- [x] 排名算法准确 ✅
- [x] 评分系统合理 ✅
- [x] Excel报告生成成功 ✅
- [x] 单元测试通过（2/2）✅
- [ ] 集成测试通过 ⚠️ 待完整数据

**完成日期：** 2026-10-10  
**Git提交：** 待提交

**文件清单：**
```
internal/lifecycle/
├── classifier.go         # 阶段分类器
├── swing.go              # 摆动点识别
├── probability.go        # 概率表查询
├── raw_lifespan.go       # ±25%生命视角
└── lifecycle_test.go

data/
└── lifecycle_prob_table.json  # 概率表
```

**验收标准：**
- [ ] 摆动点识别算法正确（滞后4周）
- [ ] 九阶段分类逻辑正确
- [ ] 概率表查询返回正确
- [ ] ±25%生命算法正确
- [ ] Excel报告生成成功
- [ ] 执行时间 < 2分钟（50只股票）

#### 📌 Day 11-12: Tachibana引擎（tachibana）

**任务：** 实现20:00立花交易计划引擎

**文件清单：**
```
internal/tachibana/
├── planner.go            # 交易计划生成器
├── position.go           # 仓位计算
└── tachibana_test.go

internal/tracker/
├── tracker.go            # 跟踪池状态机
└── tracker_test.go
```

**验收标准：**
- [ ] 梯价计算正确（0.75/1.5/2.25 × ATR）
- [ ] 仓位计算正确（风险预算1%，40/30/30配比）
- [ ] 跟踪池状态机正确（成交/止损/过期）
- [ ] Excel报告生成成功（两个Sheet）
- [ ] 执行时间 < 1分钟

#### 📌 Day 13-14: CLI工具 + 每日复盘（cli + reviewer）⭐

**任务：** 实现tb-cli命令行工具和**每日复盘功能**

**文件清单：**
```
cmd/tb-cli/
├── main.go               # CLI入口
├── cmd/
│   ├── root.go           # 根命令
│   ├── record.go         # 成交录入
│   ├── tracker.go        # 跟踪池管理
│   ├── review.go         # 每日复盘 ⭐
│   ├── run.go            # 手动运行任务
│   └── import.go         # 数据导入

internal/reviewer/        # ⭐ 新增
├── reviewer.go           # 复盘引擎
└── reviewer_test.go
```

**每日复盘功能：** ⭐
```go
// 核心功能
- 更新跟踪池状态（检查成交/止损/过期）
- 账本恒等式校验
- 计算每日收益率
- 滑点统计
- 生成复盘报告Excel
```

**CLI命令清单：**
```bash
# 成交录入
tb-cli record buy --code 600519 --price 1850 --qty 100 --note "梯①"

# 每日复盘 ⭐
tb-cli review daily --date 2026-10-09
tb-cli review weekly --week 2026-W41
tb-cli review monthly --month 2026-10

# 跟踪池管理
tb-cli tracker list
tb-cli tracker update --date 2026-10-09
```

**验收标准：**
- [ ] CLI所有命令可用
- [ ] 成交录入写入数据库
- [ ] **跟踪池更新逻辑正确** ⭐
- [ ] **账本恒等式校验通过** ⭐
- [ ] **复盘报告生成成功** ⭐
- [ ] 命令响应时间 < 1秒

---

### 📦 第一期交付物检查清单

- [x] **数据模型**（internal/model/）✅
- [x] **数据访问层**（internal/store/）✅
- [x] **数据源适配**（internal/datasource/）✅
- [ ] 计算工具（internal/calc/）
- [ ] Screener引擎（internal/screener/）
- [ ] Lifecycle引擎（internal/lifecycle/）
- [ ] Tachibana引擎（internal/tachibana/）
- [ ] Tracker跟踪池（internal/tracker/）
- [ ] **Reviewer复盘引擎**（internal/reviewer/）⭐
- [ ] Reporter报告生成（internal/reporter/）
- [ ] CLI工具（cmd/tb-cli/）

**测试：**
- [x] Store层单元测试（9/9通过）✅
- [x] DataSource单元测试（11/11通过）✅
- [ ] Screener golden test（10-08数据）
- [ ] 其他模块单元测试

**文档：**
- [x] README.md
- [x] docs/DESIGN.md
- [x] docs/REQUIRE.md
- [x] docs/TASK.md
- [x] docs/BACKTEST.md ⭐
- [ ] USER_GUIDE.md

---

## 🔬 第二期：回测引擎（2周）⭐ 新增

**开始日期：** 2026-10-24  
**预计完成：** 2026-11-07

**目标：** 历史策略验证 + 参数优化

### Week 3: 回测核心引擎（Day 15-21）

#### 📌 Day 15-16: 时间回放引擎

**文件清单：**
```
internal/backtest/
├── backtester.go         # 回测主引擎
├── config.go             # 回测配置
├── simulator.go          # 交易模拟器
└── backtest_test.go
```

**核心功能：**
```go
type Backtester struct {
    store       *store.Store
    screener    *screener.Screener
    lifecycle   *lifecycle.Classifier
    tachibana   *tachibana.Planner
    simulator   *Simulator
}

// 回测主流程
func (b *Backtester) Run() (*BacktestResult, error) {
    for date := range tradingDays {
        // 1. 运行选股
        screenResults := b.screener.Run(date)
        
        // 2. 生命周期分析
        lifecycleStates := b.lifecycle.Analyze(screenResults)
        
        // 3. 生成交易信号
        signals := b.tachibana.GenerateSignals(...)
        
        // 4. 模拟成交
        b.simulator.Execute(signals, date)
        
        // 5. 检查止损/止盈
        b.simulator.CheckExits(date)
        
        // 6. 记录权益
        b.recordEquity(date)
    }
}
```

**验收标准：**
- [ ] 能回测指定日期区间
- [ ] 避免前视偏差（代码审查）
- [ ] 模拟成交逻辑正确
- [ ] 止损/止盈检查准确

#### 📌 Day 17-18: 交易模拟器

**核心功能：**
```go
type Simulator struct {
    cash            float64
    positions       map[string]*Position
    transactions    []*Transaction
    equity          []EquityPoint
}

// 模拟成交（含滑点和手续费）
func (s *Simulator) ExecuteSignal(signal *Signal) *Transaction

// 检查止损/止盈
func (s *Simulator) CheckExits(date string) []*Transaction

// 计算每日权益
func (s *Simulator) CalculateEquity(date string) float64
```

**验收标准：**
- [ ] 滑点模型正确（默认0.2%）
- [ ] 手续费计算准确（0.03%）
- [ ] 现金和持仓更新正确
- [ ] 权益曲线连续

#### 📌 Day 19-21: 绩效统计

**文件清单：**
```
internal/backtest/
├── metrics.go            # 绩效指标计算
├── report.go             # 回测报告生成
└── metrics_test.go
```

**核心指标：**
```go
type BacktestResult struct {
    // 收益
    TotalReturn      float64
    AnnualReturn     float64
    SharpeRatio      float64
    
    // 风险
    MaxDrawdown      float64
    Volatility       float64
    
    // 交易
    TotalTrades      int
    WinRate          float64
    ProfitFactor     float64
    
    // 明细
    EquityCurve      []EquityPoint
    Transactions     []*Transaction
}
```

**验收标准：**
- [ ] 夏普比率计算正确
- [ ] 最大回撤计算准确
- [ ] 胜率统计正确
- [ ] Excel报告导出成功

---

### Week 4: 参数优化 + CLI（Day 22-28）

#### 📌 Day 22-24: 参数优化引擎

**文件清单：**
```
internal/backtest/
├── optimizer.go          # 参数优化器
└── optimizer_test.go
```

**核心功能：**
```go
// 网格搜索
type ParamGrid struct {
    MinPctChange    []float64  // [5.0, 6.0, 7.0]
    LookbackDays    []int      // [15, 20, 25]
    RiskPerStock    []float64  // [0.005, 0.01, 0.015]
}

func (b *Backtester) OptimizeParams(grid *ParamGrid) *OptimizationResult
```

**验收标准：**
- [ ] 网格搜索逻辑正确
- [ ] 能找到最优参数组合
- [ ] 避免过度拟合（样本外验证）

#### 📌 Day 25-26: 回测CLI命令

**文件清单：**
```
cmd/tb-cli/cmd/
├── backtest.go           # 回测命令
└── backtest_compare.go   # 对比命令
```

**CLI命令：**
```bash
# 运行回测
tb-cli backtest run \
    --name "2023年回测" \
    --start 2023-01-01 \
    --end 2023-12-31 \
    --capital 1000000

# 参数优化
tb-cli backtest optimize \
    --start 2022-01-01 \
    --end 2022-12-31

# 查看回测列表
tb-cli backtest list

# 对比实盘vs回测
tb-cli review compare \
    --live-start 2026-09-01 \
    --backtest-id 1
```

**验收标准：**
- [ ] 所有回测命令可用
- [ ] 回测结果存入数据库
- [ ] 报告导出成功

#### 📌 Day 27-28: 数据库扩展 + 测试

**新增表：**
```sql
-- 回测运行记录
CREATE TABLE backtest_runs (...)

-- 回测交易明细
CREATE TABLE backtest_transactions (...)

-- 回测权益曲线
CREATE TABLE backtest_equity (...)
```

**验收标准：**
- [ ] 数据库迁移成功
- [ ] 回测数据正确存储
- [ ] 单元测试覆盖率 ≥ 60%

---

### 📦 第二期交付物检查清单

- [ ] **回测引擎**（internal/backtest/）
- [ ] **参数优化器**（optimizer.go）
- [ ] **回测CLI命令**（cmd/tb-cli/cmd/backtest.go）
- [ ] **数据库扩展**（新增3张表）
- [ ] **回测报告Excel模板**
- [ ] **回测文档更新**（USER_GUIDE.md）

**测试：**
- [ ] 回测引擎单元测试
- [ ] 回测2023年数据（完整验证）
- [ ] 参数优化测试

---

## 🎨 第三期：简化GUI（1周）

**开始日期：** 2026-11-08  
**预计完成：** 2026-11-14

### Week 5: Wails最小界面（Day 29-35）

#### Day 29-31: Wails项目初始化
- [ ] 安装Wails CLI
- [ ] 创建项目
- [ ] 配置前端（Vue 3 + Vite）
- [ ] 后端桥接

#### Day 32-33: 三视图开发
- [ ] 视图1：清单查看（表格+排序）
- [ ] 视图2：生命周期（阶段卡片）
- [ ] 视图3：交易计划（梯价表格）

#### Day 34-35: 手动触发任务
- [ ] 按钮：运行初选/生命周期/交易计划
- [ ] 进度条显示
- [ ] 结果刷新

### 第三期交付物
- [ ] `tb-desktop.exe`

---

## 🚀 第四期：完整GUI + 打包（2周）

**开始日期：** 2026-11-15  
**预计完成：** 2026-11-27

### Week 6-7: 完整功能（Day 36-49）

#### Week 6: 仪表盘 + 回测可视化
- [ ] 仪表盘（任务状态/统计图表）
- [ ] 跟踪池管理界面
- [ ] **回测报告可视化** ⭐
  - [ ] 权益曲线图
  - [ ] 回撤图
  - [ ] 月度收益热力图
- [ ] 复盘报告查看器

#### Week 7: 定时调度 + 打包
- [ ] 内置cron调度器
- [ ] 通知推送（桌面通知）
- [ ] 单文件打包（Windows/Linux/macOS）
- [ ] 安装程序制作（可选）

### 第四期交付物
- [ ] Windows: `tradebuddy.exe`
- [ ] Linux: `tradebuddy`
- [ ] macOS: `tradebuddy`
- [ ] 安装与配置手册

---

## 📈 里程碑进度

| 里程碑 | 预计日期 | 状态 |
|--------|----------|------|
| M1: 数据层完成 | 2026-10-11 | ✅ 已完成 |
| M2: Screener完成 | 2026-10-14 | ⏳ 待开始 |
| M3: Lifecycle完成 | 2026-10-18 | ⏳ 待开始 |
| M4: Tachibana+Review完成 | 2026-10-23 | ⏳ 待开始 |
| M5: 第一期交付 | 2026-10-23 | ⏳ 待开始 |
| M6: 回测引擎完成 | 2026-11-07 | ⏳ 待开始 |
| M7: 第二期交付 | 2026-11-07 | ⏳ 待开始 |
| M8: 简化GUI完成 | 2026-11-14 | ⏳ 待开始 |
| M9: 完整系统交付 | 2026-11-27 | ⏳ 待开始 |

---

## 📊 工作量估算

| 阶段 | 周数 | 工作日 | 模块数 |
|------|-----|--------|--------|
| 第一期（核心+复盘） | 2周 | 14天 | 8个模块 |
| 第二期（回测引擎） | 2周 | 14天 | 4个模块 ⭐ |
| 第三期（简化GUI） | 1周 | 7天 | 3个视图 |
| 第四期（完整系统） | 2周 | 14天 | 完善 |
| **总计** | **7周** | **49天** | **15+模块** |

---

## 🎯 关键变更说明

### v2.0 主要变更（2026-10-09）

1. ⭐ **新增第二期：回测引擎**（2周）
   - 历史策略验证
   - 参数优化
   - 绩效统计

2. ⭐ **强化第一期：加入每日复盘**
   - Day 13-14 实现复盘引擎
   - 账本恒等式校验
   - 跟踪池状态机

3. 📅 **总工期延长：5周 → 7周**
   - 原3期 → 现4期
   - 回测功能作为独立阶段

4. 📋 **数据库扩展**
   - 新增3张回测相关表

---

**下一步：** Day 5-7 Screener引擎开发

准备好开始了吗？ 🚀
