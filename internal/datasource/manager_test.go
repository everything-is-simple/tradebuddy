package datasource

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	cfg := &Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	}

	mgr := NewManager(cfg)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.tdxReader == nil {
		t.Error("tdxReader is nil")
	}
}

func TestManagerGetDailyBars(t *testing.T) {
	cfg := &Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	}

	mgr := NewManager(cfg)
	ctx := context.Background()

	// 测试获取贵州茅台数据
	t.Run("GetMaotaiLast20", func(t *testing.T) {
		bars, err := mgr.GetDailyBars(ctx, "sh600519", 20)
		if err != nil {
			t.Fatalf("failed to get daily bars: %v", err)
		}

		if len(bars) == 0 {
			t.Error("expected bars, got 0")
		}

		if len(bars) > 20 {
			t.Errorf("expected max 20 bars, got %d", len(bars))
		}

		t.Logf("✓ Got %d bars for sh600519", len(bars))
		t.Logf("  Latest: %s C=%.2f V=%d",
			bars[len(bars)-1].Date, bars[len(bars)-1].Close, bars[len(bars)-1].Volume)
	})

	// 测试获取平安银行数据
	t.Run("GetPinganLast50", func(t *testing.T) {
		bars, err := mgr.GetDailyBars(ctx, "sz000001", 50)
		if err != nil {
			t.Fatalf("failed to get daily bars: %v", err)
		}

		if len(bars) == 0 {
			t.Error("expected bars, got 0")
		}

		t.Logf("✓ Got %d bars for sz000001", len(bars))
	})
}

func TestManagerGetMultipleDailyBars(t *testing.T) {
	cfg := &Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	}

	mgr := NewManager(cfg)
	ctx := context.Background()

	codes := []string{"sh600519", "sz000001", "sz300750"}

	result, err := mgr.GetMultipleDailyBars(ctx, codes, 10)
	if err != nil {
		t.Fatalf("failed to get multiple bars: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected results, got 0")
	}

	for code, bars := range result {
		t.Logf("✓ Code: %s, Bars: %d", code, len(bars))
		if len(bars) > 0 {
			t.Logf("  Latest: %s C=%.2f",
				bars[len(bars)-1].Date, bars[len(bars)-1].Close)
		}
	}
}

func TestManagerListAvailableStocks(t *testing.T) {
	cfg := &Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	}

	mgr := NewManager(cfg)

	stocks, err := mgr.ListAvailableStocks()
	if err != nil {
		t.Fatalf("failed to list stocks: %v", err)
	}

	if len(stocks) == 0 {
		t.Error("expected stocks, got 0")
	}

	t.Logf("✓ Found %d available stocks", len(stocks))

	// 验证是否包含常见股票
	hasStock := func(code string) bool {
		for _, s := range stocks {
			if s == code {
				return true
			}
		}
		return false
	}

	if !hasStock("sh600519") {
		t.Error("expected sh600519 in stock list")
	}

	if !hasStock("sz000001") {
		t.Error("expected sz000001 in stock list")
	}
}

func TestManagerWithBothSources(t *testing.T) {
	cfg := &Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: true, // 启用API作为备份
	}

	mgr := NewManager(cfg)
	if mgr.tdxReader == nil {
		t.Error("tdxReader should not be nil")
	}

	if mgr.tencentAPI == nil {
		t.Error("tencentAPI should not be nil")
	}

	t.Log("✓ Manager created with both TDX and Tencent API")
}
