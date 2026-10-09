package calc

import (
	"math"
	"testing"
	"tradebuddy/internal/model"
)

func TestMA(t *testing.T) {
	bars := []*model.DailyBar{
		{Close: 10.0},
		{Close: 11.0},
		{Close: 12.0},
		{Close: 13.0},
		{Close: 14.0},
	}

	// 测试5日均线
	ma, err := MA(bars, 5)
	if err != nil {
		t.Fatalf("MA failed: %v", err)
	}

	expected := 12.0 // (10+11+12+13+14)/5
	if math.Abs(ma[4]-expected) > 0.0001 {
		t.Errorf("MA(5) = %.2f, want %.2f", ma[4], expected)
	}

	t.Logf("✓ MA(5) = %.2f", ma[4])
}

func TestMALast(t *testing.T) {
	bars := []*model.DailyBar{
		{Close: 10.0},
		{Close: 11.0},
		{Close: 12.0},
		{Close: 13.0},
		{Close: 14.0},
	}

	ma, err := MALast(bars, 3)
	if err != nil {
		t.Fatalf("MALast failed: %v", err)
	}

	expected := 13.0 // (12+13+14)/3
	if math.Abs(ma-expected) > 0.0001 {
		t.Errorf("MALast(3) = %.2f, want %.2f", ma, expected)
	}

	t.Logf("✓ MALast(3) = %.2f", ma)
}

func TestHighestHigh(t *testing.T) {
	bars := []*model.DailyBar{
		{High: 10.0},
		{High: 15.0},
		{High: 12.0},
		{High: 18.0},
		{High: 14.0},
	}

	high, err := HighestHigh(bars, 5)
	if err != nil {
		t.Fatalf("HighestHigh failed: %v", err)
	}

	if high != 18.0 {
		t.Errorf("HighestHigh(5) = %.2f, want 18.0", high)
	}

	t.Logf("✓ HighestHigh(5) = %.2f", high)
}

func TestLowestLow(t *testing.T) {
	bars := []*model.DailyBar{
		{Low: 10.0},
		{Low: 8.0},
		{Low: 12.0},
		{Low: 9.0},
		{Low: 11.0},
	}

	low, err := LowestLow(bars, 5)
	if err != nil {
		t.Fatalf("LowestLow failed: %v", err)
	}

	if low != 8.0 {
		t.Errorf("LowestLow(5) = %.2f, want 8.0", low)
	}

	t.Logf("✓ LowestLow(5) = %.2f", low)
}

func TestPctChange(t *testing.T) {
	tests := []struct {
		current  float64
		previous float64
		expected float64
	}{
		{110.0, 100.0, 10.0},
		{95.0, 100.0, -5.0},
		{100.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		result := PctChange(tt.current, tt.previous)
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("PctChange(%.2f, %.2f) = %.2f, want %.2f",
				tt.current, tt.previous, result, tt.expected)
		}
	}

	t.Log("✓ PctChange tests passed")
}
