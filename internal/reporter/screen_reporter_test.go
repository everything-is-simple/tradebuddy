package reporter

import (
	"os"
	"testing"
	"tradebuddy/internal/model"
)

func TestGenerateScreenReport(t *testing.T) {
	reporter := NewScreenReporter()

	// 创建测试数据
	results := []*model.ScreenResult{
		{
			TradeDate:  "2026-10-08",
			Code:       "605133",
			Name:       "世名科技",
			PctChange:  8.5,
			ClosePrice: 45.6,
			HighPrice:  46.2,
			Volume:     1234567,
			Amount:     56000000,
			DD52:       -15.2,
			WeeklyMA:   44.5,
			MonthlyMA:  43.8,
			DataSource: "tdx",
		},
		{
			TradeDate:  "2026-10-08",
			Code:       "600825",
			Name:       "新华传媒",
			PctChange:  7.2,
			ClosePrice: 12.3,
			HighPrice:  12.5,
			Volume:     2345678,
			Amount:     28000000,
			DD52:       -20.1,
			WeeklyMA:   12.0,
			MonthlyMA:  11.8,
			DataSource: "tdx",
		},
	}

	// 生成报告
	outputPath := "test_screen_report.xlsx"
	defer os.Remove(outputPath) // 清理测试文件

	err := reporter.GenerateScreenReport("2026-10-08", results, outputPath)
	if err != nil {
		t.Fatalf("GenerateScreenReport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	t.Logf("✓ Excel report generated: %s", outputPath)
	t.Logf("  Records: %d", len(results))
}

func TestGenerateEmptyReport(t *testing.T) {
	reporter := NewScreenReporter()

	// 空数据
	results := []*model.ScreenResult{}

	outputPath := "test_empty_report.xlsx"
	defer os.Remove(outputPath)

	err := reporter.GenerateScreenReport("2026-10-08", results, outputPath)
	if err != nil {
		t.Fatalf("GenerateScreenReport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	t.Log("✓ Empty report generated successfully")
}
