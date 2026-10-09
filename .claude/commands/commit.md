---
description: 提交代码到Git
argument-hint: 提交信息（可选）
model: claude-haiku-5-5
allowed-tools: Bash(git *)
---

# Git 工作流

自动完成Git提交流程：

**提交信息：** $ARGUMENTS

## 执行步骤

### 1. 检查状态
```bash
git status
git diff --stat
```

### 2. 生成提交信息

如果用户未提供提交信息，根据改动自动生成：

**格式规范：**
```
<type>(<scope>): <subject>

<body>

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
```

**Type类型：**
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- test: 测试相关
- refactor: 重构
- chore: 杂项（构建、依赖等）

**Scope范围：**
- screener: 初选引擎
- lifecycle: 生命周期
- tachibana: 立花计划
- store: 数据访问层
- cli: 命令行工具
- docs: 文档

**示例：**
```
feat(screener): 实现三条件过滤器

- 条件A：涨幅>6%
- 条件B：创20日新高
- 条件C：周月线趋势过滤

通过Golden test验证：10-08产出27只

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
```

### 3. 暂存文件

```bash
git add .
```

### 4. 提交

```bash
git commit -m "<生成的提交信息>"
```

### 5. 推送（询问用户）

询问是否推送到远程：
```bash
git push origin main
```

## 安全检查

提交前检查：
- [ ] 是否有 .env 文件被暂存（拒绝提交）
- [ ] 是否有 *.db 文件被暂存（提醒用户）
- [ ] 是否有大文件 >5MB（提醒用户）
- [ ] 是否在主分支上（提醒用户创建feature分支）

如发现问题，终止提交并给出建议。
