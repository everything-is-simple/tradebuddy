# TradeBuddy 使用手册

## 目录

- [快速开始](#快速开始)
- [命令参考](#命令参考)
- [工作流程](#工作流程)
- [输出说明](#输出说明)
- [常见问题](#常见问题)

---

## 快速开始

### 第一次使用

```bash
# 1. 确保通达信数据目录存在
# 默认路径: H:\new_tdx64\vipdoc

# 2. 编译程序
go build -o tradebuddy.exe ./cmd/tradebuddy

# 3. 运行完整流程（使用今天日期）
./tradebuddy.exe run

# 4. 查看生成的报告
# 报告目录: reports/
```

### 典型使用场景

**场景1: 每日19:00选股**
```bash
# 运行完整流程，生成当日选股和分析结果
./tradebuddy.exe run --date 2026-10-10
```

**场景2: 历史数据回测**
```bash
# 分析历史某天的数据
./tradebuddy.exe run --date 2026-06-15
```

**场景3: 查看历史报告**
```bash
# 查看某天的所有结果
./tradebuddy.exe report --date 2026-06-15

# 只查看生命周期分析
./tradebuddy.exe report --date 2026-06-15 --type lifecycle
```

---

## 命令参考

### 主命令

```bash
tradebuddy <command> [options]
```

### 子命令

#### 1. `run` - 运行完整流程

**功能**: 依次执行 Screener → Lifecycle → Tachibana，生成3份报告

**用法**:
```bash
tradebuddy run [options]
```

**选项**:
- `--date YYYY-MM-DD`: 交易日期（默认今天）
- `--db PATH`: 数据库路径（默认 `data/tradebuddy.db`）
- `--tdx PATH`: 通达信数据根目录（默认 `H:\new_tdx64\vipdoc`）
- `--output PATH`: 报告输出目录（默认 `reports/`）
- `--min-gain FLOAT`: 最小涨幅百分比（默认 6.0）
- `--lookback INT`: 回溯天数（默认 20）
- `--dd52 FLOAT`: 距52周高点最大距离（默认 25.0）

**示例**:
```bash
# 使用默认参数
./tradebuddy.exe run

# 指定日期
./tradebuddy.exe run --date 2026-06-15

# 自定义筛选条件（涨幅8%）
./tradebuddy.exe run --date 2026-06-15 --min-gain 8.0

# 使用自定义TDX路径
./tradebuddy.exe run --tdx "D:\TDX\vipdoc"
```

#### 2. `screen` - 强势股初选

**功能**: 筛选符合条件的强势股

**用法**:
```bash
tradebuddy screen [options]
```

**筛选条件**:
- 当日涨幅 ≥ 6%（可调整）
- 距52周高点 ≤ 25%（可调整）
- 有足够的历史数据

**输出**:
- 控制台: 进度显示、结果统计
- Excel: `reports/screen_result_YYYYMMDD.xlsx`
- 数据库: `screen_results` 表

**示例**:
```bash
# 默认参数
./tradebuddy.exe screen --date 2026-06-15

# 更严格的筛选（涨幅8%，距52周高点15%）
./tradebuddy.exe screen --date 2026-06-15 --min-gain 8.0 --dd52 15.0
```

#### 3. `lifecycle` - 生命周期分析

**功能**: 分析股票的波段特征和强度

**前置条件**: 必须先运行 `screen` 命令

**用法**:
```bash
tradebuddy lifecycle [options]
```

**分析内容**:
- 波段持续时间
- 价格幅度
- ATR标准化
- 市场排名
- 综合评级（A/B/C/D）

**输出**:
- 控制台: 评级分布、前5名
- Excel: `reports/lifecycle_result_YYYYMMDD.xlsx`
- 数据库: `lifecycle_metrics` 表

**示例**:
```bash
# 默认参数
./tradebuddy.exe lifecycle --date 2026-06-15

# 更严格的样本要求
./tradebuddy.exe lifecycle --date 2026-06-15 --min-sample 50
```

#### 4. `tachibana` - 立花提示分析

**功能**: 生成交易建议信号

**前置条件**: 必须先运行 `lifecycle` 命令

**用法**:
```bash
tradebuddy tachibana [options]
```

**信号类型**:
- **试探建仓** (trend_probe_entry): A/B级股票，值得关注
- **同向加码** (trend_confirmation_add): 趋势确认，可考虑加仓
- **分批减仓** (distribution_reduce): 进入派发阶段
- **节奏失败** (exit_on_rhythm_failure): 建议退出
- **等待观望** (wait_no_action): 暂不操作

**输出**:
- 控制台: 信号类型分布、值得关注的信号
- Excel: `reports/tachibana_result_YYYYMMDD.xlsx`（按信号类型分Sheet）
- 数据库: `tachibana_signals` 表

**示例**:
```bash
./tradebuddy.exe tachibana --date 2026-06-15
```

#### 5. `report` - 查看历史报告

**功能**: 查看数据库中的历史分析结果

**用法**:
```bash
tradebuddy report [options]
```

**选项**:
- `--date YYYY-MM-DD`: 查询日期（必需）
- `--type TYPE`: 报告类型（screen/lifecycle/tachibana/all，默认 all）

**示例**:
```bash
# 查看所有报告
./tradebuddy.exe report --date 2026-06-15

# 只查看生命周期分析
./tradebuddy.exe report --date 2026-06-15 --type lifecycle

# 只查看立花提示
./tradebuddy.exe report --date 2026-06-15 --type tachibana
```

#### 6. `version` - 版本信息

```bash
./tradebuddy.exe version
# 输出: tradebuddy version 0.1.0
```

#### 7. `help` - 帮助信息

```bash
# 主帮助
./tradebuddy.exe help

# 子命令帮助
./tradebuddy.exe run --help
./tradebuddy.exe screen --help
```

---

## 工作流程

### 完整流程（推荐）

```
1. 运行 run 命令
   ↓
2. Screener 初选
   - 从 TDX 读取所有股票数据
   - 筛选符合条件的股票
   - 保存到数据库和Excel
   ↓
3. Lifecycle 分析
   - 读取初选结果
   - 分析每只股票的波段特征
   - 计算综合评分和评级
   - 保存到数据库和Excel
   ↓
4. Tachibana 提示
   - 读取生命周期结果
   - 根据评级和指标生成信号
   - 分类保存（按信号类型）
   - 保存到数据库和Excel
   ↓
5. 完成
   - 3份Excel报告
   - 数据库记录完整
```

### 分步执行

如果需要单独执行某个步骤：

```bash
# 步骤1: 初选
./tradebuddy.exe screen --date 2026-06-15

# 步骤2: 生命周期分析（依赖步骤1）
./tradebuddy.exe lifecycle --date 2026-06-15

# 步骤3: 立花提示（依赖步骤2）
./tradebuddy.exe tachibana --date 2026-06-15
```

---

## 输出说明

### 控制台输出

#### 进度显示
```
进度: 5000/11722 (42.7%) - 耗时: 2m15s
进度: 6000/11722 (51.2%) - 耗时: 2m40s
进度: 11722/11722 (100.0%) - 耗时: 5m30s
```

#### 结果摘要
```
✓ Screener 完成: 151 只股票, 耗时: 3m25s
✓ Lifecycle 完成: 151 只股票, 耗时: 1m45s
✓ Tachibana 完成: 151 个信号, 耗时: 0m35s
✓ 总耗时: 5m45s
```

### Excel 报告

#### 1. Screen Result (`screen_result_YYYYMMDD.xlsx`)

**内容**:
- 股票代码、名称
- 当日涨幅、收盘价、最高价
- 成交量、成交额、换手率
- 距52周高点距离
- 周均线、月均线

**排序**: 按涨幅降序

#### 2. Lifecycle Result (`lifecycle_result_YYYYMMDD.xlsx`)

**内容**:
- 股票代码、名称
- 评级（A/B/C/D）
- 综合评分
- 波段持续天数
- 价格幅度%
- ATR标准化
- 距20日高点、距52周高点
- 持续排名、幅度排名、ATR排名

**排序**: 按综合评分降序

**评级说明**:
- **A级**: 综合评分 ≥ 75，强势股
- **B级**: 综合评分 60-75，中等
- **C级**: 综合评分 40-60，一般
- **D级**: 综合评分 < 40，弱势

#### 3. Tachibana Result (`tachibana_result_YYYYMMDD.xlsx`)

**Sheet结构**: 按信号类型分Sheet
- Sheet1: 试探建仓
- Sheet2: 同向加码
- Sheet3: 分批减仓
- Sheet4: 节奏失败
- Sheet5: 等待观望
- Sheet6: 全部信号

**每个Sheet内容**:
- 股票代码、名称
- 信号类型、置信度
- 当前价格
- 入场区间（低-高）
- 止损价格
- 生命周期评级、评分
- 波段天数、距20日高点、ATR标准化
- 提示标题、描述、风险、建议

### 数据库

**位置**: `data/tradebuddy.db`

**主要表**:
- `screen_results`: 初选结果
- `lifecycle_metrics`: 生命周期指标
- `tachibana_signals`: 交易信号
- `schema_migrations`: 数据库版本

**查询示例**:
```sql
-- 查看某天的初选结果
SELECT * FROM screen_results 
WHERE trade_date = '2026-06-15' 
ORDER BY pct_change DESC;

-- 查看A级股票
SELECT * FROM lifecycle_metrics 
WHERE trade_date = '2026-06-15' AND grade = 'A'
ORDER BY lifecycle_score DESC;

-- 查看试探建仓信号
SELECT * FROM tachibana_signals 
WHERE trade_date = '2026-06-15' AND signal_type = 'trend_probe_entry';
```

---

## 常见问题

### Q1: 提示"打开数据库失败"

**原因**: 数据库目录不存在或权限不足

**解决**:
```bash
# 手动创建目录
mkdir data

# 或使用完整路径
./tradebuddy.exe run --db "I:/tradebuddy/data/tradebuddy.db"
```

### Q2: 提示"列出股票失败"

**原因**: TDX数据路径不正确

**解决**:
```bash
# 检查路径是否存在
ls "H:\new_tdx64\vipdoc\sh\lday"
ls "H:\new_tdx64\vipdoc\sz\lday"

# 使用正确的路径
./tradebuddy.exe run --tdx "D:\TDX\vipdoc"
```

### Q3: 筛选结果为0

**原因**: 当日没有符合条件的股票，或日期数据不存在

**解决**:
```bash
# 降低筛选条件
./tradebuddy.exe screen --date 2026-06-15 --min-gain 3.0 --dd52 50.0

# 使用已知有数据的日期
./tradebuddy.exe screen --date 2026-06-15
```

### Q4: 运行很慢

**原因**: 需要处理11000+只股票

**正常耗时**:
- Screener: 3-6分钟（全市场扫描）
- Lifecycle: 1-3分钟（50-200只股票）
- Tachibana: 30秒-1分钟

**优化**:
- 使用 SSD 存储 TDX 数据
- 确保数据库在本地磁盘

### Q5: Excel 打不开

**原因**: 文件正在被其他程序占用

**解决**:
- 关闭Excel后重新运行
- 或删除旧报告后重新生成

### Q6: 提示"未找到初选结果"

**原因**: 没有先运行 `screen` 命令

**解决**:
```bash
# 先运行初选
./tradebuddy.exe screen --date 2026-06-15

# 再运行生命周期
./tradebuddy.exe lifecycle --date 2026-06-15

# 或直接运行完整流程
./tradebuddy.exe run --date 2026-06-15
```

### Q7: 想修改默认参数

**方法1**: 命令行参数（推荐）
```bash
./tradebuddy.exe run --min-gain 8.0 --dd52 20.0
```

**方法2**: 修改源码
编辑 `cmd/tradebuddy/cmd_screen.go`:
```go
fs.Float64Var(&minGain, "min-gain", 8.0, "最小涨幅百分比")  // 改为8.0
```

---

## 性能指标

### 预期性能

| 任务 | 数据量 | 预期时间 |
|------|--------|----------|
| Screener | 11722只股票 | 3-6分钟 |
| Lifecycle | 50-200只股票 | 1-3分钟 |
| Tachibana | 50-200只股票 | 30秒-1分钟 |
| 完整流程 | 全流程 | 5-10分钟 |

### 硬件要求

- **CPU**: 双核及以上
- **内存**: 4GB及以上
- **磁盘**: 100MB（程序） + 500MB（数据库）
- **磁盘类型**: SSD 推荐（读取TDX数据更快）

---

## 进阶使用

### 批量处理历史数据

```bash
# 批量处理多个日期
for date in 2026-06-01 2026-06-02 2026-06-03; do
    ./tradebuddy.exe run --date $date
done
```

### 定时任务（Windows）

创建 `run_daily.bat`:
```batch
@echo off
cd I:\tradebuddy
tradebuddy.exe run
```

添加到任务计划程序，每天19:00执行。

### 自定义输出目录

```bash
# 按月份组织报告
./tradebuddy.exe run --date 2026-06-15 --output "reports/2026-06"
./tradebuddy.exe run --date 2026-07-15 --output "reports/2026-07"
```

---

**需要更多帮助？** 查看 [项目文档](../README.md) 或提交 Issue。
