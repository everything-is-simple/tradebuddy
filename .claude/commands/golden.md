---
description: 运行Golden Test验证
model: claude-sonnet-5-5
allowed-tools: Bash(go test *)
---

# Golden Test 验证

运行Golden Test验证实现正确性。

## 什么是Golden Test？

用已知正确结果验证新实现，确保：
1. 算法实现与理论一致
2. 代码重构不破坏功能
3. 不同实现产出相同结果

## 当前Golden Test数据集

### 1. Screener - 2026-10-08初选
**数据源：** `docs/证据①19_00备选清单_10-08.xlsx`
**预期结果：** 27只股票
**关键股票：** 605133（世名科技）、600825（新华传媒）

```bash
go test -v ./internal/screener -run TestScreener_20261008
```

### 2. Lifecycle - 2026-10-08生命周期
**数据源：** `docs/证据②19_30生命周期_10-08.xlsx`
**预期结果：** 27只股票的阶段和概率
**验证项：** 阶段分类准确率 ≥ 90%，概率偏差 ≤ 2.5pp

```bash
go test -v ./internal/lifecycle -run TestLifecycle_20261008
```

### 3. Tachibana - 2026-10-08交易计划
**数据源：** `docs/证据③20_00立花交易计划_10-08.xlsx`
**预期结果：** 布网/观察/排除决策一致
**验证项：** 梯价误差 ≤ 0.5%，仓位误差 ≤ 100股

```bash
go test -v ./internal/tachibana -run TestTachibana_20261008
```

### 4. 立花账本 - 1975-1976验证
**数据源：** `docs/立花玉帳数字化_1975-1976.xlsx`
**预期结果：** 159笔成交，100%通过恒等式
**验证项：** 权益曲线一致，无差值

```bash
go test -v ./internal/tachibana -run TestTachibana_19751976
```

## 执行步骤

1. **运行全部Golden Tests：**
   ```bash
   go test -v ./... -run Golden
   ```

2. **查看详细输出：**
   ```bash
   go test -v -json ./... -run Golden | tee golden_test.log
   ```

3. **失败时的调试：**
   ```bash
   go test -v -run Golden -count=1  # 禁用缓存
   go test -v -run Golden -timeout 5m  # 延长超时
   ```

## 验收标准

- [ ] 所有Golden Tests通过
- [ ] 误差在允许范围内
- [ ] 执行时间未显著增加（< 10%）

## 失败处理

如Golden Test失败：
1. 确认是算法错误还是测试数据问题
2. 如果是算法错误，修复后重新验证
3. 如果是测试数据问题，更新Golden数据集（需用户确认）
4. 记录失败原因到 `docs/TASK.md`

Golden Test是质量保证的最后防线，必须100%通过才能合并到main分支。
