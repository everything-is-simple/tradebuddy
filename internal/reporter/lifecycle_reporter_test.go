package reporter

import (
	"os"
	"testing"
	"tradebuddy/internal/model"
)

func TestGenerateLifecycleReport(t *testing.T) {
	reporter := NewLifecycleReporter()

	// 创建测试数据
	metrics := []*model.LifecycleMetrics{
		{
			Code:           "sh600519",
			Name:           "贵州茅台",
			Grade:          "A",
			LifecycleScore: 85.5,
			SpanDays:       15,
			PriceRangePct:  25.3,
			ATRNormalized:  3.2,
			DistFrom20DH:   5.5,
			DistFrom52WH:   -15.2,
			SpanRank:       0.8,
			RangeRank:      0.9,
			ATRRank:        0.85,
		},
		{
			Code:           "sz000001",
			Name:           "平安银行",
			Grade:          "B",
			LifecycleScore: 65.2,
			SpanDays:       10,
			PriceRangePct:  18.5,
			ATRNormalized:  2.5,
			DistFrom20DH:   3.2,
			DistFrom52WH:   -22.1,
			SpanRank:       0.6,
			RangeRank:      0.7,
			ATRRank:        0.65,
		},
	}

	// 生成报告
	outputPath := "test_lifecycle_report.xlsx"
	defer os.Remove(outputPath)

	err := reporter.GenerateLifecycleReport("2026-06-15", metrics, outputPath)
	if err != nil {
		t.Fatalf("GenerateLifecycleReport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	t.Logf("✓ Lifecycle report generated: %s", outputPath)
	t.Logf("  Records: %d", len(metrics))
}

func TestGenerateEmptyLifecycleReport(t *testing.T) {
	reporter := NewLifecycleReporter()

	// 空数据
	metrics := []*model.LifecycleMetrics{}

	outputPath := "test_empty_lifecycle_report.xlsx"
	defer os.Remove(outputPath)

	err := reporter.GenerateLifecycleReport("2026-06-15", metrics, outputPath)
	if err != nil {
		t.Fatalf("GenerateLifecycleReport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	t.Log("✓ Empty lifecycle report generated successfully")
}
