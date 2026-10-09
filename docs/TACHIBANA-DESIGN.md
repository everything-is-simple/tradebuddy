# Tachibana 立花提示引擎设计文档（简化版 Phase 1）

> 基于立花义正波段交易法的简化实现
> 
> 创建日期：2026-10-10
> 
> 实现范围：核心交易提示，不含完整 Position Management

---

## 一、设计目标

### 1.1 业务需求

**20:00 立花交易提示：** 对生命周期分析后的股票给出立花式交易提示：
- 试探建仓机会（trend_probe_entry）
- 同向加码机会（trend_confirmation_add）
- 分批减仓提示（distribution_reduce）
- 节奏失败警示（exit_on_rhythm_failure）
- 等待观望建议（wait_no_action）

### 1.2 Phase 1 范围（简化版）

**实现内容：**
- ✅ 5 种核心交易依据分类识别
- ✅ 基于 Lifecycle 指标的条件判断
- ✅ 回踩买入区间计算（ATR 基础）
- ✅ 风险提示和持仓建议
- ✅ 交易理由说明

**不实现内容（Phase 2 完整版）：**
- ❌ 完整 MALF 前置认知过滤器
- ❌ Position Management（中心单/加码单/锁单）
- ❌ 心理状态建模
- ❌ 历史交易记录跟踪
- ❌ 双侧库存管理

### 1.3 与完整立花法的关系

```
完整立花法 = MALF 状态机 + 前置过滤器 + Position Management + 心理建模

Phase 1 简化版 ≈ 25-30% 覆盖度

核心价值：
- 给出"值得关注"的提示
- 不是交易信号，是参考建议
- 帮助用户理解立花式思维
```

---

## 二、核心逻辑

### 2.1 交易依据分类

#### A. 试探建仓（trend_probe_entry）

**条件：**
```
1. Lifecycle Grade = A 或 B（波段质量好）
2. 距20日高点回落 < 10%（轻度回踩）
3. ATR 标准化 > 2.0（波动充足）
4. SpanDays < 20天（波段年轻）
```

**提示信息：**
```
【试探建仓机会】
- 波段处于健康推进状态
- 当前轻度回踩，可考虑小仓试探
- 建议买入区间：XXX - XXX（基于 ATR）
- 风险：不要重仓，分批建仓
```

#### B. 同向加码（trend_confirmation_add）

**条件：**
```
1. Lifecycle Grade = A（波段强势）
2. 价格创新高（DistFrom20DH > 0）
3. ATR Rank > 0.7（波动强度领先）
4. SpanDays > 5天且 < 30天（有一定持续）
```

**提示信息：**
```
【同向加码机会】
- 波段持续推进，创近期新高
- 节奏健康，可考虑加码
- 建议加码区间：XXX - XXX
- 风险：保持分批纪律，设止损
```

#### C. 分批减仓（distribution_reduce）

**条件：**
```
1. SpanDays > 30天（波段较长）
2. DistFrom52WH > -5%（接近52周高点）
3. ATR Rank < 0.3（波动衰减）
4. PriceRangePct > 40%（涨幅较大）
```

**提示信息：**
```
【分批减仓提示】
- 波段持续时间较长，获利较多
- 接近历史高点，波动衰减
- 建议分批兑现利润
- 风险：不要一次性清仓，保留观察仓
```

#### D. 节奏失败（exit_on_rhythm_failure）

**条件：**
```
1. Lifecycle Grade = C 或 D（波段弱化）
2. 距20日高点回落 > 15%（深度回踩）
3. ATR Rank 快速下降（相比前期）
4. SpanRank < 0.3（持续性差）
```

**提示信息：**
```
【节奏失败警示】
- 波段推进力度减弱
- 价格深度回踩，结构疲弱
- 建议考虑退出或减仓
- 风险：不要死扛，承认失败
```

#### E. 等待观望（wait_no_action）

**条件：**
```
1. Lifecycle Grade = N/A（样本不足）
2. 或：SpanDays < 3天（波段太新）
3. 或：结构位置不明确
```

**提示信息：**
```
【等待观望】
- 当前结构不够清晰
- 不强迫市场给出机会
- 建议继续观察，不急于行动
- 风险：耐心等待更好的时机
```

---

## 三、核心指标计算

### 3.1 回踩买入区间

**基于 ATR 的价格带计算：**

```go
func CalcEntryZone(currentPrice float64, atr14 float64) (lower, upper float64) {
    // 理想买入区间：当前价 -1.5ATR 到 -0.5ATR
    lower = currentPrice - 1.5*atr14
    upper = currentPrice - 0.5*atr14
    return lower, upper
}
```

**说明：**
- 不追高，等待回踩
- 回踩幅度基于个股波动率（ATR）
- 不是精确价位，是一个区间

### 3.2 加码区间

```go
func CalcAddZone(currentPrice float64, atr14 float64) (lower, upper float64) {
    // 加码区间：当前价 +0.5ATR 到 +1.5ATR
    // （等待突破确认后）
    lower = currentPrice + 0.5*atr14
    upper = currentPrice + 1.5*atr14
    return lower, upper
}
```

### 3.3 止损位

```go
func CalcStopLoss(entryPrice float64, atr14 float64) float64 {
    // 止损位：入场价 -2ATR
    return entryPrice - 2*atr14
}
```

---

## 四、数据结构

### 4.1 TachibanaSignal 结构

```go
type TachibanaSignal struct {
    Code       string   // 股票代码
    Name       string   // 股票名称
    TradeDate  string   // 交易日期
    
    // 信号类型
    SignalType string   // trend_probe_entry / trend_confirmation_add / 
                         // distribution_reduce / exit_on_rhythm_failure / 
                         // wait_no_action
    
    // 价格区间
    EntryZoneLow   float64  // 建议买入下限
    EntryZoneHigh  float64  // 建议买入上限
    StopLoss       float64  // 止损位
    
    // 依据指标
    LifecycleGrade string   // A/B/C/D
    LifecycleScore float64  // 综合评分
    SpanDays       int      // 波段天数
    DistFrom20DH   float64  // 距20日高点
    ATRNormalized  float64  // ATR标准化
    
    // 提示信息
    Title       string   // 提示标题
    Description string   // 详细说明
    Risk        string   // 风险提示
    Suggestion  string   // 操作建议
    
    // 置信度
    Confidence  string   // High / Medium / Low
}
```

---

## 五、决策规则表

### 5.1 试探建仓决策树

```
IF LifecycleGrade IN [A, B]
  AND DistFrom20DH BETWEEN -10% AND 0%
  AND ATRNormalized > 2.0
  AND SpanDays < 20
THEN
  SignalType = "trend_probe_entry"
  Confidence = "High" IF LifecycleGrade = A ELSE "Medium"
  EntryZone = CalcEntryZone(currentPrice, atr14)
```

### 5.2 同向加码决策树

```
IF LifecycleGrade = A
  AND DistFrom20DH > 0%
  AND ATRRank > 0.7
  AND SpanDays BETWEEN 5 AND 30
THEN
  SignalType = "trend_confirmation_add"
  Confidence = "High"
  AddZone = CalcAddZone(currentPrice, atr14)
```

### 5.3 分批减仓决策树

```
IF SpanDays > 30
  AND DistFrom52WH > -5%
  AND ATRRank < 0.3
  AND PriceRangePct > 40%
THEN
  SignalType = "distribution_reduce"
  Confidence = "Medium"
  Suggestion = "分3-5次逐步减仓"
```

### 5.4 节奏失败决策树

```
IF LifecycleGrade IN [C, D]
  AND DistFrom20DH < -15%
  AND (ATRRank < 0.3 OR SpanRank < 0.3)
THEN
  SignalType = "exit_on_rhythm_failure"
  Confidence = "High"
  Suggestion = "考虑退出或大幅减仓"
```

---

## 六、输出格式

### 6.1 控制台输出

```
═══════════════════════════════════════════════
📊 立花交易提示 - 2026-10-10
═══════════════════════════════════════════════

【试探建仓机会】sh600519 贵州茅台 ⭐⭐⭐
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  当前价格：1800.00
  建议买入区间：1760.00 - 1790.00
  止损位：1720.00
  
  📈 波段状态：
     - 生命周期评级：A（85.5分）
     - 波段天数：15天（年轻健康）
     - 距20日高：-5.2%（轻度回踩）
     - ATR标准化：3.2（波动充足）
  
  💡 操作建议：
     - 小仓试探，不追高
     - 回踩到买入区间时分批建仓
     - 设置止损，保持纪律
  
  ⚠️ 风险提示：
     - 不要重仓，分批建仓
     - 破止损坚决退出
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【等待观望】sz000001 平安银行
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  当前价格：12.50
  
  📈 波段状态：
     - 生命周期评级：C（45.2分）
     - 波段天数：8天
     - 结构位置不够清晰
  
  💡 操作建议：
     - 继续观察，不急于行动
     - 等待更好的时机
  
  ⚠️ 风险提示：
     - 不强迫市场给出机会
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

统计：
  试探建仓：3只
  同向加码：2只
  分批减仓：1只
  节奏失败：0只
  等待观望：5只
```

### 6.2 Excel报告

**Sheet 1: 交易提示**
```
代码 | 名称 | 信号类型 | 置信度 | 当前价 | 买入区间 | 止损位 | 评级 | 评分 | 建议
```

**Sheet 2: 按类型分组**
```
试探建仓：
  - sh600519 贵州茅台 ⭐⭐⭐
  - ...

同向加码：
  - ...
```

---

## 七、实现要点

### 7.1 与 Lifecycle 的集成

```go
func (t *Tachibana) Run(
    ctx context.Context, 
    tradeDate string, 
    lifecycleMetrics []*model.LifecycleMetrics,
) ([]*model.TachibanaSignal, error) {
    
    var signals []*model.TachibanaSignal
    
    for _, m := range lifecycleMetrics {
        signal := t.analyzeOne(ctx, m, tradeDate)
        if signal != nil {
            signals = append(signals, signal)
        }
    }
    
    return signals, nil
}
```

### 7.2 决策优先级

```
1. 先判断节奏失败（exit_on_rhythm_failure）
2. 再判断分批减仓（distribution_reduce）
3. 再判断同向加码（trend_confirmation_add）
4. 再判断试探建仓（trend_probe_entry）
5. 默认等待观望（wait_no_action）
```

### 7.3 置信度计算

```
High：所有核心条件都满足，且 LifecycleGrade = A
Medium：核心条件满足，但 LifecycleGrade = B 或部分条件不理想
Low：仅部分条件满足
```

---

## 八、数据库 Schema

```sql
CREATE TABLE IF NOT EXISTS tachibana_signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trade_date TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT,
    
    signal_type TEXT NOT NULL,
    confidence TEXT,
    
    entry_zone_low REAL,
    entry_zone_high REAL,
    stop_loss REAL,
    
    lifecycle_grade TEXT,
    lifecycle_score REAL,
    span_days INTEGER,
    dist_from_20dh REAL,
    atr_normalized REAL,
    
    title TEXT,
    description TEXT,
    risk TEXT,
    suggestion TEXT,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trade_date, code)
);

CREATE INDEX idx_tachibana_trade_date ON tachibana_signals(trade_date);
CREATE INDEX idx_tachibana_signal_type ON tachibana_signals(signal_type);
CREATE INDEX idx_tachibana_confidence ON tachibana_signals(confidence DESC);
```

---

## 九、测试策略

### 9.1 单元测试

```
1. 决策规则测试
   - TestTrendProbeEntry
   - TestTrendConfirmationAdd
   - TestDistributionReduce
   - TestExitOnRhythmFailure
   - TestWaitNoAction

2. 价格区间计算测试
   - TestCalcEntryZone
   - TestCalcAddZone
   - TestCalcStopLoss

3. 置信度测试
   - TestConfidenceHigh
   - TestConfidenceMedium
   - TestConfidenceLow
```

### 9.2 集成测试

```
使用 Lifecycle 结果（151只股票）：
1. 输入：LifecycleMetrics
2. 处理：分类决策
3. 验证：
   - 所有股票都有信号类型
   - 置信度分布合理
   - 价格区间逻辑正确
   - Excel报告生成成功
```

---

## 十、Phase 2 升级路径

当完整 MALF 成型后：

```
1. 引入 MALF 前置认知过滤器
2. 实现 Position Management
   - 中心单/加码单/锁单
   - 库存管理
   - 仓位演化
3. 加入心理状态建模
4. 历史交易记录跟踪
5. 双侧库存管理
```

**迁移成本：** Phase 1 代码可复用约 50%，主要是决策规则和价格计算。

---

## 十一、重要声明

**⚠️ 这不是交易信号系统！**

- ✅ 这是"值得关注"的提示
- ✅ 这是立花式思维的参考
- ❌ 不是自动交易信号
- ❌ 不保证盈利
- ❌ 需要人工判断和确认

**风险提示：**
- 所有提示仅供参考
- 用户需自行判断和决策
- 市场有风险，投资需谨慎

---

## 十二、参考文档

- 立花义正：《你也能成为股票操作高手》
- asteria-trading-lab: 立花交易依据分类表
- MALF-立花映射总表
- 1975-1976 月度交易记录
