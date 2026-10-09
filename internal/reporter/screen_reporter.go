package reporter

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"tradebuddy/internal/model"
)

// ScreenReporter 筛选报告生成器
type ScreenReporter struct {
}

// NewScreenReporter 创建报告生成器
func NewScreenReporter() *ScreenReporter {
	return &ScreenReporter{}
}

// GenerateScreenReport 生成筛选报告Excel
func (r *ScreenReporter) GenerateScreenReport(tradeDate string, results []*model.ScreenResult, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	sheetName := "强势股初选"
	f.SetSheetName("Sheet1", sheetName)

	// 设置标题行
	headers := []string{
		"代码", "名称", "收盘价", "涨幅%", "最高价", "成交量", "成交额",
		"距52周高点%", "周均线", "月均线", "数据源",
	}

	// 写入标题
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置标题样式
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create title style: %w", err)
	}

	for i := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellStyle(sheetName, cell, cell, titleStyle)
	}

	// 写入数据
	for i, result := range results {
		row := i + 2

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), result.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), result.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), result.ClosePrice)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%.2f", result.PctChange))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), result.HighPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), result.Volume)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), result.Amount)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("%.2f", result.DD52))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("%.2f", result.WeeklyMA))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("%.2f", result.MonthlyMA))
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), result.DataSource)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 10)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "E", 10)
	f.SetColWidth(sheetName, "F", "G", 12)
	f.SetColWidth(sheetName, "H", "J", 12)
	f.SetColWidth(sheetName, "K", "K", 10)

	// 添加筛选器
	lastCol := string(rune('A' + len(headers) - 1))
	lastRow := len(results) + 1
	f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)

	// 添加摘要信息
	summarySheet := "摘要"
	f.NewSheet(summarySheet)

	f.SetCellValue(summarySheet, "A1", "筛选日期")
	f.SetCellValue(summarySheet, "B1", tradeDate)

	f.SetCellValue(summarySheet, "A2", "筛选数量")
	f.SetCellValue(summarySheet, "B2", len(results))

	f.SetCellValue(summarySheet, "A3", "筛选条件")
	f.SetCellValue(summarySheet, "B3", "涨幅≥6%，突破20日高点，距52周高点≤25%")

	// 保存文件
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
