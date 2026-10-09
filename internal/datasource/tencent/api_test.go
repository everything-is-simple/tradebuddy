package tencent

import (
	"context"
	"testing"
	"time"
)

func TestNewAPI(t *testing.T) {
	api := NewAPI(nil)
	if api == nil {
		t.Fatal("NewAPI returned nil")
	}

	if api.client == nil {
		t.Error("client is nil")
	}

	if api.rateLimiter == nil {
		t.Error("rateLimiter is nil")
	}
}

func TestGetQFQKLine(t *testing.T) {
	api := NewAPI(&Config{
		RequestInterval: 100 * time.Millisecond,
		Timeout:         10 * time.Second,
	})

	ctx := context.Background()

	// 测试获取贵州茅台日K线
	t.Run("GetMaotaiDaily", func(t *testing.T) {
		bars, err := api.GetQFQKLine(ctx, "sh600519", "day", 10)
		if err != nil {
			t.Fatalf("failed to get kline: %v", err)
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

		if bar.Volume <= 0 {
			t.Errorf("invalid volume: %d", bar.Volume)
		}

		if bar.Source != "tencent" {
			t.Errorf("expected source tencent, got %s", bar.Source)
		}

		t.Logf("✓ Got %d bars for sh600519", len(bars))
		t.Logf("  Latest: %s O=%.2f H=%.2f L=%.2f C=%.2f V=%d",
			bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	})

	// 测试获取平安银行
	t.Run("GetPinganDaily", func(t *testing.T) {
		bars, err := api.GetQFQKLine(ctx, "sz000001", "day", 5)
		if err != nil {
			t.Fatalf("failed to get kline: %v", err)
		}

		if len(bars) == 0 {
			t.Error("expected bars, got 0")
		}

		t.Logf("✓ Got %d bars for sz000001", len(bars))
	})
}

func TestGetMultipleQFQKLine(t *testing.T) {
	api := NewAPI(nil)
	ctx := context.Background()

	codes := []string{"sh600519", "sz000001", "sz300750"}

	result, err := api.GetMultipleQFQKLine(ctx, codes, "day", 5)
	if err != nil {
		t.Fatalf("failed to get multiple klines: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected results, got 0")
	}

	for code, bars := range result {
		t.Logf("✓ Code: %s, Bars: %d", code, len(bars))
		if len(bars) > 0 {
			t.Logf("  Latest: %s C=%.2f V=%d",
				bars[0].Date, bars[0].Close, bars[0].Volume)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	start := time.Now()

	// 第一次请求应该立即通过
	limiter.Wait()
	elapsed1 := time.Since(start)
	if elapsed1 > 10*time.Millisecond {
		t.Errorf("first request should be immediate, took %v", elapsed1)
	}

	// 第二次请求应该等待约100ms
	limiter.Wait()
	elapsed2 := time.Since(start)
	if elapsed2 < 90*time.Millisecond || elapsed2 > 120*time.Millisecond {
		t.Errorf("second request should wait ~100ms, took %v", elapsed2)
	}

	t.Logf("✓ Rate limiter working: 1st=%v, 2nd=%v", elapsed1, elapsed2)
}

func TestConvertToStringArray(t *testing.T) {
	data := []interface{}{
		[]interface{}{"2024-10-08", "45.60", "46.20", "46.50", "45.30", "1234567"},
		[]interface{}{"2024-10-07", "44.80", "45.50", "45.80", "44.50", "1111111"},
	}

	result := convertToStringArray(data)

	if len(result) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result))
	}

	if len(result[0]) != 6 {
		t.Errorf("expected 6 columns, got %d", len(result[0]))
	}

	if result[0][0] != "2024-10-08" {
		t.Errorf("expected date 2024-10-08, got %s", result[0][0])
	}
}

func TestParseBarRow(t *testing.T) {
	row := []string{"2024-10-08", "45.60", "46.20", "46.50", "45.30", "1234567", "56000000"}

	bar, err := parseBarRow("sh600519", row)
	if err != nil {
		t.Fatalf("failed to parse bar: %v", err)
	}

	if bar.Date != "2024-10-08" {
		t.Errorf("expected date 2024-10-08, got %s", bar.Date)
	}

	if bar.Open != 45.60 {
		t.Errorf("expected open 45.60, got %.2f", bar.Open)
	}

	if bar.Close != 46.20 {
		t.Errorf("expected close 46.20, got %.2f", bar.Close)
	}

	if bar.Volume != 1234567 {
		t.Errorf("expected volume 1234567, got %d", bar.Volume)
	}
}
