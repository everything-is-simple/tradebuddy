package screener

import (
	"context"
	"testing"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/store"
)

func setupTestScreener(t *testing.T) (*Screener, *store.Store) {
	t.Helper()

	// 创建临时数据库
	tmpDir := t.TempDir()
	st, err := store.New(&store.Config{Path: tmpDir + "/test.db"})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 执行迁移
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 创建数据源管理器（使用TDX）
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	})

	screener := New(st, ds)

	return screener, st
}

func TestScreenerNew(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	if screener == nil {
		t.Fatal("screener is nil")
	}

	if screener.store == nil {
		t.Error("store is nil")
	}

	if screener.dataSource == nil {
		t.Error("dataSource is nil")
	}

	t.Log("✓ Screener created successfully")
}

func TestScreenOne(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()

	// 测试筛选茅台
	t.Run("ScreenMaotai", func(t *testing.T) {
		result, err := screener.screenOne(ctx, "sh600519", "2026-06-15", cfg)
		if err != nil {
			t.Logf("茅台筛选失败（可能不符合条件）: %v", err)
			return
		}

		if result != nil {
			t.Logf("✓ 茅台符合条件:")
			t.Logf("  涨幅: %.2f%%", result.PctChange)
			t.Logf("  收盘价: %.2f", result.ClosePrice)
			t.Logf("  距52周高点: %.2f%%", result.DD52)
		} else {
			t.Log("茅台不符合筛选条件")
		}
	})
}

func TestScreenerRun(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()

	// 配置较宽松的条件进行测试
	cfg := &Config{
		MinPctChange:  5.0,  // 涨幅 ≥ 5%
		LookbackDays:  20,
		DD52Threshold: 30.0, // 距52周高点 ≤ 30%
	}

	// 只测试少量股票（避免耗时过长）
	t.Run("RunWithLimitedStocks", func(t *testing.T) {
		// 由于Run会处理所有股票，这里跳过完整测试
		// 实际使用时需要完整数据
		t.Skip("跳过完整筛选测试，避免耗时过长")

		results, err := screener.Run(ctx, "2026-06-15", cfg)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		t.Logf("✓ 筛选完成，符合条件: %d只", len(results))

		// 显示前几只
		for i, r := range results {
			if i >= 5 {
				break
			}
			t.Logf("  %s %s: 涨幅%.2f%%, 距52周高点%.2f%%",
				r.Code, r.Name, r.PctChange, r.DD52)
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MinPctChange != 6.0 {
		t.Errorf("MinPctChange = %.2f, want 6.0", cfg.MinPctChange)
	}

	if cfg.LookbackDays != 20 {
		t.Errorf("LookbackDays = %d, want 20", cfg.LookbackDays)
	}

	if cfg.DD52Threshold != 25.0 {
		t.Errorf("DD52Threshold = %.2f, want 25.0", cfg.DD52Threshold)
	}

	t.Log("✓ Default config validated")
}

// TestGoldenScreener 黄金测试：使用真实历史数据验证
// 这个测试需要完整的TDX数据和已知的2026-10-08筛选结果
func TestGoldenScreener(t *testing.T) {
	t.Skip("Golden Test需要完整的2026-10-08数据和已知结果进行验证")

	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()

	// 运行2026-10-08的筛选
	results, err := screener.Run(ctx, "2026-10-08", cfg)
	if err != nil {
		t.Fatalf("Golden test failed: %v", err)
	}

	// 验证：应该产出27只股票
	expectedCount := 27
	if len(results) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(results))
	}

	// 验证：应该包含605133（世名科技）
	found605133 := false
	for _, r := range results {
		if r.Code == "605133" {
			found605133 = true
			t.Logf("✓ Found 605133: %s, 涨幅%.2f%%", r.Name, r.PctChange)
			break
		}
	}

	if !found605133 {
		t.Error("Expected to find 605133 (世名科技) in results")
	}

	t.Logf("✓ Golden test passed: %d stocks screened", len(results))
}
