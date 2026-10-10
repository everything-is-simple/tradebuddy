# TradeBuddy 定时任务配置指南

## 📋 快速配置 Windows 任务计划

### 方法 1: 使用批处理脚本 (推荐)

#### 步骤 1: 测试脚本
```bash
# 双击运行测试
run_daily.bat
```

#### 步骤 2: 创建计划任务
1. 按 `Win + R`，输入 `taskschd.msc`，回车
2. 右侧点击 "创建基本任务"
3. 填写信息：
   - **名称**: `TradeBuddy每日分析`
   - **描述**: `19:00自动运行选股和分析`
4. 触发器: 选择 "每天"
   - 开始时间: **19:00:00**
   - 每隔: **1天**
5. 操作: 选择 "启动程序"
   - 程序/脚本: `I:\tradebuddy\run_daily.bat`
   - 起始位置: `I:\tradebuddy`
6. 完成后勾选 "打开此任务的属性对话框"
7. 在属性对话框中设置：
   - **常规** 标签：
     - ☑ 不管用户是否登录都要运行
     - ☑ 使用最高权限运行
   - **条件** 标签：
     - ☐ 只有在计算机使用交流电源时才启动 (取消勾选)
     - ☑ 如果错过了计划的开始时间，立即启动任务
   - **设置** 标签：
     - ☑ 允许按需运行任务
     - ☑ 如果任务失败，每隔 1 分钟重试 3 次

---

### 方法 2: 使用 PowerShell (备选)

#### 步骤 1: 测试脚本
```powershell
# 右键以管理员身份运行 PowerShell
powershell -ExecutionPolicy Bypass -File run_daily.ps1
```

#### 步骤 2: 创建计划任务
同方法 1，但在"操作"步骤中填写：
- 程序/脚本: `powershell.exe`
- 添加参数: `-ExecutionPolicy Bypass -File "I:\tradebuddy\run_daily.ps1"`
- 起始位置: `I:\tradebuddy`

---

## 🎯 当前可配置的定时任务

### 核心任务 (必需)

| 时间 | 任务名称 | 执行内容 | 耗时 |
|------|---------|---------|------|
| 19:00 | 每日分析 | 完整流程 (screen→lifecycle→tachibana) | 5-10分钟 |

**使用脚本**: `run_daily.bat` 或 `run_daily.ps1`

---

### 可选任务

#### 1. 报告查看提醒
- **时间**: 20:30
- **命令**: `tradebuddy.exe report --date %date%`
- **用途**: 查看当日分析结果统计

#### 2. 周末回测
- **时间**: 周六 10:00
- **命令**: 批量处理历史数据
- **用途**: 历史数据验证

---

## 📊 查看执行日志

### 日志位置
```
I:\tradebuddy\logs\daily_YYYYMMDD.log
```

### 查看最新日志
```bash
# 使用记事本
notepad logs\daily_%date:~0,4%%date:~5,2%%date:~8,2%.log

# 或使用 PowerShell
Get-Content logs\daily_*.log -Tail 50
```

---

## ✅ 验证任务是否正常运行

### 检查清单

1. **任务计划程序中查看**
   ```
   任务计划程序库 → TradeBuddy每日分析
   → 右键 → 运行 (手动测试)
   ```

2. **查看上次运行结果**
   - 状态应为 "成功 (0x0)"
   - 上次运行时间应为 19:00 左右

3. **检查生成的文件**
   ```bash
   # 检查报告
   dir reports\*%date:~0,4%%date:~5,2%%date:~8,2%*.xlsx
   
   # 检查数据库
   dir data\tradebuddy.db
   
   # 检查日志
   type logs\daily_%date:~0,4%%date:~5,2%%date:~8,2%.log
   ```

---

## 🔧 故障排查

### 问题 1: 任务未执行

**检查**:
- 计算机是否在 19:00 开机
- 任务计划程序服务是否运行
- 用户权限是否足够

**解决**:
```bash
# 检查任务计划程序服务
services.msc → Task Scheduler → 确保"正在运行"
```

---

### 问题 2: 执行失败

**检查日志**:
```bash
type logs\daily_*.log
```

**常见错误**:
- TDX 数据目录不存在 → 检查 `H:\new_tdx64\vipdoc`
- 权限不足 → 任务使用"最高权限运行"
- 路径错误 → 确认起始位置为 `I:\tradebuddy`

---

### 问题 3: 报告未生成

**检查**:
```bash
# 查看 reports 目录
dir reports

# 手动运行测试
tradebuddy.exe run --date 2026-06-15
```

---

## 📝 配置示例

### 完整的任务计划 XML 导出

如需备份或迁移任务配置：
1. 任务计划程序 → 选择任务
2. 右键 → 导出
3. 保存为 `TradeBuddy每日分析.xml`

---

## 🚀 推荐配置

### 标准配置 (推荐)

```
任务名称: TradeBuddy每日分析
触发时间: 每日 19:00
执行脚本: run_daily.bat
重试策略: 失败后每1分钟重试，最多3次
执行权限: 最高权限
```

### 高级配置 (可选)

如果需要更复杂的调度，可以创建多个任务：

**任务 1**: 初选
- 时间: 19:00
- 命令: `tradebuddy.exe screen`

**任务 2**: 生命周期
- 时间: 19:10
- 命令: `tradebuddy.exe lifecycle`

**任务 3**: 立花提示
- 时间: 19:20
- 命令: `tradebuddy.exe tachibana`

---

## 📅 下一步

1. ✅ **立即执行**: 配置 Windows 任务计划程序
2. ⏳ **明天验证**: 检查任务是否自动运行
3. ⏳ **一周后**: 检查所有日志和报告

---

**配置完成后，系统将每天 19:00 自动运行分析流程！** 🎉
