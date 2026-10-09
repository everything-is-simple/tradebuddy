# TradeBuddy - 个人量化交易系统

> 从选股到交易到复盘的完整闭环系统  
> 纯Go实现，单文件打包，跨平台运行

---

## 🎯 项目定位

**核心目标：** 构建一个完整闭环的个人量化交易系统，包含选股、分析、交易、复盘全流程。

**理论基础：**
- **选股层** - 欧奈尔CANSLIM + 达瓦斯箱体 + Minervini趋势模板
- **分析层** - 斯波朗迪生命周期统计 + MALF波段标尺
- **交易层** - 立花义正うねり取り + 林辉太郎分批建仓
- **评价层** - 布伦特·奔富R倍数系统 + 账本恒等式校验

---

## 🚀 快速开始

### 环境要求

- **Go 1.22+** - [下载地址](https://go.dev/dl/)
- **Windows 10/11** (主要开发平台)
- **通达信数据** - `H:\new_tdx64\vipdoc`

### 编译

```bash
# 克隆项目
cd I:\tradebuddy

# 编译主程序
go build -o tradebuddy.exe ./cmd/tradebuddy

# 编译集成测试
go build -o integration-test.exe ./cmd/integration-test
```

### 基本使用

```bash
# 运行完整流程（初选 → 生命周期 → 立花提示）
./tradebuddy.exe run --date 2026-06-15

# 只运行初选
./tradebuddy.exe screen --date 2026-06-15

# 只运行生命周期分析
./tradebuddy.exe lifecycle --date 2026-06-15

# 只运行立花提示
./tradebuddy.exe tachibana --date 2026-06-15

# 查看历史报告
./tradebuddy.exe report --date 2026-06-15

# 查看帮助
./tradebuddy.exe --help
```

### 输出文件

运行后会生成：
- **数据库：** `data/tradebuddy.db`
- **报告目录：** `reports/`
  - `screen_result_YYYYMMDD.xlsx` - 初选结果
  - `lifecycle_result_YYYYMMDD.xlsx` - 生命周期分析
  - `tachibana_result_YYYYMMDD.xlsx` - 立花提示信号

### 运行集成测试

```bash
# 端到端测试（使用 2026-06-15 数据）
./integration-test.exe
```

---

## 📊 系统工作流

```
┌─────────────────────────────────────────┐
│  每日自动化流程                          │
├─────────────────────────────────────────┤
│  19:00  强势股初选                       │
│    ↓    涨幅>6% + 创20日新高 + 趋势过滤  │
│                                          │
│  19:30  生命周期分析                     │
│    ↓    九阶段分类 + 存活概率统计        │
│                                          │
│  20:00  立花交易计划                     │
│    ↓    回档布网 + 梯价 + 止损 + 跟踪池  │
│                                          │
│  20:05  人工执行                         │
│    ↓    看Excel → 券商APP挂单           │
│    ↓    tb-cli 录入成交信息             │
│                                          │
│  15:30  每日复盘                         │
│         跟踪池更新 → 复盘报告            │
└─────────────────────────────────────────┘
```

---

## 📂 项目结构

```
tradebuddy/
├── cmd/
│   ├── verify/          # 技术验证Demo
│   ├── tb-cli/          # 命令行工具
│   └── tb-desktop/      # Wails桌面端
│
├── internal/
│   ├── store/           # SQLite数据访问层
│   │   └── migrations/  # 数据库迁移脚本
│   ├── model/           # 数据模型
│   ├── datasource/      # 数据源适配（TDX/腾讯/新浪）
│   ├── calc/            # 计算工具（MA/ATR/MACD/摆动点）
│   ├── screener/        # 19:00 强势股初选引擎
│   ├── lifecycle/       # 19:30 生命周期分析引擎
│   ├── tachibana/       # 20:00 立花交易计划引擎
│   ├── tracker/         # 跟踪池状态机
│   └── reporter/        # 报告生成器
│
├── data/
│   ├── tradebuddy.db    # SQLite数据库（自动创建）
│   └── config.yaml      # 配置文件
│
├── reports/             # 每日Excel报告输出
│   ├── screen/
│   ├── lifecycle/
│   ├── tachibana/
│   └── review/
│
├── docs/
│   ├── DESIGN.md        # 系统设计文档 ⭐
│   ├── REQUIRE.md       # 需求规格说明 ⭐
│   └── TASK.md          # 开发任务清单 ⭐
│
├── go.mod
├── go.sum
├── README.md            # 本文件
├── AGENTS.md            # Agent工作规范
└── VERIFY.md            # 技术验证指南
```

---

## 📖 核心文档

| 文档 | 说明 | 状态 |
|------|------|------|
| [README.md](README.md) | 项目概览（本文件） | ✓ |
| [CLAUDE.md](CLAUDE.md) | 项目配置和编码规范 | ✓ |
| [VERIFY.md](VERIFY.md) | 技术验证指南 | ✓ |
| [DESIGN.md](docs/DESIGN.md) | 系统设计文档 | ✓ |
| [REQUIRE.md](docs/REQUIRE.md) | 需求规格说明 | ✓ |
| [TASK.md](docs/TASK.md) | 开发任务清单 | ✓ |
| [DAY15-SUMMARY.md](docs/DAY15-SUMMARY.md) | Day 15 完成总结 | ✓ |

---

## 🛠️ 技术栈

- **语言：** Go 1.22+
- **数据库：** SQLite 3 (modernc.org/sqlite)
- **GUI：** Wails v2（第二期）
- **Excel：** excelize/v2
- **数据源：** TDX本地 + 腾讯/新浪API

**为什么选Go？**
```
✓ 单文件打包（~15MB，无依赖）
✓ 跨平台（Windows/Linux/macOS）
✓ 启动快（<100ms）
✓ 并发性能强（goroutine）
✓ 交付简单（2个文件：exe + db）
```

---

## 📅 开发计划（三期）

### 第一期：核心引擎 + CLI（2周）✅

**目标：** 完成选股→分析→交易计划的自动化闭环

- [x] 技术验证（Demo 1-3）
- [x] 数据库Schema + 数据访问层
- [x] Screener引擎（19:00）
- [x] Lifecycle引擎（19:30）
- [x] Tachibana引擎（20:00）
- [x] CLI工具（完整流程）
- [x] Reporter报告生成
- [x] 集成测试

**交付物：** `tradebuddy.exe` + `tradebuddy.db` + Excel报告

**状态：** ✅ 已完成 (Day 15, 2026-10-10)

---

### 第二期：简化GUI（1周）

**目标：** 可视化查看清单、生命周期、交易计划

- [ ] Wails项目初始化
- [ ] 三视图（清单/生命周期/计划）
- [ ] 手动触发任务按钮

**交付物：** `tb-desktop.exe`

---

### 第三期：完整GUI + 打包（2周）

**目标：** 单文件桌面应用，内置定时调度

- [ ] 仪表盘（任务状态/阶段分布）
- [ ] 跟踪池管理
- [ ] 复盘报告查看
- [ ] 内置定时调度
- [ ] 单文件打包

**最终交付：** 
- Windows: `tradebuddy.exe` (~15MB)
- Linux/macOS: `tradebuddy`
- + `tradebuddy.db`

---

## 📌 当前状态

**✅ Phase 1 完成 - Day 15 (2026-10-10)**

```
[✓] 技术验证完成
[✓] 项目结构初始化
[✓] 数据库Schema + Store层
[✓] 数据源适配（TDX + 腾讯API）
[✓] 计算工具（MA/ATR/MACD）
[✓] Screener 强势股初选引擎
[✓] Lifecycle 生命周期分析引擎
[✓] Tachibana 立花提示引擎
[✓] Reporter 报告生成
[✓] CLI 命令行工具
[✓] 集成测试框架
[ ] 第二期：简化GUI
```

**可执行文件：**
- `tradebuddy.exe` - 主CLI工具 (22MB)
- `integration-test.exe` - 集成测试 (22MB)

详见：[Day 15 总结](docs/DAY15-SUMMARY.md)

---

## 🔗 外部资源

**数据源：**
- 通达信本地数据：`H:\new_tdx64\vipdoc`
- 炒股手训练软件：`H:\2025炒股手训练软件`

**理论参考：**
- 欧奈尔《笑傲股市》
- 达瓦斯《我如何在股市赚了200万》
- Minervini《股票魔法师》
- 斯波朗迪《专业投机原理》
- 立花义正《你也可以成为股票操作高手》
- 林辉太郎相关著作
- 布伦特·奔富《交易圣经》

**归档文档：**
- GLM生成的文档：`I:\tradebuddy-archived\glm-docs-20261009\`

---

## 📞 项目信息

- **项目路径：** `I:\tradebuddy\`
- **归档路径：** `I:\tradebuddy-archived\`
- **创建日期：** 2026-10-09
- **开发模式：** 纯Go，单文件打包，跨平台

---

**让我们构建完整闭环的个人量化交易系统！** 🚀
