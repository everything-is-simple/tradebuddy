package reporter

import (
	"os"
	"testing"
	"tradebuddy/internal/model"
)

func TestGenerateTachibanaReport(t *testing.T) {
	reporter := NewTachibanaReporter()

	// 创建测试数据
	signals := []*model.TachibanaSignal{
		{
			Code:           "sh600519",
			Name:           "贵州茅台",
			SignalType:     model.SignalTrendProbeEntry,
			Confidence:     model.ConfidenceHigh,
			CurrentPrice:   1800.0,
			EntryZoneLow:   1760.0,
			EntryZoneHigh:  1790.0,
			StopLoss:       1720.0,
			LifecycleGrade: "A",
			LifecycleScore: 85.5,
			SpanDays:       15,
			DistFrom20DH:   -5.2,
			ATRNormalized:  3.2,
			Title:          "【试探建仓机会】",
			Description:    "波段处于健康推进状态",
			Risk:           "不要重仓，分批建仓",
			Suggestion:     "小仓试探，不追高",
		},
		{
			Code:           "sz000001",
			Name:           "平安银行",
			SignalType:     model.SignalWaitNoAction,
			Confidence:     model.ConfidenceLow,
			CurrentPrice:   12.50,
			LifecycleGrade: "C",
			LifecycleScore: 45.2,
			SpanDays:       8,
			DistFrom20DH:   -2.5,
			ATRNormalized:  1.8,
			Title:          "【等待观望】",
			Description:    "结构位置不够清晰",
			Risk:           "不要勉强交易",
			Suggestion:     "继续观察，不急于行动",
		},
		{
			Code:           "sh600036",
			Name:           "招商银行",
			SignalType:     model.SignalTrendConfirmationAdd,
			Confidence:     model.ConfidenceHigh,
			CurrentPrice:   45.80,
			EntryZoneLow:   46.50,
			EntryZoneHigh:  47.20,
			StopLoss:       43.50,
			LifecycleGrade: "A",
			LifecycleScore: 88.3,
			SpanDays:       12,
			DistFrom20DH:   2.5,
			ATRNormalized:  3.5,
			Title:          "【同向加码机会】",
			Description:    "波段持续推进，创近期新高",
			Risk:           "保持分批纪律",
			Suggestion:     "等待突破确认后加码",
		},
	}

	// 生成报告
	outputPath := "test_tachibana_report.xlsx"
	defer os.Remove(outputPath)

	err := reporter.GenerateTachibanaReport("2026-06-15", signals, outputPath)
	if err != nil {
		t.Fatalf("GenerateTachibanaReport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	t.Logf("✓ Tachibana report generated: %s", outputPath)
	t.Logf("  Signals: %d", len(signals))
	t.Logf("  - 试探建仓: 1")
	t.Logf("  - 同向加码: 1")
	t.Logf("  - 等待观望: 1")
}
