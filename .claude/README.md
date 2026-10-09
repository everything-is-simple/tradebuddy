# TradeBuddy Claude Code 工具链配置指南

> 完整的开发辅助工具配置

---

## 📁 配置文件结构

```
I:\tradebuddy/
├── CLAUDE.md                    # 项目配置（自动加载）
├── .claude/
│   ├── settings.json            # Claude Code设置
│   ├── commands/                # 自定义命令
│   │   ├── review.md            # /review - 代码审查
│   │   ├── test.md              # /test - 生成测试
│   │   ├── doc.md               # /doc - 生成文档
│   │   ├── commit.md            # /commit - Git提交
│   │   └── golden.md            # /golden - Golden Test
│   ├── skills/                  # Skills目录（待安装）
│   └── hooks/                   # Git Hooks
│       └── pre-commit           # 提交前检查
└── .gitignore                   # Git忽略规则
```

---

## 🚀 快速开始

### 1. 验证配置

```bash
# 检查Claude Code版本
claude --version

# 查看可用命令
claude /help

# 测试项目配置加载
cd I:\tradebuddy
claude -p "读取CLAUDE.md，总结项目信息"
```

### 2. 使用自定义命令

```bash
# 代码审查
/review internal/store/store.go

# 生成测试
/test internal/calc/ma.go

# 生成文档
/doc CalculateTierPrices

# Git提交
/commit "实现数据访问层"

# 运行Golden Test
/golden
```

---

## 🛠️ 工具说明

### 1. CLAUDE.md（项目配置）

**作用：** 每次会话自动加载，让Claude了解项目上下文

**内容包括：**
- 技术栈和项目结构
- 编码规范和命名约定
- 测试要求和性能目标
- 数据源约束和安全规则
- 常用命令和工作流

**更新时机：**
- 技术栈变化
- 新增编码规范
- Claude频繁违反规范时

---

### 2. settings.json（权限和模型配置）

**当前配置：**
- **模型：** claude-sonnet-5-5（主力）
- **权限模式：** acceptEdits（读写自动，命令需确认）
- **安全保护：** 禁止读取/写入 .env 文件，禁止删除数据库

**权限模式说明：**

| 模式 | 读文件 | 写文件 | 执行命令 | 适用场景 |
|------|--------|--------|----------|----------|
| default | 自动 | 询问 | 询问 | 初次使用 |
| **acceptEdits** | 自动 | 自动 | 询问 | 日常开发（推荐）|
| plan | 自动 | 拒绝 | 拒绝 | 只读分析 |
| dontAsk | 自动 | 自动 | 自动 | 高度自动化 |
| bypassPermissions | 跳过所有检查 | | 仅容器内 |

**切换权限模式：**
```bash
# 临时切换
claude -p "切换到plan模式" /permissions

# 永久切换：编辑 .claude/settings.json
```

---

### 3. 自定义命令（Commands）

#### `/review` - Go代码审查
**检查项：**
- 代码规范（gofmt、命名、注释）
- 错误处理（error传播、资源释放）
- 性能考虑（内存分配、连接管理）
- 安全性（SQL注入、输入验证）
- 测试覆盖
- 项目特定（算法引用、统计概率）

**使用示例：**
```bash
/review internal/store/bars.go
/review internal/screener/
```

#### `/test` - 生成单元测试
**生成内容：**
- 表驱动测试（Table-Driven Tests）
- 正常路径 + 边界条件 + 错误处理
- Mock处理建议

**使用示例：**
```bash
/test internal/calc/ma.go
```

#### `/doc` - 生成文档注释
**生成规范：**
- 函数注释（功能、参数、返回值）
- 包注释
- 类型注释
- 特殊：关键算法附理论来源

**使用示例：**
```bash
/doc CalculateTierPrices
/doc internal/store/
```

#### `/commit` - Git提交工作流
**自动化步骤：**
1. 检查git status
2. 生成规范的commit message
3. 暂存文件
4. 提交
5. 询问是否推送

**安全检查：**
- 拒绝提交 .env 文件
- 提醒数据库文件
- 提醒大文件 >5MB
- 提醒在主分支上操作

**使用示例：**
```bash
/commit
/commit "实现Screener引擎"
```

#### `/golden` - Golden Test验证
**数据集：**
1. 2026-10-08初选（27只）
2. 2026-10-08生命周期
3. 2026-10-08交易计划
4. 1975-1976立花账本（159笔）

**使用示例：**
```bash
/golden
```

---

### 4. Hooks（Git钩子）

#### pre-commit Hook
**检查项：**
1. Go格式检查（gofmt）
2. Go vet静态分析
3. 运行测试（短测试模式）
4. 敏感文件检查

**激活方式：**
```bash
# 复制到Git hooks目录
cp .claude/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# 测试
git commit -m "test"
```

**跳过检查（紧急情况）：**
```bash
git commit --no-verify -m "emergency fix"
```

---

### 5. Skills（待安装）

**推荐Skills（来自 laolaoshiren/claude-code-skills-zh）：**

| Skill | 功能 | 优先级 |
|-------|------|--------|
| **zh-code-reviewer** | 中文代码审查 | 高 |
| **test-generator** | 生成单元测试 | 高 |
| **refactor-advisor** | 识别代码坏味道 | 中 |
| **perf-profiler** | 性能分析 | 中 |
| **error-translator** | 错误翻译 | 低 |
| **log-analyzer** | 日志分析 | 低 |

**安装方式：**
```bash
# 克隆仓库
git clone https://github.com/laolaoshiren/claude-code-skills-zh.git /tmp/skills

# 安装单个Skill
cp -r /tmp/skills/skills/zh-code-reviewer ~/.claude/skills/

# 或安装到项目（推荐）
cp -r /tmp/skills/skills/zh-code-reviewer .claude/skills/

# 使用
/zh-code-reviewer internal/store/store.go
```

---

## 📊 模型选择策略

| 模型 | Token成本 | 速度 | 适用场景 |
|-----|-----------|------|---------|
| **Haiku** | 最低 | 最快 | commit message、文档注释、简单格式化 |
| **Sonnet** | 中等 | 快 | 日常编码、代码审查、测试、Bug修复（80%任务）|
| **Opus** | 最高 | 慢 | 架构设计、复杂重构、算法实现 |

**当前配置：**
- 默认：Sonnet 5.5
- `/doc` 和 `/commit` 命令：Haiku 5.5（已在frontmatter指定）
- `/review` 和 `/test` 命令：Sonnet 5.5

**切换模型：**
```bash
# 临时切换
/model claude-opus-5-5

# 查看当前模型
/model
```

---

## 🔐 安全配置

### Deny规则（settings.json）

**禁止读取：**
- `**/*.env` - 环境变量文件
- `**/*.env.*` - 环境变量变体
- `data/*.db` - 数据库文件
- `**/*.log` - 日志文件

**禁止写入：**
- `**/*.env` - 防止泄露凭证
- `**/*.env.*`

**禁止执行：**
- `rm -rf /` - 危险删除命令
- `rm -rf *` - 危险删除命令
- `format C:` - Windows格式化命令
- `> .env` - 覆写环境变量文件

### .gitignore规则

**已忽略：**
- 可执行文件（*.exe, *.dll）
- 数据库文件（*.db）
- 生成的报告（reports/**/*.xlsx）
- 日志文件（*.log）
- IDE配置（.idea/, .vscode/）
- 依赖目录（vendor/, node_modules/）

---

## 📝 工作流示例

### 场景1：实现新功能

```bash
# 1. 创建功能分支
git checkout -b feat/screener-engine

# 2. 开始开发（自动加载CLAUDE.md）
claude

# 3. 实现功能
> 根据 docs/TASK.md Day 5-7任务，实现Screener引擎

# 4. 代码审查
/review internal/screener/screener.go

# 5. 生成测试
/test internal/screener/screener.go

# 6. 运行Golden Test
/golden

# 7. 提交代码
/commit "实现Screener引擎，通过Golden Test"

# 8. 推送并创建PR
git push origin feat/screener-engine
gh pr create --title "feat(screener): 实现三条件过滤器" --body "..."
```

### 场景2：修复Bug

```bash
# 1. 切换到plan模式（只读分析）
> /permissions plan

# 2. 分析问题
> 分析为什么 bars.go 中日K线会重复插入

# 3. 切换回acceptEdits模式
> /permissions acceptEdits

# 4. 修复并测试
> 修复bars.go中的重复插入问题，并添加单元测试

# 5. 验证修复
go test -v ./internal/store -run TestInsertDailyBars

# 6. 提交
/commit "修复日K线重复插入问题"
```

### 场景3：代码重构

```bash
# 1. 运行Golden Test建立基线
/golden

# 2. 执行重构
> 重构 lifecycle.go，将摆动点识别逻辑提取为独立函数

# 3. 代码审查
/review internal/lifecycle/

# 4. 再次运行Golden Test验证
/golden

# 5. 提交
/commit "重构lifecycle模块，提取摆动点识别函数"
```

---

## 🐛 常见问题

### Q1: 命令不生效
**A:** 检查命令文件是否存在：
```bash
ls .claude/commands/
cat .claude/commands/review.md
```

### Q2: Hook不触发
**A:** 确保hook已复制到.git/hooks/并有执行权限：
```bash
cp .claude/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Q3: Claude不加载CLAUDE.md
**A:** 
1. 文件名必须全大写：`CLAUDE.md`
2. 必须在项目根目录
3. 重新启动Claude会话

### Q4: 权限模式切换不生效
**A:** 编辑 `.claude/settings.json`，然后重启Claude

---

## 📚 参考资料

- **Claude Code官方文档：** https://docs.anthropic.com/claude/docs/code
- **中文Skills仓库：** https://github.com/laolaoshiren/claude-code-skills-zh
- **腾讯云教程系列：** https://cloud.tencent.com/developer/article/2637676

---

## ✅ 配置检查清单

- [x] CLAUDE.md已创建并包含项目信息
- [x] .claude/settings.json已配置（模型+权限）
- [x] 5个自定义命令已创建（/review, /test, /doc, /commit, /golden）
- [x] pre-commit hook已创建
- [ ] pre-commit hook已复制到.git/hooks/（需手动）
- [ ] Skills已安装（可选，按需安装）
- [x] .gitignore已配置
- [x] 安全Deny规则已设置

---

**下一步：** 测试配置，开始Day 1-2开发任务
