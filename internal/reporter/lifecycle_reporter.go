package reporter

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"tradebuddy/internal/model"
)

// LifecycleReporter 生命周期报告生成器
type LifecycleReporter struct {
}

// NewLifecycleReporter 创建报告生成器
func NewLifecycleReporter() *LifecycleReporter {
	return &LifecycleReporter{}
}

// GenerateLifecycleReport 生成生命周期报告Excel
func (r *LifecycleReporter) GenerateLifecycleReport(tradeDate string, metrics []*model.LifecycleMetrics, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	sheetName := "生命周期分析"
	f.SetSheetName("Sheet1", sheetName)

	// 设置标题行
	headers := []string{
		"代码", "名称", "评级", "综合评分",
		"波段天数", "价格幅度%", "ATR标准化",
		"距20日高%", "距52周高%",
		"持续排名", "幅度排名", "ATR排名",
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
	for i, m := range metrics {
		row := i + 2

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), m.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), m.Grade)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%.2f", m.LifecycleScore))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), m.SpanDays)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.2f", m.PriceRangePct))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", m.ATRNormalized))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("%.2f", m.DistFrom20DH))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("%.2f", m.DistFrom52WH))

		if m.SpanRank >= 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("%.2f", m.SpanRank*100))
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), "N/A")
		}

		if m.RangeRank >= 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("%.2f", m.RangeRank*100))
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), "N/A")
		}

		if m.ATRRank >= 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("%.2f", m.ATRRank*100))
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "N/A")
		}

		// 根据评级设置行颜色
		var gradeStyle *excelize.Style
		switch m.Grade {
		case "A":
			gradeStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
			}
		case "B":
			gradeStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFEB9C"}, Pattern: 1},
			}
		case "C":
			gradeStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFC7CE"}, Pattern: 1},
			}
		}

		if gradeStyle != nil {
			style, _ := f.NewStyle(gradeStyle)
			for col := 0; col < len(headers); col++ {
				cell := fmt.Sprintf("%s%d", string(rune('A'+col)), row)
				f.SetCellStyle(sheetName, cell, cell, style)
			}
		}
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 10)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 8)
	f.SetColWidth(sheetName, "D", "L", 12)

	// 添加筛选器
	lastCol := string(rune('A' + len(headers) - 1))
	lastRow := len(metrics) + 1
	f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)

	// 添加摘要Sheet
	summarySheet := "摘要"
	f.NewSheet(summarySheet)

	f.SetCellValue(summarySheet, "A1", "分析日期")
	f.SetCellValue(summarySheet, "B1", tradeDate)

	f.SetCellValue(summarySheet, "A2", "总股票数")
	f.SetCellValue(summarySheet, "B2", len(metrics))

	// 统计各评级数量
	gradeCount := make(map[string]int)
	for _, m := range metrics {
		gradeCount[m.Grade]++
	}

	row := 3
	f.SetCellValue(summarySheet, "A3", "评级分布")
	for _, grade := range []string{"A", "B", "C", "D", "N/A"} {
		row++
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("  %s级", grade))
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), gradeCount[grade])
	}

	// 保存文件
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
