# TradeBuddy 项目状态

> 最后更新：2026-10-09 20:00

---

## 📊 整体进度

**当前阶段：** 第一期 - 核心引擎 + CLI + 每日复盘（2周）  
**开始日期：** 2026-10-09  
**当前状态：** Day 1-2 完成 ✅  
**总工期调整：** 5周 → **7周**（新增回测引擎）

---

## 🎯 重要更新（v2.0）

### ⭐ 新增功能规划

1. **第二期：回测引擎**（2周）
   - 历史策略验证
   - 参数优化
   - 绩效统计（夏普比率、最大回撤等）

2. **强化第一期：每日复盘**
   - 实盘交易事后分析
   - 账本恒等式校验
   - 跟踪池状态机

### 📋 新增文档

- ✅ [docs/BACKTEST.md](docs/BACKTEST.md) - 回测系统设计（完整）
- ✅ [docs/TASK.md](docs/TASK.md) - 任务清单v2.0（已更新）

---

## ✅ 已完成任务

### Day 0: 项目准备（2026-10-09 完成）
- ✅ 技术验证（Demo 1-3）
- ✅ 归档GLM文档
- ✅ 创建项目结构
- ✅ 编写核心文档（README/DESIGN/REQUIRE/TASK/BACKTEST）
- ✅ 数据库Schema设计（0001_init.sql）
- ✅ 数据模型定义（model/*.go）
- ✅ Claude Code工具链配置
- ✅ Git仓库初始化并推送到GitHub

### Day 1-2: 数据访问层 Store（2026-10-09 完成）✅

**实现文件：**
```
internal/store/
├── store.go              # 主接口 ✅
├── migrate.go            # 数据库迁移 ✅
├── instrument.go         # 股票元数据操作 ✅
├── bars.go               # K线数据操作 ✅
├── screen.go             # 初选结果操作 ✅
├── lifecycle.go          # 生命周期操作 ✅
├── tachibana.go          # 立花信号操作 ✅
├── tracker.go            # 跟踪池操作 ✅
├── transaction.go        # 成交记录操作 ✅
├── review.go             # 复盘报告操作 ✅
└── store_test.go         # 单元测试 ✅
```

**核心功能：**
- ✅ SQLite连接管理（modernc.org/sqlite）
- ✅ 数据库迁移系统（embed SQL）
- ✅ 8个数据访问模块完整实现
- ✅ 批量操作支持事务
- ✅ 时间类型和NULL值正确处理
- ✅ WAL模式优化并发性能

**测试结果：**
```bash
✅ 9个测试全部通过
⏱️  总耗时: 1.398s
📦 覆盖率: 核心CRUD操作全覆盖
```

**Git提交：**
- Commit: 50b438f - `feat(store): 实现完整的数据访问层`
- 已推送到: https://github.com/everything-is-simple/tradebuddy

---

## 🎯 下一步任务

### Day 3-4: 数据源适配（datasource）⏳ 当前任务

**目标：** 实现TDX文件读取 + **腾讯API前复权K线获取**

**待创建文件：**
```
internal/datasource/
├── tdx/
│   ├── reader.go         # 读取.day文件
│   └── reader_test.go
├── tencent/
│   ├── api.go            # 腾讯前复权K线API ⭐
│   └── api_test.go
└── sina/
    ├── api.go            # 新浪API（备选）
    └── api_test.go
```

**核心任务：**
- [ ] TDX .day文件32字节定长记录解析
- [ ] **腾讯API前复权K线获取**（重点）⭐
- [ ] 限流控制（0.08-0.1s间隔）
- [ ] 单元测试（使用sh600519.day验证）

**验收标准：**
- [ ] 能正确读取 sh600519.day（茅台数据）
- [ ] 能从腾讯API获取前复权日K
- [ ] 限流策略有效
- [ ] 数据格式与Demo 3一致

---

## 📈 开发路线图（7周）

```
第一期（2周）：核心引擎 + CLI + 每日复盘
├─ Week 1: 数据层 + Screener
│  ├─ ✅ Day 1-2: Store层
│  ├─ ⏳ Day 3-4: 数据源适配
│  └─ ⏳ Day 5-7: Screener引擎
│
└─ Week 2: Lifecycle + Tachibana + CLI + Review
   ├─ Day 8-10: Lifecycle引擎
   ├─ Day 11-12: Tachibana引擎
   └─ Day 13-14: CLI工具 + 每日复盘 ⭐

第二期（2周）：回测引擎 ⭐ 新增
├─ Week 3: 回测核心引擎
│  ├─ Day 15-16: 时间回放引擎
│  ├─ Day 17-18: 交易模拟器
│  └─ Day 19-21: 绩效统计
│
└─ Week 4: 参数优化 + CLI
   ├─ Day 22-24: 参数优化引擎
   ├─ Day 25-26: 回测CLI命令
   └─ Day 27-28: 数据库扩展 + 测试

第三期（1周）：简化GUI
└─ Week 5: Wails最小界面
   ├─ Day 29-31: 项目初始化
   ├─ Day 32-33: 三视图
   └─ Day 34-35: 手动触发任务

第四期（2周）：完整GUI + 打包
├─ Week 6: 仪表盘 + 回测可视化
└─ Week 7: 定时调度 + 打包
```

---

## 📈 里程碑进度

| 里程碑 | 预计日期 | 状态 |
|--------|----------|------|
| M1: 数据层完成 | 2026-10-11 | ✅ 已完成（提前2天） |
| M2: Screener完成 | 2026-10-14 | ⏳ 待开始 |
| M3: Lifecycle完成 | 2026-10-18 | ⏳ 待开始 |
| M4: Tachibana+Review完成 | 2026-10-23 | ⏳ 待开始 |
| M5: 第一期交付 | 2026-10-23 | ⏳ 待开始 |
| M6: 回测引擎完成 | 2026-11-07 | ⏳ 待开始 |
| M7: 第二期交付 | 2026-11-07 | ⏳ 待开始 |
| M8: 简化GUI完成 | 2026-11-14 | ⏳ 待开始 |
| M9: 完整系统交付 | 2026-11-27 | ⏳ 待开始 |

---

## 📦 交付物检查清单

### 第一期交付物（2周后）
- [x] 数据模型定义
- [x] 数据访问层
- [ ] 数据源适配
- [ ] 计算工具
- [ ] Screener引擎
- [ ] Lifecycle引擎
- [ ] Tachibana引擎
- [ ] Tracker跟踪池
- [ ] **Reviewer复盘引擎** ⭐
- [ ] Reporter报告生成
- [ ] CLI工具

### 第二期交付物（4周后）⭐ 新增
- [ ] 回测引擎
- [ ] 参数优化器
- [ ] 回测CLI命令
- [ ] 数据库扩展（3张新表）
- [ ] 回测报告Excel模板

### 第三期交付物（5周后）
- [ ] `tb-desktop.exe`
- [ ] 三视图界面

### 第四期交付物（7周后）
- [ ] 完整GUI应用
- [ ] 单文件打包
- [ ] 安装手册

---

## 🔗 快速链接

- **GitHub仓库：** https://github.com/everything-is-simple/tradebuddy
- **最新提交：** 50b438f
- **分支：** main

---

## 💡 技术亮点

1. **纯Go实现** - 无Python依赖，单文件打包
2. **SQLite嵌入式** - modernc.org/sqlite纯Go驱动
3. **前复权数据** - 确保价格连续性和回测准确性
4. **回测引擎** - 历史策略验证 + 参数优化 ⭐
5. **每日复盘** - 账本恒等式校验 + 跟踪池状态机 ⭐
6. **完整测试** - 单元测试 + Golden Test
7. **WAL模式** - 提升并发性能

---

## 📚 核心文档

| 文档 | 状态 | 说明 |
|------|-----|------|
| [README.md](README.md) | ✅ | 项目概览 |
| [CLAUDE.md](CLAUDE.md) | ✅ | Claude Code配置 |
| [docs/DESIGN.md](docs/DESIGN.md) | ✅ | 系统设计 |
| [docs/REQUIRE.md](docs/REQUIRE.md) | ✅ | 需求规格 |
| [docs/TASK.md](docs/TASK.md) | ✅ v2.0 | 任务清单（含回测）⭐ |
| [docs/BACKTEST.md](docs/BACKTEST.md) | ✅ 新增 | 回测系统设计 ⭐ |
| [docs/ADR.md](docs/ADR.md) | ✅ | 架构决策记录 |
| [docs/API.md](docs/API.md) | ✅ | API接口规范 |
| [docs/ERD.md](docs/ERD.md) | ✅ | 数据库设计 |
| [docs/TRD.md](docs/TRD.md) | ✅ | 技术参考 |

---

## 🎯 复权问题解答

**Q: 复权在哪个任务处理？**  
A: **Day 3-4** 数据源适配中实现

**Q: 选择的复权方式？**  
A: **前复权（Forward Adjustment）**
- 当前价格真实
- 历史价格等比缩小
- 腾讯API直接返回前复权数据
- 详见：[ADR-004](docs/ADR.md#ADR-004)

---

## 🔬 回测 vs 复盘

| 维度 | 历史回测（Backtest） | 每日复盘（Review） |
|------|---------------------|-------------------|
| **目的** | 策略验证 | 实盘评估 |
| **时间** | 历史任意区间 | 每日15:30 |
| **数据** | 历史K线 | 当日成交 |
| **实现** | 第二期（2周后）⭐ | 第一期Day 13-14 ⭐ |
| **核心** | 绩效统计、参数优化 | 账本校验、状态更新 |

**两者都需要，相互补充！** ✅

---

**准备好了吗？** 回复"继续"或"开始Day 3-4"开始数据源适配实现！ 🚀
