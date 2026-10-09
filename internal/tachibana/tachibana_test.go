package tachibana

import (
	"context"
	"testing"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/model"
	"tradebuddy/internal/store"
)

func setupTestTachibana(t *testing.T) (*Tachibana, *store.Store) {
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

	tachibana := New(st, ds)

	return tachibana, st
}

func TestNew(t *testing.T) {
	tachibana, st := setupTestTachibana(t)
	defer st.Close()

	if tachibana == nil {
		t.Fatal("tachibana is nil")
	}

	if tachibana.store == nil {
		t.Error("store is nil")
	}

	if tachibana.dataSource == nil {
		t.Error("dataSource is nil")
	}

	t.Log("✓ Tachibana created successfully")
}

func TestCheckTrendProbeEntry(t *testing.T) {
	tachibana, st := setupTestTachibana(t)
	defer st.Close()

	tests := []struct {
		name     string
		metrics  *model.LifecycleMetrics
		expected bool
	}{
		{
			name: "Valid probe entry",
			metrics: &model.LifecycleMetrics{
				Grade:         "A",
				DistFrom20DH:  -5.0,
				ATRNormalized: 3.0,
				SpanDays:      15,
			},
			expected: true,
		},
		{
			name: "Grade too low",
			metrics: &model.LifecycleMetrics{
				Grade:         "C",
				DistFrom20DH:  -5.0,
				ATRNormalized: 3.0,
				SpanDays:      15,
			},
			expected: false,
		},
		{
			name: "Deep pullback",
			metrics: &model.LifecycleMetrics{
				Grade:         "A",
				DistFrom20DH:  -15.0,
				ATRNormalized: 3.0,
				SpanDays:      15,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tachibana.checkTrendProbeEntry(tt.metrics)
			if result != tt.expected {
				t.Errorf("checkTrendProbeEntry() = %v, want %v", result, tt.expected)
			}
		})
	}

	t.Log("✓ TrendProbeEntry checks passed")
}

func TestCheckExitOnRhythmFailure(t *testing.T) {
	tachibana, st := setupTestTachibana(t)
	defer st.Close()

	tests := []struct {
		name     string
		metrics  *model.LifecycleMetrics
		expected bool
	}{
		{
			name: "Valid rhythm failure",
			metrics: &model.LifecycleMetrics{
				Grade:        "C",
				DistFrom20DH: -18.0,
				ATRRank:      0.2,
				SpanRank:     0.5,
			},
			expected: true,
		},
		{
			name: "Grade too high",
			metrics: &model.LifecycleMetrics{
				Grade:        "A",
				DistFrom20DH: -18.0,
				ATRRank:      0.2,
				SpanRank:     0.5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tachibana.checkExitOnRhythmFailure(tt.metrics)
			if result != tt.expected {
				t.Errorf("checkExitOnRhythmFailure() = %v, want %v", result, tt.expected)
			}
		})
	}

	t.Log("✓ ExitOnRhythmFailure checks passed")
}

func TestAnalyzeOne(t *testing.T) {
	tachibana, st := setupTestTachibana(t)
	defer st.Close()

	ctx := context.Background()

	// 测试试探建仓
	m1 := &model.LifecycleMetrics{
		Code:          "sh600519",
		Name:          "贵州茅台",
		Grade:         "A",
		LifecycleScore: 85.5,
		SpanDays:      15,
		DistFrom20DH:  -5.0,
		ATRNormalized: 3.2,
		ClosePrice:    1800.0,
		PriceRange:    100.0,
	}

	signal := tachibana.analyzeOne(ctx, m1, "2026-06-15")

	if signal == nil {
		t.Fatal("signal is nil")
	}

	if signal.SignalType != model.SignalTrendProbeEntry {
		t.Errorf("SignalType = %s, want %s", signal.SignalType, model.SignalTrendProbeEntry)
	}

	if signal.Confidence != model.ConfidenceHigh {
		t.Errorf("Confidence = %s, want %s", signal.Confidence, model.ConfidenceHigh)
	}

	t.Logf("✓ Signal generated: %s", signal.Title)
	t.Logf("  Entry Zone: %.2f - %.2f", signal.EntryZoneLow, signal.EntryZoneHigh)
	t.Logf("  Stop Loss: %.2f", signal.StopLoss)
	t.Logf("  Description: %s", signal.Description)
}
