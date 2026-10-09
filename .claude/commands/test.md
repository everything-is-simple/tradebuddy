---
description: 生成单元测试
argument-hint: 文件路径
model: claude-sonnet-5-5
---

# Go 单元测试生成

为以下代码生成单元测试：

**目标文件：** $ARGUMENTS

## 测试要求

### 覆盖范围
- 正常路径（happy path）
- 边界条件（空值、零值、极值）
- 错误处理路径
- 并发安全（如适用）

### 测试模式
使用表驱动测试（Table-Driven Tests）：

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "正常情况",
            input: ...,
            want:  ...,
        },
        {
            name:    "错误情况：空输入",
            input:   nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Mock处理
- 数据库操作：使用内存SQLite或mock
- 外部API：使用httptest或mock
- 时间依赖：接受time.Time参数而非time.Now()

### 断言库
优先使用标准库testing，复杂场景可用：
- `github.com/stretchr/testify/assert`
- `github.com/stretchr/testify/require`

生成测试后，执行 `go test -v` 验证通过。
