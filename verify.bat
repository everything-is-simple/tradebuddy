@echo off
echo ============================================
echo TradeBuddy Technical Verification
echo ============================================
echo.

REM Check Go installation
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go environment not detected
    echo.
    echo Please install Go 1.22+
    echo Download: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo [OK] Go environment detected
go version
echo.

REM Initialize Go module
echo [1/4] Initializing Go module...
if not exist go.mod (
    go mod init tradebuddy
)
go mod tidy
if %errorlevel% neq 0 (
    echo [ERROR] Go module initialization failed
    pause
    exit /b 1
)
echo [OK] Go module initialized
echo.

REM Create directories
echo [2/4] Creating directory structure...
if not exist data mkdir data
if not exist cmd\verify mkdir cmd\verify
if not exist internal\store mkdir internal\store
if not exist internal\model mkdir internal\model
echo [OK] Directory structure created
echo.

REM Download dependencies
echo [3/4] Downloading dependencies...
echo This may take a few minutes, please wait...
go get modernc.org/sqlite
go get github.com/xuri/excelize/v2
if %errorlevel% neq 0 (
    echo [WARNING] Download may have failed, trying China proxy...
    go env -w GOPROXY=https://goproxy.cn,direct
    go mod tidy
)
echo [OK] Dependencies downloaded
echo.

REM Run verification demos
echo [4/4] Running verification demos...
echo.
echo ============================================
echo Demo 1: SQLite Read/Write Test
echo ============================================
go run cmd\verify\demo1_sqlite.go
if %errorlevel% neq 0 (
    echo [ERROR] Demo 1 failed
    pause
    exit /b 1
)
echo.
echo.

echo ============================================
echo Demo 2: Excel Parsing Test
echo ============================================
go run cmd\verify\demo2_excel.go
if %errorlevel% neq 0 (
    echo [WARNING] Demo 2 failed, may be Excel file path issue
    echo Check VERIFY.md for troubleshooting
)
echo.
echo.

echo ============================================
echo Demo 3: TDX File Reading Test
echo ============================================
go run cmd\verify\demo3_tdx.go
if %errorlevel% neq 0 (
    echo [WARNING] Demo 3 failed, TDX directory may not be found
    echo Check VERIFY.md for troubleshooting
)
echo.
echo.

echo ============================================
echo Verification Complete!
echo ============================================
echo.
echo At least Demo 1 must pass to continue development
echo Demo 2 and Demo 3 can be fixed later
echo.
echo For details, see: docs\DESIGN.md and VERIFY.md
echo.
pause
