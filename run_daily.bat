@echo off
REM TradeBuddy 每日自动分析脚本
REM 用途: Windows 任务计划程序自动执行

cd /d I:\tradebuddy

REM 创建日志目录
if not exist "logs" mkdir logs

REM 获取当前日期 (YYYYMMDD)
set today=%date:~0,4%%date:~5,2%%date:~8,2%

REM 执行完整流程
echo [%date% %time%] 开始每日分析... >> logs\daily_%today%.log
tradebuddy.exe run >> logs\daily_%today%.log 2>&1

REM 记录完成时间
echo [%date% %time%] 分析完成 >> logs\daily_%today%.log

REM 如果失败，等待用户查看
if errorlevel 1 (
    echo 执行失败，请查看日志: logs\daily_%today%.log
    pause
)
