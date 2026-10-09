# Lifecycle 引擎设计文档（简化版 Phase 1）

> 基于 MALF v2.1 Lifespan 层的简化实现
> 
> 创建日期：2026-10-10
> 
> 实现范围：Lifespan 核心指标，不含完整状态机

---

## 一、设计目标

### 1.1 业务需求

**19:30 生命周期分析：** 对初选的强势股进行波段生命周期评估，识别：
- 波段持续时间（天数）
- 价格幅度（涨跌幅）
- 波段强度（ATR 标准化）
- 结构位置（距离关键价格点）
- 市场排名（百分位）

### 1.2 Phase 1 范围（简化版）

**实现内容：**
- ✅ 生命周期指标计算（span_days, price_range, atr_normalized）
- ✅ 市场排名（percentile_rank，全市场单池）
- ✅ 关键价格位置（距20日高、距52周高）
- ✅ 波段强度评分

**不实现内容（Phase 2 升级）：**
- ❌ Core 状态机（Pivot/Wave/Guard/Break）
- ❌ Range 区间识别
- ❌ Structural Position（rank 向量差）
- ❌ 分形 k=2 延迟确认
- ❌ 同方向分池排名

### 1.3 替代度评估

| 功能 | 完整MALF | 简化版 | 替代度 |
|------|---------|--------|--------|
| 生命周期指标 | ✅ | ✅ | 70% |
| 排名计算 | ✅ | ✅ | 60% |
| 波段强度 | ✅ | ✅ | 80% |
| 结构位置 | ✅ | ❌ | 0% |
| 状态机 | ✅ | ❌ | 0% |
| **总体** | 100% | **~70%** | - |

---

## 二、核心指标定义

### 2.1 生命周期指标

#### LifecycleMetrics 结构

```go
type LifecycleMetrics struct {
    Code       string  // 股票代码
    Name       string  // 股票名称
    TradeDate  string  // 交易日期
    
    // 时间维度
    SpanDays       int     // 波段持续天数（从突破20日高到现在）
    DaysSince20DH  int     // 距离突破20日高点的天数
    
    // 价格维度
    PriceRange     float64 // 价格幅度（当前价 - 低点价）
    PriceRangePct  float64 // 价格幅度百分比
    ATRNormalized  float64 // ATR标准化幅度（price_range / atr14）
    
    // 结构位置
    DistFrom20DH   float64 // 距20日高点（%）
    DistFrom52WH   float64 // 距52周高点（%）
    
    // 排名指标
    SpanRank       float64 // 持续时间排名（percentile）
    RangeRank      float64 // 价格幅度排名（percentile）
    ATRRank        float64 // ATR标准化排名（percentile）
    
    // 综合评分
    LifecycleScore float64 // 生命周期综合评分（0-100）
    Grade          string  // 评级（A/B/C/D）
}
```

### 2.2 计算公式

#### 2.2.1 波段持续天数（SpanDays）

```
从 Screener 结果中的交易日向前追溯：
1. 找到突破20日高点的那一天（breakout_date）
2. span_days = trade_date - breakout_date
```

#### 2.2.2 价格幅度（PriceRange）

```
1. low_price = 突破前20天的最低价
2. current_price = 当前收盘价
3. price_range = current_price - low_price
4. price_range_pct = (price_range / low_price) * 100
```

#### 2.2.3 ATR 标准化（ATRNormalized）

```
1. 计算14日ATR（使用 calc.ATR）
2. atr_normalized = price_range / atr14
3. 意义：排除个股波动率差异，标准化幅度
```

#### 2.2.4 百分位排名（PercentileRank）

```go
func PercentileRank(value float64, sample []float64) float64 {
    if len(sample) < 30 {
        return -1.0 // 样本不足
    }
    
    count := 0
    for _, v := range sample {
        if v < value {
            count++
        }
    }
    
    return float64(count) / float64(len(sample))
}
```

#### 2.2.5 生命周期评分（LifecycleScore）

```
综合评分（0-100分）：

score = w1 * span_rank * 100 +
        w2 * range_rank * 100 +
        w3 * atr_rank * 100

权重默认：
w1 = 0.2  // 持续时间权重
w2 = 0.4  // 价格幅度权重
w3 = 0.4  // ATR标准化权重

评级：
A: score >= 80
B: score >= 60
C: score >= 40
D: score < 40
```

---

## 三、数据流程

### 3.1 输入

```
来源：Screener 的筛选结果
格式：[]*model.ScreenResult

筛选条件：
- 涨幅 ≥ 6%
- 突破20日最高价
- 距52周高点 ≤ 25%
```

### 3.2 处理流程

```
1. 获取筛选结果（N只股票）
   ↓
2. 对每只股票：
   a. 获取历史K线数据（至少60天）
   b. 计算生命周期指标
   c. 记录到数组
   ↓
3. 计算市场排名：
   a. 收集所有股票的 span_days
   b. 收集所有股票的 price_range_pct
   c. 收集所有股票的 atr_normalized
   d. 计算每只股票的 percentile_rank
   ↓
4. 计算综合评分
   ↓
5. 保存到数据库（lifecycle_analysis 表）
   ↓
6. 生成Excel报告
```

### 3.3 输出

```
1. 数据库记录（lifecycle_analysis 表）
2. Excel报告（lifecycle_result_YYYYMMDD.xlsx）
3. 返回：[]*LifecycleMetrics
```

---

## 四、实现要点

### 4.1 防止前视偏差

```go
// ❌ 错误：使用未来数据
func calcSpanDays(bars []*model.DailyBar, tradeDate string) int {
    // 不能使用 tradeDate 之后的数据
}

// ✅ 正确：只使用历史数据
func calcSpanDays(bars []*model.DailyBar, tradeDate string) int {
    targetIdx := findBarIndex(bars, tradeDate)
    historicalBars := bars[:targetIdx+1] // 只用到 tradeDate 的数据
    // ...
}
```

### 4.2 样本量不足处理

```go
if len(sample) < 30 {
    // 返回 -1.0 表示样本不足
    // 前端显示为 "N/A"
    return -1.0
}
```

### 4.3 数据缺失处理

```go
// 如果某只股票数据不足60天
if len(bars) < 60 {
    // 跳过该股票，记录警告日志
    log.Printf("Skipped %s: insufficient data (%d bars)", code, len(bars))
    continue
}
```

---

## 五、数据库Schema

```sql
-- lifecycle_analysis 表（已在 0001_init.sql 中定义）
CREATE TABLE IF NOT EXISTS lifecycle_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trade_date TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT,
    
    -- 时间维度
    span_days INTEGER,
    days_since_20dh INTEGER,
    
    -- 价格维度
    price_range REAL,
    price_range_pct REAL,
    atr_normalized REAL,
    
    -- 结构位置
    dist_from_20dh REAL,
    dist_from_52wh REAL,
    
    -- 排名指标
    span_rank REAL,
    range_rank REAL,
    atr_rank REAL,
    
    -- 综合评分
    lifecycle_score REAL,
    grade TEXT,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trade_date, code)
);

CREATE INDEX idx_lifecycle_trade_date ON lifecycle_analysis(trade_date);
CREATE INDEX idx_lifecycle_score ON lifecycle_analysis(lifecycle_score DESC);
CREATE INDEX idx_lifecycle_grade ON lifecycle_analysis(grade);
```

---

## 六、测试策略

### 6.1 单元测试

```
1. 指标计算测试
   - TestCalcSpanDays
   - TestCalcPriceRange
   - TestCalcATRNormalized
   - TestPercentileRank

2. 边界条件测试
   - 样本不足（N < 30）
   - 数据不足（bars < 60）
   - 相等值处理
   - 空输入

3. 排名测试
   - 最小值排名 = 0
   - 最大值排名 = 1
   - 中位数排名 ≈ 0.5
```

### 6.2 集成测试

```
使用 2026-06-15 真实数据：
1. 输入：151只筛选结果
2. 处理：计算生命周期指标
3. 验证：
   - 所有股票都有评分
   - 排名分布合理（不全是0或1）
   - 评级分布合理（A/B/C/D都有）
   - Excel报告生成成功
```

---

## 七、Phase 2 升级路径

当需要完整 MALF 功能时，升级步骤：

```
1. 保留 Phase 1 的简化接口（向后兼容）
2. 实现 Core 状态机（malf/core.go）
3. 实现 Range 层（malf/range.go）
4. 实现 Structural Position（malf/position.go）
5. 新增接口：GetMALFFullAnalysis()
6. 原接口：GetLifecycleMetrics() 继续可用
```

**迁移成本：** Phase 1 代码可复用约 40%，主要是指标计算和排名逻辑。

---

## 八、参考文档

- MALF v2.1 Definitive: `H:\malf-Definitive\MALF_Definitive_v2_1-deepseek-20260726\`
- MALF Engine: `H:\malf-components\malf-engine\`
- 本项目设计文档: `docs/DESIGN.md`
- 任务清单: `docs/TASK.md`
