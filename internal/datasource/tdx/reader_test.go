package tdx

import (
	"context"
	"os"
	"testing"
)

func TestNewReader(t *testing.T) {
	reader := NewReader("H:\\new_tdx64\\vipdoc")
	if reader == nil {
		t.Fatal("NewReader returned nil")
	}

	if reader.rootPath == "" {
		t.Error("rootPath is empty")
	}
}

func TestBuildFilePath(t *testing.T) {
	reader := NewReader("H:\\new_tdx64\\vipdoc")

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "上海股票",
			code:     "sh600519",
			expected: "H:\\new_tdx64\\vipdoc\\sh\\lday\\sh600519.day",
		},
		{
			name:     "深圳股票",
			code:     "sz000001",
			expected: "H:\\new_tdx64\\vipdoc\\sz\\lday\\sz000001.day",
		},
		{
			name:     "大写代码",
			code:     "SH600519",
			expected: "H:\\new_tdx64\\vipdoc\\sh\\lday\\sh600519.day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := reader.buildFilePath(tt.code)
			if path != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, path)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		date     uint32
		expected string
	}{
		{20241008, "2024-10-08"},
		{20230115, "2023-01-15"},
		{20220701, "2022-07-01"},
	}

	for _, tt := range tests {
		result := formatDate(tt.date)
		if result != tt.expected {
			t.Errorf("formatDate(%d) = %s, want %s", tt.date, result, tt.expected)
		}
	}
}

// TestReadDayFile 测试读取.day文件
// 注意：这个测试需要实际的TDX数据文件
func TestReadDayFile(t *testing.T) {
	// 尝试两个可能的TDX路径
	tdxPaths := []string{
		"H:\\new_tdx64\\vipdoc",
		"H:\\2025炒股手训练软件\\stockdata",
	}

	var reader *Reader
	var validPath string

	// 找到第一个有效的路径
	for _, path := range tdxPaths {
		if _, err := os.Stat(path); err == nil {
			reader = NewReader(path)
			validPath = path
			break
		}
	}

	if reader == nil {
		t.Skip("TDX data directory not found, skipping test")
		return
	}

	t.Logf("Using TDX path: %s", validPath)

	ctx := context.Background()

	// 测试读取贵州茅台
	t.Run("ReadMaotai", func(t *testing.T) {
		if !reader.FileExists("sh600519") {
			t.Skip("sh600519.day not found")
			return
		}

		bars, err := reader.ReadDayFile(ctx, "sh600519")
		if err != nil {
			t.Fatalf("failed to read day file: %v", err)
		}

		if len(bars) == 0 {
			t.Error("expected bars, got 0")
		}

		// 验证第一条数据
		bar := bars[0]
		if bar.Code != "sh600519" {
			t.Errorf("expected code sh600519, got %s", bar.Code)
		}

		if bar.Date == "" {
			t.Error("date is empty")
		}

		if bar.Open <= 0 || bar.High <= 0 || bar.Low <= 0 || bar.Close <= 0 {
			t.Errorf("invalid prices: open=%.2f, high=%.2f, low=%.2f, close=%.2f",
				bar.Open, bar.High, bar.Low, bar.Close)
		}

		if bar.Source != "tdx" {
			t.Errorf("expected source tdx, got %s", bar.Source)
		}

		t.Logf("✓ Read %d bars from sh600519.day", len(bars))
		t.Logf("  First: %s O=%.2f H=%.2f L=%.2f C=%.2f V=%d",
			bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)

		// 验证最后一条数据
		if len(bars) > 1 {
			lastBar := bars[len(bars)-1]
			t.Logf("  Last:  %s O=%.2f H=%.2f L=%.2f C=%.2f V=%d",
				lastBar.Date, lastBar.Open, lastBar.High, lastBar.Low, lastBar.Close, lastBar.Volume)
		}
	})

	// 测试读取平安银行
	t.Run("ReadPingan", func(t *testing.T) {
		if !reader.FileExists("sz000001") {
			t.Skip("sz000001.day not found")
			return
		}

		bars, err := reader.ReadDayFile(ctx, "sz000001")
		if err != nil {
			t.Fatalf("failed to read day file: %v", err)
		}

		if len(bars) == 0 {
			t.Error("expected bars, got 0")
		}

		t.Logf("✓ Read %d bars from sz000001.day", len(bars))
	})
}

func TestListAllDayFiles(t *testing.T) {
	// 尝试两个可能的TDX路径
	tdxPaths := []string{
		"H:\\new_tdx64\\vipdoc",
		"H:\\2025炒股手训练软件\\stockdata",
	}

	var reader *Reader
	for _, path := range tdxPaths {
		if _, err := os.Stat(path); err == nil {
			reader = NewReader(path)
			break
		}
	}

	if reader == nil {
		t.Skip("TDX data directory not found, skipping test")
		return
	}

	files, err := reader.ListAllDayFiles()
	if err != nil {
		t.Fatalf("failed to list day files: %v", err)
	}

	t.Logf("✓ Found %d .day files", len(files))

	// 验证文件列表
	if len(files) > 0 {
		t.Logf("  Sample files: %v", files[:min(5, len(files))])
	}
}

func TestGetFileModTime(t *testing.T) {
	// 尝试两个可能的TDX路径
	tdxPaths := []string{
		"H:\\new_tdx64\\vipdoc",
		"H:\\2025炒股手训练软件\\stockdata",
	}

	var reader *Reader
	for _, path := range tdxPaths {
		if _, err := os.Stat(path); err == nil {
			reader = NewReader(path)
			break
		}
	}

	if reader == nil {
		t.Skip("TDX data directory not found, skipping test")
		return
	}

	if !reader.FileExists("sh600519") {
		t.Skip("sh600519.day not found")
		return
	}

	modTime, err := reader.GetFileModTime("sh600519")
	if err != nil {
		t.Fatalf("failed to get file mod time: %v", err)
	}

	if modTime.IsZero() {
		t.Error("mod time is zero")
	}

	t.Logf("✓ sh600519.day last modified: %s", modTime.Format("2006-01-02 15:04:05"))
}

func TestReadMultipleDayFiles(t *testing.T) {
	// 尝试两个可能的TDX路径
	tdxPaths := []string{
		"H:\\new_tdx64\\vipdoc",
		"H:\\2025炒股手训练软件\\stockdata",
	}

	var reader *Reader
	for _, path := range tdxPaths {
		if _, err := os.Stat(path); err == nil {
			reader = NewReader(path)
			break
		}
	}

	if reader == nil {
		t.Skip("TDX data directory not found, skipping test")
		return
	}

	ctx := context.Background()
	codes := []string{"sh600519", "sz000001", "sz300750"}

	result, err := reader.ReadMultipleDayFiles(ctx, codes)
	if err != nil {
		t.Fatalf("failed to read multiple day files: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected results, got 0")
	}

	for code, bars := range result {
		t.Logf("✓ Code: %s, Bars: %d", code, len(bars))
		if len(bars) > 0 {
			t.Logf("  Latest: %s C=%.2f V=%d",
				bars[len(bars)-1].Date, bars[len(bars)-1].Close, bars[len(bars)-1].Volume)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
