---
description: 生成函数文档注释
argument-hint: 函数名或文件路径
model: claude-haiku-5-5
---

# Go 文档注释生成

为以下代码生成规范的文档注释：

**目标：** $ARGUMENTS

## 注释规范

### 函数注释模板

```go
// FunctionName 功能简述（中文，一句话）
//
// 详细说明（可选，如有复杂逻辑）
//
// 参数说明：
//   - param1: 参数说明
//   - param2: 参数说明
//
// 返回值：
//   - 返回值说明
//   - error: 错误说明
//
// 理论来源（如适用）：原著名称 + 页码
// 验证依据（如适用）：Golden test数据集
//
// 示例：
//
//	result := FunctionName(param1, param2)
//	if result != nil {
//	    // ...
//	}
func FunctionName(param1 Type1, param2 Type2) (ResultType, error) {
    // ...
}
```

### 包注释

```go
// Package store 提供SQLite数据访问层
//
// 包含所有数据表的CRUD操作和数据库迁移逻辑。
// 使用modernc.org/sqlite纯Go实现，无CGO依赖。
package store
```

### 类型注释

```go
// ScreenResult 表示19:00强势股初选结果
//
// 包含筛选后的股票基本信息和技术指标。
// 筛选条件：涨幅>6% + 创20日新高 + 周月线趋势过滤
type ScreenResult struct {
    // TradeDate 交易日期 YYYY-MM-DD
    TradeDate string
    // Code 股票代码 sh600519
    Code string
    // ...
}
```

## 特殊要求

1. 所有导出标识符（首字母大写）必须有注释
2. 注释第一句必须以标识符名称开头
3. 关键算法必须附理论来源
4. 统计方法必须说明样本量和验证方式

生成注释后，运行 `go vet` 检查格式。
