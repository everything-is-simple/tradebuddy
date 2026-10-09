#!/usr/bin/env pwsh
# PowerShell version of verify script
# Usage: .\verify.ps1

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "TradeBuddy Technical Verification" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check Go installation
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] Go environment not detected" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go 1.22+" -ForegroundColor Yellow
    Write-Host "Download: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "[OK] Go environment detected" -ForegroundColor Green
go version
Write-Host ""

# Initialize Go module
Write-Host "[1/4] Initializing Go module..." -ForegroundColor Cyan
if (-not (Test-Path go.mod)) {
    go mod init tradebuddy
}
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Go module initialization failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host "[OK] Go module initialized" -ForegroundColor Green
Write-Host ""

# Create directories
Write-Host "[2/4] Creating directory structure..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path data | Out-Null
New-Item -ItemType Directory -Force -Path cmd\verify | Out-Null
New-Item -ItemType Directory -Force -Path internal\store | Out-Null
New-Item -ItemType Directory -Force -Path internal\model | Out-Null
Write-Host "[OK] Directory structure created" -ForegroundColor Green
Write-Host ""

# Download dependencies
Write-Host "[3/4] Downloading dependencies..." -ForegroundColor Cyan
Write-Host "This may take a few minutes, please wait..." -ForegroundColor Yellow
go get modernc.org/sqlite
go get github.com/xuri/excelize/v2
if ($LASTEXITCODE -ne 0) {
    Write-Host "[WARNING] Download may have failed, trying China proxy..." -ForegroundColor Yellow
    go env -w GOPROXY=https://goproxy.cn,direct
    go mod tidy
}
Write-Host "[OK] Dependencies downloaded" -ForegroundColor Green
Write-Host ""

# Run verification demos
Write-Host "[4/4] Running verification demos..." -ForegroundColor Cyan
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Demo 1: SQLite Read/Write Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
go run cmd\verify\demo1_sqlite.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Demo 1 failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host ""
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Demo 2: Excel Parsing Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
go run cmd\verify\demo2_excel.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "[WARNING] Demo 2 failed, may be Excel file path issue" -ForegroundColor Yellow
    Write-Host "Check VERIFY.md for troubleshooting" -ForegroundColor Yellow
}
Write-Host ""
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Demo 3: TDX File Reading Test" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
go run cmd\verify\demo3_tdx.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "[WARNING] Demo 3 failed, TDX directory may not be found" -ForegroundColor Yellow
    Write-Host "Check VERIFY.md for troubleshooting" -ForegroundColor Yellow
}
Write-Host ""
Write-Host ""

Write-Host "============================================" -ForegroundColor Green
Write-Host "Verification Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "At least Demo 1 must pass to continue development" -ForegroundColor Yellow
Write-Host "Demo 2 and Demo 3 can be fixed later" -ForegroundColor Yellow
Write-Host ""
Write-Host "For details, see: docs\DESIGN.md and VERIFY.md" -ForegroundColor Cyan
Write-Host ""
Read-Host "Press Enter to exit"
