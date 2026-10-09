# TradeBuddy 技术验证指南

## 环境准备

### 1. 安装 Go 1.22+

**下载地址：** https://go.dev/dl/

- Windows: `go1.22.8.windows-amd64.msi`
- 安装后验证：打开 PowerShell 运行 `go version`

### 2. 验证环境变量

```powershell
# 检查 Go 是否在 PATH 中
go version

# 如果报错，手动添加到 PATH：
# C:\Program Files\Go\bin
```

---

## 三个验证Demo

### Demo 1: SQLite 读写测试

**目的：** 验证 Go 能否正常读写 SQLite 数据库

**运行：**
```powershell
cd I:\tradebuddy
go mod tidy
go run cmd/verify/demo1_sqlite.go
```

**预期输出：**
```
=== Demo 1: Go + SQLite 验证 ===

查询结果:
代码		名称		板块
----------------------------------------
sh600519	贵州茅台	主板
sh688981	中芯国际	科创板
sz000001	平安银行	主板
sz300750	宁德时代	创业板
----------------------------------------
✓ 验证成功: 插入4条，查询4条
✓ 数据库文件: data\test_demo1.db (12.00 KB)

✓✓✓ Demo 1 通过：Go + SQLite 读写正常 ✓✓✓
```

---

### Demo 2: Excel 解析测试

**目的：** 验证 Go 能否读取立花交易计划 Excel

**运行：**
```powershell
go run cmd/verify/demo2_excel.go
```

**预期输出：**
```
=== Demo 2: Go 读取 Excel 验证 ===

文件: 证据③20_00立花交易计划_10-08（回档布网+跟踪池）.xlsx
Sheet: 今晚交易计划
总行数: 35

表头字段:
  列0: 代码
  列1: 名称
  列2: 决策
  列3: z值
  ...

前5条数据:
代码	名称	决策	z值
----------------------------------------
600519	贵州茅台	布网	1.23
000001	平安银行	观察	0.45
...
----------------------------------------
共找到 27 条数据记录

✓✓✓ Demo 2 通过：Excel 读取正常 ✓✓✓
```

---

### Demo 3: 通达信 .day 文件读取

**目的：** 验证 Go 能否读取通达信本地日线数据

**运行：**
```powershell
go run cmd/verify/demo3_tdx.go
```

**预期输出：**
```
=== Demo 3: 通达信 .day 文件验证 ===
找到通达信目录: C:\new_tdx64\vipdoc

文件: sh600519.day
K线数量: 5432

最近5根K线:
日期		开盘	最高	最低	收盘	成交量
------------------------------------------------------------
20261004	1850.00	1875.50	1842.30	1868.20	123456789
20261007	1870.00	1890.00	1865.00	1882.50	98765432
20261008	1880.00	1895.00	1878.00	1888.00	87654321
20261009	1885.00	1900.00	1880.00	1895.50	76543210
20261010	1895.00	1910.00	1892.00	1905.20	65432109
------------------------------------------------------------

✓✓✓ Demo 3 通过：TDX .day 文件读取正常 ✓✓✓
```

---

## 验证清单

完成以下3项验证后，即可开始正式开发：

- [ ] Demo 1: SQLite 读写 ✓
- [ ] Demo 2: Excel 解析 ✓
- [ ] Demo 3: TDX 文件读取 ✓

---

## 常见问题

### Q1: go: command not found

**A:** Go 未安装或未加入 PATH，请重新安装 Go 并确保环境变量配置正确。

### Q2: modernc.org/sqlite 安装失败

**A:** 
```powershell
# 方法1：使用国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy

# 方法2：手动下载
go get -v modernc.org/sqlite@latest
```

### Q3: Demo 3 提示"未找到通达信安装目录"

**A:** 手动修改 `demo3_tdx.go` 中的 `tdxPaths` 数组，添加你的实际路径：
```go
tdxPaths := []string{
    "你的通达信路径\\vipdoc",  // 例如 D:\\Software\\TDX\\vipdoc
    "C:\\new_tdx64\\vipdoc",
    "C:\\tdx\\vipdoc",
}
```

### Q4: Excel 文件路径错误

**A:** 确保已归档 glm 文档到 `I:\tradebuddy-archived\glm-docs-20261009\`，或修改 `demo2_excel.go` 中的路径。

---

## 验证成功后

所有3个Demo通过后，说明技术栈可行，可以开始正式开发：

1. **第一步：** 初始化数据库结构（`internal/store/migrations/0001_init.sql`）
2. **第二步：** 实现数据层（`internal/store/store.go`）
3. **第三步：** 实现核心引擎（screener → lifecycle → tachibana）
4. **第四步：** 实现 CLI 工具（`cmd/tb-cli/main.go`）

详见 `docs/DESIGN.md` 开发计划。
