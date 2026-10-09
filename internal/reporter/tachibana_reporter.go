package reporter

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"tradebuddy/internal/model"
)

// TachibanaReporter 立花提示报告生成器
type TachibanaReporter struct {
}

// NewTachibanaReporter 创建报告生成器
func NewTachibanaReporter() *TachibanaReporter {
	return &TachibanaReporter{}
}

// GenerateTachibanaReport 生成立花提示报告Excel
func (r *TachibanaReporter) GenerateTachibanaReport(tradeDate string, signals []*model.TachibanaSignal, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	// 创建主Sheet
	sheetName := "立花交易提示"
	f.SetSheetName("Sheet1", sheetName)

	// 设置标题行
	headers := []string{
		"代码", "名称", "信号类型", "置信度",
		"当前价", "买入区间下限", "买入区间上限", "止损位",
		"评级", "评分", "波段天数", "距20日高%", "ATR标准化",
		"提示", "说明", "风险", "建议",
	}

	// 写入标题
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置标题样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellStyle(sheetName, cell, cell, titleStyle)
	}

	// 写入数据
	for i, sig := range signals {
		row := i + 2

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), sig.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sig.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), getSignalTypeName(sig.SignalType))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), sig.Confidence)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%.2f", sig.CurrentPrice))

		if sig.EntryZoneLow > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.2f", sig.EntryZoneLow))
		}
		if sig.EntryZoneHigh > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", sig.EntryZoneHigh))
		}
		if sig.StopLoss > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("%.2f", sig.StopLoss))
		}

		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), sig.LifecycleGrade)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("%.2f", sig.LifecycleScore))
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), sig.SpanDays)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("%.2f", sig.DistFrom20DH))
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), fmt.Sprintf("%.2f", sig.ATRNormalized))
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), sig.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), sig.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), sig.Risk)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), sig.Suggestion)

		// 根据信号类型设置行颜色
		var rowStyle *excelize.Style
		switch sig.SignalType {
		case model.SignalTrendProbeEntry:
			rowStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
			}
		case model.SignalTrendConfirmationAdd:
			rowStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#9BCD9B"}, Pattern: 1},
			}
		case model.SignalDistributionReduce:
			rowStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFEB9C"}, Pattern: 1},
			}
		case model.SignalExitOnRhythmFailure:
			rowStyle = &excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFC7CE"}, Pattern: 1},
			}
		}

		if rowStyle != nil {
			style, _ := f.NewStyle(rowStyle)
			for col := 0; col < len(headers); col++ {
				cell := fmt.Sprintf("%s%d", getColumnName(col), row)
				f.SetCellStyle(sheetName, cell, cell, style)
			}
		}
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 10)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 12)
	f.SetColWidth(sheetName, "D", "D", 10)
	f.SetColWidth(sheetName, "E", "M", 12)
	f.SetColWidth(sheetName, "N", "Q", 30)

	// 添加筛选器
	lastCol := getColumnName(len(headers) - 1)
	lastRow := len(signals) + 1
	f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)

	// 创建按类型分组的Sheet
	r.createGroupedSheets(f, signals)

	// 创建摘要Sheet
	r.createSummarySheet(f, tradeDate, signals)

	// 保存文件
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// createGroupedSheets 创建按类型分组的Sheet
func (r *TachibanaReporter) createGroupedSheets(f *excelize.File, signals []*model.TachibanaSignal) {
	// 按类型分组
	groups := make(map[string][]*model.TachibanaSignal)
	for _, sig := range signals {
		groups[sig.SignalType] = append(groups[sig.SignalType], sig)
	}

	// 为每个类型创建Sheet
	sheetOrder := []string{
		model.SignalTrendProbeEntry,
		model.SignalTrendConfirmationAdd,
		model.SignalDistributionReduce,
		model.SignalExitOnRhythmFailure,
		model.SignalWaitNoAction,
	}

	for _, signalType := range sheetOrder {
		sigs, ok := groups[signalType]
		if !ok || len(sigs) == 0 {
			continue
		}

		sheetName := getSignalTypeName(signalType)
		f.NewSheet(sheetName)

		// 简化的表头
		headers := []string{"代码", "名称", "当前价", "买入区间", "止损位", "评级", "说明", "建议"}

		for i, header := range headers {
			cell := fmt.Sprintf("%s1", getColumnName(i))
			f.SetCellValue(sheetName, cell, header)
		}

		for i, sig := range sigs {
			row := i + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), sig.Code)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sig.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.2f", sig.CurrentPrice))

			if sig.EntryZoneLow > 0 && sig.EntryZoneHigh > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", row),
					fmt.Sprintf("%.2f - %.2f", sig.EntryZoneLow, sig.EntryZoneHigh))
			}

			if sig.StopLoss > 0 {
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%.2f", sig.StopLoss))
			}

			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), sig.LifecycleGrade)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), sig.Description)
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), sig.Suggestion)
		}

		f.SetColWidth(sheetName, "A", "A", 10)
		f.SetColWidth(sheetName, "B", "B", 15)
		f.SetColWidth(sheetName, "C", "F", 12)
		f.SetColWidth(sheetName, "G", "H", 40)
	}
}

// createSummarySheet 创建摘要Sheet
func (r *TachibanaReporter) createSummarySheet(f *excelize.File, tradeDate string, signals []*model.TachibanaSignal) {
	sheetName := "摘要"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "分析日期")
	f.SetCellValue(sheetName, "B1", tradeDate)

	f.SetCellValue(sheetName, "A2", "总信号数")
	f.SetCellValue(sheetName, "B2", len(signals))

	// 统计各信号类型数量
	typeCount := make(map[string]int)
	confidenceCount := make(map[string]int)
	for _, sig := range signals {
		typeCount[sig.SignalType]++
		confidenceCount[sig.Confidence]++
	}

	row := 4
	f.SetCellValue(sheetName, "A4", "信号类型分布")
	for _, signalType := range []string{
		model.SignalTrendProbeEntry,
		model.SignalTrendConfirmationAdd,
		model.SignalDistributionReduce,
		model.SignalExitOnRhythmFailure,
		model.SignalWaitNoAction,
	} {
		row++
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "  "+getSignalTypeName(signalType))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), typeCount[signalType])
	}

	row += 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "置信度分布")
	for _, conf := range []string{model.ConfidenceHigh, model.ConfidenceMedium, model.ConfidenceLow} {
		row++
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "  "+conf)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), confidenceCount[conf])
	}
}

// getColumnName 获取列名（A, B, C, ...）
func getColumnName(index int) string {
	if index < 26 {
		return string(rune('A' + index))
	}
	return string(rune('A'+index/26-1)) + string(rune('A'+index%26))
}

// getSignalTypeName 获取信号类型中文名
func getSignalTypeName(signalType string) string {
	switch signalType {
	case model.SignalTrendProbeEntry:
		return "试探建仓"
	case model.SignalTrendConfirmationAdd:
		return "同向加码"
	case model.SignalDistributionReduce:
		return "分批减仓"
	case model.SignalExitOnRhythmFailure:
		return "节奏失败"
	case model.SignalWaitNoAction:
		return "等待观望"
	default:
		return signalType
	}
}
