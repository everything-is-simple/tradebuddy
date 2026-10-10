# TradeBuddy 每日自动分析脚本 (PowerShell)
# 用途: Windows 任务计划程序自动执行

Set-Location "I:\tradebuddy"

# 创建日志目录
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

# 获取当前日期
$date = Get-Date -Format "yyyy-MM-dd"
$logDate = Get-Date -Format "yyyyMMdd"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# 执行完整流程
Write-Output "[$timestamp] 开始每日分析..." | Out-File -Append "logs\daily_$logDate.log"

.\tradebuddy.exe run --date $date *>&1 | Out-File -Append "logs\daily_$logDate.log"

$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Write-Output "[$timestamp] 分析完成" | Out-File -Append "logs\daily_$logDate.log"

# 检查执行结果
if ($LASTEXITCODE -ne 0) {
    Write-Host "执行失败，请查看日志: logs\daily_$logDate.log" -ForegroundColor Red
    exit 1
}

Write-Host "执行成功！" -ForegroundColor Green
