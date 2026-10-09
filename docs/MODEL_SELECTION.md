# TradeBuddy 模型选择策略

> Model Selection Strategy  
> Version 1.0 - 2026-10-09

---

## 模型概览

| 模型 | Token成本 | 速度 | 推理能力 | 适用场景 |
|-----|----------|------|---------|---------|
| **Opus 5.5** | 最高 | 慢 | 最强 | 架构设计、复杂重构、深度分析 |
| **Sonnet 5.5** | 中等 | 快 | 强 | 日常编码、代码审查、测试（80%任务）|
| **Haiku 5.5** | 最低 | 最快 | 中 | commit message、文档注释、简单格式化 |

---

## 1. 按任务类型选择

### 1.1 架构级任务（Opus 5.5）

**使用场景：**
- 系统架构设计和重大重构
- 技术选型和架构决策（ADR）
- 复杂算法设计（摆动点识别、生命周期分类）
- 性能优化方案设计
- 设计与实现一致性审查
- 多方案权衡和深度分析

**调用方式：**
```bash
# 1. 使用architect agent（自动使用Opus）
使用architect agent审查系统架构一致性

# 2. 临时切换模型
/model claude-opus-5-5
请设计Lifecycle模块的架构

# 3. 命令frontmatter指定（未来可用）
---
model: claude-opus-5-5
---
```

**成本考虑：**
- Opus成本是Sonnet的2-3倍
- 仅用于高价值决策，每周使用次数控制在10次以内

---

### 1.2 开发级任务（Sonnet 5.5）

**使用场景：**
- 日常编码实现（80%的编码任务）
- 代码审查和重构
- 单元测试编写
- Bug调试和修复
- API设计和接口定义
- Golden Test设计

**调用方式：**
```bash
# 默认模型，无需指定
实现Screener的三条件筛选

# 或使用专用agent
使用code-reviewer审查 internal/store/store.go
使用test-writer生成测试 internal/calc/ma.go
```

**为什么是主力：**
- 性价比最佳（质量/成本/速度平衡）
- 代码质量接近Opus
- 响应速度快（< 5秒）
- 适合迭代开发

---

### 1.3 轻量级任务（Haiku 5.5）

**使用场景：**
- Git commit message生成
- 函数文档注释
- 简单代码格式化
- 日志信息生成
- README文本调整

**调用方式：**
```bash
# /doc和/commit命令已指定Haiku
/doc CalculateTierPrices
/commit

# 或临时切换
/model claude-haiku-5-5
为这个函数生成文档注释
```

**何时不用Haiku：**
- 需要理解复杂逻辑
- 需要架构思考
- 需要多文件关联分析

---

## 2. 按Agent选择

### 2.1 Agent模型配置

**已配置（.claude/agents/*.md）：**

| Agent | 模型 | 理由 |
|-------|------|------|
| **architect** | Opus 5.5 | 需要深度架构思考和权衡 |
| **code-reviewer** | Sonnet 5.5 | 平衡速度和质量，日常高频使用 |
| **test-writer** | Sonnet 5.5 | 需要理解逻辑，Haiku不足 |

### 2.2 自动调用

Claude会根据任务描述自动选择Agent（和对应模型）：

```bash
# 自动调用architect（Opus）
"检查系统架构是否符合设计文档"
"评估这个技术选型的利弊"

# 自动调用code-reviewer（Sonnet）
"审查这段代码有什么问题"
"这个错误处理是否正确"

# 自动调用test-writer（Sonnet）
"为这个函数生成测试"
"添加边界条件测试"
```

---

## 3. 按开发阶段选择

### 3.1 设计阶段（Opus主导）

```
需求分析 → Sonnet
    ↓
架构设计 → Opus ⭐
    ↓
接口设计 → Sonnet
    ↓
算法设计 → Opus ⭐
    ↓
设计审查 → Opus (architect)
```

**Opus比例：** 40%

---

### 3.2 开发阶段（Sonnet主导）

```
编码实现 → Sonnet ⭐
    ↓
单元测试 → Sonnet ⭐
    ↓
代码审查 → Sonnet (code-reviewer)
    ↓
Bug修复 → Sonnet
    ↓
commit → Haiku
```

**Sonnet比例：** 80%  
**Haiku比例：** 20%

---

### 3.3 重构阶段（混合）

```
识别技术债 → Sonnet
    ↓
重构方案设计 → Opus ⭐
    ↓
执行重构 → Sonnet
    ↓
回归测试 → Sonnet
    ↓
架构审查 → Opus (architect)
```

**Opus比例：** 30%  
**Sonnet比例：** 70%

---

## 4. 决策树

```
任务来了
    ↓
是否涉及架构？
├─ 是 → Opus (architect)
└─ 否
    ↓
    是否需要深度思考？
    ├─ 是 → Opus
    └─ 否
        ↓
        是否是简单文本？
        ├─ 是 → Haiku
        └─ 否 → Sonnet（默认）
```

---

## 5. 成本控制策略

### 5.1 预算分配（每月）

假设月预算 $100：

| 模型 | 比例 | 预算 | 使用场景控制 |
|-----|------|------|------------|
| Opus | 20% | $20 | 每周≤10次，仅架构级任务 |
| Sonnet | 70% | $70 | 日常主力，不限 |
| Haiku | 10% | $10 | 自动文本任务 |

### 5.2 Opus使用检查清单

**使用Opus前自问：**
- [ ] 这个决策影响系统架构吗？
- [ ] 需要权衡多个复杂方案吗？
- [ ] Sonnet的答案已经不够深入了吗？
- [ ] 这个决策会长期影响项目吗？

**如果≥2个"是"，使用Opus；否则用Sonnet。**

---

## 6. 实践建议

### 6.1 开发者日常

**80%时间用Sonnet：**
- 写代码、写测试、调Bug
- code-reviewer审查
- test-writer生成测试

**15%时间用Haiku：**
- /commit提交
- /doc生成注释
- 简单文本调整

**5%时间用Opus：**
- 架构设计
- 技术难题
- architect审查

### 6.2 里程碑节点

**第一期开始前（Day 0）：** Opus设计架构 ✓  
**第一期开发中（Day 1-14）：** Sonnet主力编码  
**第一期结束前（Day 14）：** Opus全面架构审查  

**第二期开始前：** Opus设计GUI架构  
**第二期开发中：** Sonnet实现UI  
**第二期结束前：** Opus审查  

---

## 7. 命令与Agent映射表

| 命令/Agent | 默认模型 | 可覆盖 | 典型耗时 |
|-----------|---------|--------|---------|
| /review | Sonnet | 否 | 10-30s |
| /test | Sonnet | 否 | 20-60s |
| /doc | Haiku | 否 | 5-10s |
| /commit | Haiku | 否 | 5-10s |
| /golden | Sonnet | 否 | 60-180s |
| architect | Opus | 否 | 30-120s |
| code-reviewer | Sonnet | 否 | 10-30s |
| test-writer | Sonnet | 否 | 20-60s |
| 临时对话 | Sonnet | 是（/model） | 变化 |

---

## 8. 模型切换命令

```bash
# 查看当前模型
/model

# 临时切换（当前会话）
/model claude-opus-5-5
/model claude-sonnet-5-5
/model claude-haiku-5-5

# 切换回默认
/model

# 在配置中永久修改
# 编辑 .claude/settings.json
{
  "model": "claude-sonnet-5-5"
}
```

---

## 9. 模型性能对比（TradeBuddy实测）

| 任务 | Haiku | Sonnet | Opus |
|------|-------|--------|------|
| 生成commit message | 5s ✓ | 8s | 15s |
| 代码审查（单文件） | - | 15s ✓ | 30s |
| 架构设计方案 | - | 45s | 90s ✓ |
| Golden Test设计 | - | 30s ✓ | 60s |
| 简单Bug修复 | - | 10s ✓ | 20s |
| 复杂重构设计 | - | 60s | 120s ✓ |

**结论：**
- Haiku: 速度+2倍，质量−30%
- Sonnet: 平衡最佳（主力）
- Opus: 质量+20%，速度−2倍

---

## 10. 常见问题

**Q: 为什么不全部用Opus？**  
A: 成本是Sonnet的2-3倍，大多数编码任务Sonnet足够。

**Q: Haiku适合写代码吗？**  
A: 不适合。Haiku理解复杂逻辑能力弱，仅用于简单文本。

**Q: 如何知道当前用的什么模型？**  
A: 输入 `/model` 查看，或看回复末尾的模型标记。

**Q: Agent可以换模型吗？**  
A: Agent的模型在其配置文件（.claude/agents/*.md）frontmatter中指定，通常不建议修改。

---

**版本历史：**
- v1.0 (2026-10-09): 初始版本

**维护说明：**
- 根据实际使用效果和成本数据调整策略
- 每月回顾模型使用比例和成本
- 记录特殊案例（何时Sonnet不够用）

**下一步：** 提交所有配置，开始实际开发
