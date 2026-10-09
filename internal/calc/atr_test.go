package calc

import (
	"math"
	"testing"
	"tradebuddy/internal/model"
)

func TestATR(t *testing.T) {
	// 测试数据：连续3天
	bars := []*model.DailyBar{
		{High: 50.0, Low: 48.0, Close: 49.0},
		{High: 51.0, Low: 49.0, Close: 50.5},
		{High: 52.0, Low: 48.5, Close: 51.0},
	}

	atr, err := ATR(bars, 2)
	if err != nil {
		t.Fatalf("ATR failed: %v", err)
	}

	// 验证长度
	if len(atr) != len(bars) {
		t.Errorf("ATR length = %d, want %d", len(atr), len(bars))
	}

	// 第一个值应该是0（不足周期）
	if atr[0] != 0 {
		t.Errorf("ATR[0] = %.2f, want 0", atr[0])
	}

	// ATR应该大于0
	if atr[len(atr)-1] <= 0 {
		t.Errorf("ATR[last] = %.2f, want > 0", atr[len(atr)-1])
	}

	t.Logf("✓ ATR = %v", atr)
}

func TestATRLast(t *testing.T) {
	bars := []*model.DailyBar{
		{High: 50.0, Low: 48.0, Close: 49.0},
		{High: 51.0, Low: 49.0, Close: 50.5},
		{High: 52.0, Low: 48.5, Close: 51.0},
		{High: 53.0, Low: 50.0, Close: 52.0},
		{High: 54.0, Low: 51.0, Close: 53.5},
	}

	atr, err := ATRLast(bars, 3)
	if err != nil {
		t.Fatalf("ATRLast failed: %v", err)
	}

	if atr <= 0 {
		t.Errorf("ATRLast = %.2f, want > 0", atr)
	}

	t.Logf("✓ ATRLast(3) = %.2f", atr)
}

func TestATRCalculation(t *testing.T) {
	// 手工计算的测试用例
	bars := []*model.DailyBar{
		{High: 48.0, Low: 47.0, Close: 47.5}, // TR = 1.0
		{High: 49.0, Low: 47.5, Close: 48.5}, // TR = max(1.5, 1.5, 0) = 1.5
		{High: 49.5, Low: 48.0, Close: 49.0}, // TR = max(1.5, 1.0, 0.5) = 1.5
	}

	atr, err := ATR(bars, 2)
	if err != nil {
		t.Fatalf("ATR failed: %v", err)
	}

	// ATR(2) = (1.0 + 1.5) / 2 = 1.25
	expected := 1.25
	if math.Abs(atr[1]-expected) > 0.01 {
		t.Errorf("ATR[1] = %.2f, want %.2f", atr[1], expected)
	}

	t.Logf("✓ ATR calculation verified: %.2f", atr[1])
}
