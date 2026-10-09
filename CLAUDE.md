# TradeBuddy 项目配置

> 个人量化交易系统 - 从选股到交易到复盘的完整闭环

---

## 技术栈

- **语言：** Go 1.22+
- **数据库：** SQLite 3 (modernc.org/sqlite)
- **数据源：** 通达信本地数据 + 腾讯/新浪API
- **GUI：** Wails v2（第二期）

---

## 项目结构

```
tradebuddy/
├── cmd/
│   ├── verify/          # 技术验证Demo
│   ├── tb-cli/          # 命令行工具
│   └── tb-desktop/      # Wails桌面端
├── internal/
│   ├── store/           # SQLite数据访问层
│   ├── datasource/      # 数据源适配（TDX/腾讯/新浪）
│   ├── calc/            # 计算工具（MA/ATR/MACD/摆动点）
│   ├── screener/        # 19:00强势股初选引擎
│   ├── lifecycle/       # 19:30生命周期分析引擎
│   ├── tachibana/       # 20:00立花交易计划引擎
│   ├── tracker/         # 跟踪池状态机
│   └── reporter/        # 报告生成器
├── data/                # 数据库和配置
└── reports/             # 每日Excel报告
```

---

## 编码规范

### Go代码风格

- 使用 `gofmt` 自动格式化
- 错误处理：优先返回 `error`，不使用 `panic`
- 命名：
  - 包名：小写单词，不使用下划线
  - 导出标识符：首字母大写
  - 私有标识符：首字母小写
  - 常量：驼峰式（非全大写）
- 注释：所有导出的函数和类型必须有文档注释

### 函数签名模式

```go
// 查询操作：context在前，条件在后
func (s *Store) GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error)

// 写操作：context在前，数据在后
func (s *Store) InsertDailyBars(ctx context.Context, bars []*model.DailyBar) error
```

### 关键算法必须附原著引用

```go
// CalculateTierPrices 计算立花分批买入梯价
// 
// 梯①②③间距 = 0.75×ATR / 1.5×ATR / 2.25×ATR
// 理论来源：立花义正《你也可以成为股票操作高手》p250-263
// 验证依据：1975-76账本159笔，梯距中位0.37ATR（档案：G0资产/立花玉帳数字化）
func CalculateTierPrices(ma20, atr14 float64) (tier1, tier2, tier3 float64) {
    tier1 = ma20 - 0.75*atr14
    tier2 = ma20 - 1.50*atr14
    tier3 = ma20 - 2.25*atr14
    return
}
```

---

## 测试要求

### 单元测试

- 文件命名：`*_test.go`
- 覆盖率目标：核心计算模块 ≥ 60%
- 表驱动测试优先：
  ```go
  func TestMA(t *testing.T) {
      tests := []struct {
          name   string
          prices []float64
          period int
          want   float64
      }{
          {"5日均线", []float64{10, 11, 12, 13, 14}, 5, 12.0},
          // ...
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              got := MA(tt.prices, tt.period)
              if math.Abs(got-tt.want) > 0.0001 {
                  t.Errorf("MA() = %v, want %v", got, tt.want)
              }
          })
      }
  }
  ```

### Golden Test（必需）

用已知正确结果验证新实现：

```go
// TestScreener_20261008 验证10-08初选结果
func TestScreener_20261008(t *testing.T) {
    result := screener.Run(ctx, "2026-10-08")
    
    // 已知产出27只股票
    assert.Equal(t, 27, result.Count)
    
    // 验证部分代码
    codes := extractCodes(result.Stocks)
    assert.Contains(t, codes, "605133") // 世名科技
    assert.Contains(t, codes, "600825") // 新华传媒
}
```

---

## 数据源约束

### 通达信数据路径

- **根目录：** `H:\new_tdx64\vipdoc`
- **上海：** `sh/lday/sh600519.day`
- **深圳：** `sz/lday/sz000001.day`
- **格式：** 32字节定长记录（小端）

### 腾讯API限流

- 请求间隔：0.08-0.1秒
- 并发数：4-6线程
- 501限流时：指数退避重试

### 炒股手数据（第二期）

- **路径：** `H:\2025炒股手训练软件\stockdata\min5`
- **用途：** 历史复盘训练、模拟盘测试

---

## 常用命令

```bash
# 开发
go run cmd/tb-cli/main.go --help
go test ./...
go test -v -run TestScreener_20261008

# 构建
go build -o tb-cli.exe cmd/tb-cli/main.go
go build -ldflags="-s -w" -o tb-cli.exe cmd/tb-cli/main.go  # 减小体积

# 验证
go run cmd/verify/demo1_sqlite.go
go run cmd/verify/demo2_excel.go
go run cmd/verify/demo3_tdx.go

# 数据库迁移
tb-cli migrate up
tb-cli migrate down

# 代码检查
gofmt -s -w .
go vet ./...
staticcheck ./...
```

---

## 权限与安全

### 敏感文件保护

以下文件**禁止提交**到Git：

- `data/*.db`（数据库文件）
- `reports/**/*.xlsx`（每日报告）
- `*.log`（日志文件）

### 配置文件管理

- `data/config.yaml` - 提交模板到Git
- `data/config.local.yaml` - 本地配置，不提交
- 优先级：local > yaml

---

## 核心约束

### 理论可验证性

所有统计概率必须附：
- 样本量
- Wilson 95%置信区间
- 样本外验证误差

示例：
```go
type ProbResult struct {
    Prob       float64 // 8周存活概率
    CILow      float64 // 置信区间下限
    CIHigh     float64 // 置信区间上限
    SampleSize int     // 样本量
}
```

### 账本恒等式（每日校验）

```
权益 = 现金 + 持仓市值
今日权益 = 昨日权益 + 今日损益 - 手续费
```

### 风控硬限制

```go
const (
    MaxRiskPerStock     = 0.01   // 单股风险1%
    MaxNotionalPerStock = 200000 // 单股市值上限20万
    MaxPositions        = 120    // 跟踪池上限120条
    ExpireDays          = 40     // 挂单过期40天
)
```

---

## 开发工作流

### 功能开发

1. 先在 `docs/TASK.md` 找到对应任务
2. 实现功能 + 单元测试
3. Golden test验证（如适用）
4. 更新TASK.md状态

### Commit规范

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
test(lifecycle): 添加摆动点识别golden test
```

### 分支策略

- `main` - 稳定版本
- `dev` - 开发分支
- `feat/xxx` - 功能分支

---

## 性能目标

| 任务 | 时间限制 |
|------|----------|
| 19:00初选 | < 3分钟（5000只股票） |
| 19:30生命周期 | < 2分钟（50只股票） |
| 20:00交易计划 | < 1分钟 |
| 15:30复盘 | < 30秒 |
| CLI成交录入 | < 1秒 |

---

## 参考文档

- **系统设计：** [docs/DESIGN.md](docs/DESIGN.md)
- **需求规格：** [docs/REQUIRE.md](docs/REQUIRE.md)
- **任务清单：** [docs/TASK.md](docs/TASK.md)
- **验证指南：** [VERIFY.md](VERIFY.md)

---

## 理论参考书目

1. 欧奈尔《笑傲股市》
2. 达瓦斯《我如何在股市赚了200万》
3. Minervini《股票魔法师》
4. 斯波朗迪《专业投机原理》
5. 立花义正《你也可以成为股票操作高手》
6. 布伦特·奔富《交易圣经》

---

**注意：** 本系统输出仅作研究参考，不构成投资建议。
