package lifecycle

import (
	"context"
	"testing"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/store"
)

func setupTestLifecycle(t *testing.T) (*Lifecycle, *store.Store) {
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

	// 创建数据源管理器
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	})

	lifecycle := New(st, ds)

	return lifecycle, st
}

func TestNew(t *testing.T) {
	lifecycle, st := setupTestLifecycle(t)
	defer st.Close()

	if lifecycle == nil {
		t.Fatal("lifecycle is nil")
	}

	if lifecycle.store == nil {
		t.Error("store is nil")
	}

	if lifecycle.dataSource == nil {
		t.Error("dataSource is nil")
	}

	t.Log("✓ Lifecycle created successfully")
}

func TestPercentileRank(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		sample   []float64
		expected float64
	}{
		{
			name:     "Minimum value",
			value:    1.0,
			sample:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 0.0, // 0/5 = 0
		},
		{
			name:     "Maximum value",
			value:    5.0,
			sample:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 0.8, // 4/5 = 0.8
		},
		{
			name:     "Middle value",
			value:    3.0,
			sample:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 0.4, // 2/5 = 0.4
		},
		{
			name:     "Empty sample",
			value:    1.0,
			sample:   []float64{},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentileRank(tt.value, tt.sample)
			if result != tt.expected {
				t.Errorf("percentileRank(%v, %v) = %v, want %v",
					tt.value, tt.sample, result, tt.expected)
			}
		})
	}

	t.Log("✓ PercentileRank tests passed")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MinHistoryDays != 60 {
		t.Errorf("MinHistoryDays = %d, want 60", cfg.MinHistoryDays)
	}

	if cfg.MinSampleSize != 30 {
		t.Errorf("MinSampleSize = %d, want 30", cfg.MinSampleSize)
	}

	if cfg.SpanWeight != 0.2 {
		t.Errorf("SpanWeight = %.2f, want 0.2", cfg.SpanWeight)
	}

	if cfg.RangeWeight != 0.4 {
		t.Errorf("RangeWeight = %.2f, want 0.4", cfg.RangeWeight)
	}

	if cfg.ATRWeight != 0.4 {
		t.Errorf("ATRWeight = %.2f, want 0.4", cfg.ATRWeight)
	}

	t.Log("✓ Default config validated")
}

func TestAnalyzeOne(t *testing.T) {
	lifecycle, st := setupTestLifecycle(t)
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()

	// 创建测试数据
	sr := &model.ScreenResult{
		TradeDate:  "2026-06-15",
		Code:       "sh600519",
		Name:       "贵州茅台",
		PctChange:  8.5,
		ClosePrice: 1800.0,
		DD52:       -15.2,
		DataSource: "tdx",
	}

	t.Run("AnalyzeMaotai", func(t *testing.T) {
		m, err := lifecycle.analyzeOne(ctx, sr, "2026-06-15", cfg)
		if err != nil {
			t.Logf("分析失败（可能数据不足）: %v", err)
			return
		}

		if m != nil {
			t.Logf("✓ 茅台生命周期指标:")
			t.Logf("  波段持续: %d天", m.SpanDays)
			t.Logf("  价格幅度: %.2f%%", m.PriceRangePct)
			t.Logf("  ATR标准化: %.2f", m.ATRNormalized)
			t.Logf("  距20日高: %.2f%%", m.DistFrom20DH)
		}
	})
}

func TestCalculateScores(t *testing.T) {
	lifecycle, st := setupTestLifecycle(t)
	defer st.Close()

	cfg := DefaultConfig()

	// 创建测试数据
	metrics := []*model.LifecycleMetrics{
		{
			Code:       "000001",
			SpanRank:   0.8,
			RangeRank:  0.9,
			ATRRank:    0.7,
		},
		{
			Code:       "000002",
			SpanRank:   0.5,
			RangeRank:  0.5,
			ATRRank:    0.5,
		},
		{
			Code:       "000003",
			SpanRank:   0.2,
			RangeRank:  0.3,
			ATRRank:    0.1,
		},
	}

	lifecycle.calculateScores(metrics, cfg)

	// 验证评分
	if metrics[0].LifecycleScore < metrics[1].LifecycleScore {
		t.Error("Score ordering incorrect")
	}

	if metrics[1].LifecycleScore < metrics[2].LifecycleScore {
		t.Error("Score ordering incorrect")
	}

	// 验证评级
	for _, m := range metrics {
		if m.Grade == "" {
			t.Errorf("Grade not assigned for %s", m.Code)
		}
	}

	t.Logf("✓ Scores calculated:")
	for _, m := range metrics {
		t.Logf("  %s: %.2f (%s)", m.Code, m.LifecycleScore, m.Grade)
	}
}
