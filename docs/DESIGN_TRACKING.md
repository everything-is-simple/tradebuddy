# TradeBuddy 设计追踪与实现一致性保障机制

> Design Tracking and Implementation Consistency Assurance  
> Version 1.0 - 2026-10-09

---

## 目标

确保实现不偏离设计，建立可验证的追踪机制。

---

## 1. 文档体系与追踪矩阵

### 1.1 核心文档关系

```
需求规格说明（REQUIRE.md）
         ↓
    系统设计（DESIGN.md）
         ↓
    接口规范（API.md）
         ↓
    技术实现（TRD.md）
         ↓
    实际代码（internal/*/）
```

**追踪矩阵：**

| 需求ID | 设计章节 | API接口 | TRD章节 | 实现文件 | 测试文件 | 状态 |
|--------|---------|---------|---------|----------|----------|------|
| REQ-001 | 2.2.1 Screener | screener.Run() | 1.1 三条件筛选 | internal/screener/screener.go | screener_test.go | ✓ |
| REQ-002 | 2.2.2 Lifecycle | lifecycle.Classify() | 1.2 摆动点识别 | internal/lifecycle/classifier.go | classifier_test.go | 进行中 |
| REQ-003 | 2.2.3 Tachibana | tachibana.BuildPlan() | 1.3 立花梯价 | internal/tachibana/planner.go | planner_test.go | 待开始 |

**维护方式：**
- 文件位置：`docs/TRACEABILITY.md`
- 更新时机：每个功能完成时更新状态
- 审查频率：每周Code Review时检查

---

## 2. 设计变更管理流程

### 2.1 变更识别

**触发条件：**
- 需求变更（用户反馈、业务调整）
- 技术约束（性能瓶颈、技术限制）
- 架构优化（重构、技术债清理）

**变更级别：**
- **重大变更：** 影响系统架构、数据模型、核心算法
- **中等变更：** 影响模块接口、数据流
- **小变更：** 实现细节优化

### 2.2 变更审批流程

```
变更提出
    ↓
architect审查（重大变更需Opus 5.5深度分析）
    ↓
更新设计文档（DESIGN.md/API.md/TRD.md）
    ↓
记录ADR（架构决策）
    ↓
实现代码
    ↓
更新测试（Golden Test必须通过）
    ↓
code-reviewer审查
    ↓
更新追踪矩阵
```

### 2.3 变更记录

**文件位置：** `docs/CHANGELOG_DESIGN.md`

**格式：**
```markdown
## [2026-10-15] - Lifecycle算法优化

**变更类型：** 中等变更  
**影响范围：** Lifecycle模块  
**变更原因：** 摆动点识别误判率高（15% → 5%）

**设计变更：**
- 窗口大小：2周 → 4周
- 判定条件：严格等于 → 严格大于

**文档更新：**
- [x] DESIGN.md §2.2.2
- [x] TRD.md §1.2
- [x] ADR.md ADR-005

**实现变更：**
- internal/lifecycle/swing.go

**测试更新：**
- Golden Test通过：✓
- 单元测试覆盖：65%

**审查人：** architect
```

---

## 3. Golden Test作为设计契约

### 3.1 Golden Test数据集

| 数据集名称 | 日期 | 数量 | 验证目标 | 位置 |
|-----------|------|------|---------|------|
| Screener初选 | 2026-10-08 | 27只 | 三条件筛选正确性 | testdata/golden/screen_20261008.json |
| Lifecycle阶段 | 2026-10-08 | 27只 | 九阶段分类准确率≥90% | testdata/golden/lifecycle_20261008.json |
| Tachibana计划 | 2026-10-08 | 27只 | 梯价误差≤0.5% | testdata/golden/tachibana_20261008.json |
| 立花账本 | 1975-1976 | 159笔 | 恒等式100%通过 | testdata/golden/tachibana_197576.json |

### 3.2 Golden Test即设计规格

**原则：**
- Golden Test = 可执行的设计文档
- Golden Test失败 = 设计偏离
- 修改Golden Test = 设计变更（需ADR记录）

**验收标准嵌入测试：**
```go
func TestScreener_20261008_Golden(t *testing.T) {
    // 这个测试就是REQUIRE.md §2.2.1的验收标准
    result, err := screener.Run(ctx, "2026-10-08")
    require.NoError(t, err)
    
    // 验收标准1：产出27只股票
    assert.Equal(t, 27, result.Count, "REQUIRE.md §2.2.1验收标准1")
    
    // 验收标准2：涨幅误差≤0.01%
    for _, stock := range result.Stocks {
        golden := goldenData.FindByCode(stock.Code)
        assert.InDelta(t, golden.PctChange, stock.PctChange, 0.01,
            "REQUIRE.md §2.2.1验收标准2: %s涨幅误差超标", stock.Code)
    }
}
```

---

## 4. CI/CD集成

### 4.1 GitHub Actions工作流

**文件位置：** `.github/workflows/design-check.yml`

```yaml
name: Design Consistency Check

on: [push, pull_request]

jobs:
  design-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Run Golden Tests
        run: go test -v -run Golden ./...
      
      - name: Check Test Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out | grep total | awk '{print $3}'
          # 核心模块覆盖率≥60%
      
      - name: Run Architecture Checks
        run: |
          # 检查循环依赖
          go mod graph | grep cycle && exit 1 || echo "No cycles"
          
          # 检查文档一致性（使用architect agent）
          # claude -p "使用architect agent审查设计一致性"
```

### 4.2 Pre-commit Hook

**已配置：** `.claude/hooks/pre-commit`

检查项：
- [x] gofmt格式
- [x] go vet静态分析
- [x] 短测试（go test -short）
- [x] 敏感文件拦截

---

## 5. 定期审查机制

### 5.1 每日自查（开发者）

**使用architect agent：**
```bash
# 检查今日代码与设计一致性
/agents

> 使用architect agent审查今日改动：
> - 检查internal/store/与API.md一致性
> - 验证Golden Test通过
> - 确认文档已更新
```

### 5.2 每周Code Review（团队）

**审查清单：**
- [ ] 追踪矩阵更新
- [ ] Golden Test全部通过
- [ ] 核心模块测试覆盖率≥60%
- [ ] ADR记录重大决策
- [ ] 设计变更日志完整
- [ ] 无循环依赖
- [ ] 性能目标达成

### 5.3 里程碑审查（阶段）

**第一期结束前（2周后）：**
- [ ] 完整运行architect审查报告
- [ ] 所有Golden Test通过
- [ ] 文档与代码100%一致
- [ ] 追踪矩阵无遗漏
- [ ] ADR记录所有重大决策

---

## 6. 工具链支持

### 6.1 Claude Code Agent

**code-reviewer：** 日常代码审查
- 检查代码规范
- 验证错误处理
- 确认TradeBuddy特定要求

**architect：** 架构级审查
- 设计与实现一致性
- 架构原则遵守
- 性能目标达成
- ADR维护

**test-writer：** 测试生成
- 表驱动测试
- Golden Test
- Mock测试

**调用方式：**
```bash
# 明确指定agent
使用code-reviewer审查 internal/store/store.go

# 或让Claude自动匹配
检查设计一致性（自动调用architect）
生成单元测试（自动调用test-writer）
```

### 6.2 自定义命令

**/review：** 代码审查
**/golden：** 运行Golden Test
**/commit：** 规范提交（附追踪信息）

---

## 7. 文档版本控制

### 7.1 文档版本号

**格式：** `vMAJOR.MINOR (YYYY-MM-DD)`

**版本规则：**
- MAJOR：架构重大变更
- MINOR：模块接口变更
- 日期：文档修订日期

**当前版本：**
- REQUIRE.md: v1.0 (2026-10-09)
- DESIGN.md: v1.0 (2026-10-09)
- API.md: v1.0 (2026-10-09)
- TRD.md: v1.0 (2026-10-09)
- ERD.md: v1.0 (2026-10-09)
- ADR.md: v1.0 (2026-10-09)

### 7.2 文档更新触发

**必须更新文档的场景：**
- 新增接口或修改接口签名 → API.md
- 修改数据库Schema → ERD.md
- 变更核心算法 → TRD.md
- 重大技术决策 → ADR.md
- 新增功能或修改需求 → REQUIRE.md
- 架构调整 → DESIGN.md

---

## 8. 偏离告警机制

### 8.1 自动检测

**检测项：**
1. **接口偏离：** API.md描述的函数签名 vs 实际代码
2. **Schema偏离：** ERD.md的表结构 vs 实际SQL
3. **Golden Test失败：** 算法输出 vs 预期结果
4. **性能退化：** 实际耗时 vs TRD.md目标

**检测方式：**
- CI/CD自动运行
- Pre-commit hook拦截
- architect agent审查

### 8.2 人工审查

**使用architect agent：**
```
> 全面审查设计一致性，生成审查报告

architect会：
1. 对比DESIGN.md与internal/*/代码
2. 对比API.md与函数签名
3. 对比ERD.md与migrations/*.sql
4. 对比TRD.md与算法实现
5. 检查Golden Test状态
6. 生成结构化审查报告
```

---

## 9. 示例：一次完整的功能开发

### 步骤1：需求到设计
```
1. 在REQUIRE.md添加需求REQ-004
2. 在DESIGN.md设计模块接口
3. 在API.md定义函数签名
4. 在TRD.md描述算法实现
5. 准备Golden Test数据集
```

### 步骤2：实现前审查
```bash
# 使用architect审查设计
使用architect agent审查REQ-004的设计方案

# architect检查：
# - 需求完整性
# - 接口合理性
# - 与现有架构兼容性
# - 性能目标可行性
```

### 步骤3：编码实现
```go
// 实现代码，严格按照API.md的接口签名
// 算法实现遵循TRD.md的描述
// 附上理论引用注释
```

### 步骤4：测试生成
```bash
# 使用test-writer生成测试
使用test-writer为internal/module/feature.go生成测试

# 包括：
# - 表驱动测试
# - Golden Test（使用准备好的数据集）
# - Mock测试
```

### 步骤5：运行Golden Test
```bash
/golden

# 验证：
# ✓ Golden Test通过
# ✓ 输出与预期一致
```

### 步骤6：代码审查
```bash
# 使用code-reviewer审查
使用code-reviewer审查 internal/module/feature.go

# 检查：
# - 代码规范
# - 错误处理
# - TradeBuddy特定要求
```

### 步骤7：提交前检查
```bash
# Pre-commit hook自动运行
git commit -m "feat(module): 实现REQ-004功能"

# hook检查：
# ✓ gofmt格式
# ✓ go vet通过
# ✓ 测试通过
# ✓ 无敏感文件
```

### 步骤8：更新追踪矩阵
```markdown
| REQ-004 | 2.2.4 Module | module.Feature() | 1.4 算法 | internal/module/feature.go | feature_test.go | ✓ |
```

### 步骤9：最终一致性审查
```bash
# 使用architect全面审查
使用architect agent审查REQ-004的完整实现

# architect验证：
# ✓ 设计与实现一致
# ✓ API文档与代码一致
# ✓ Golden Test通过
# ✓ 测试覆盖充分
# ✓ 追踪矩阵更新
```

---

## 10. 检查清单（Checklist）

### 新功能开发检查清单

- [ ] REQUIRE.md添加需求
- [ ] DESIGN.md设计模块
- [ ] API.md定义接口
- [ ] TRD.md描述算法
- [ ] ERD.md更新（如涉及数据库）
- [ ] ADR.md记录（如有架构决策）
- [ ] 准备Golden Test数据
- [ ] 实现代码
- [ ] 生成单元测试
- [ ] Golden Test通过
- [ ] code-reviewer审查通过
- [ ] architect审查通过
- [ ] 更新追踪矩阵
- [ ] 更新CHANGELOG_DESIGN.md

### 设计变更检查清单

- [ ] 识别变更级别（重大/中等/小）
- [ ] architect审查变更合理性
- [ ] 更新受影响的文档
- [ ] 记录ADR（重大变更）
- [ ] 修改Golden Test（如需要）
- [ ] 实现代码变更
- [ ] 回归测试通过
- [ ] 追踪矩阵更新
- [ ] 记录CHANGELOG_DESIGN.md

---

**维护说明：**
- 本文档随项目演进持续更新
- 每个里程碑结束后回顾机制有效性
- 根据实践经验优化流程

**下一步：** 提交所有配置到Git，开始Day 1-2开发
