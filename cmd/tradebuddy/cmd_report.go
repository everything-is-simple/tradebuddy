package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"tradebuddy/internal/store"
)

func runReport(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`用法: tradebuddy report [options]

查看历史分析结果的摘要信息。

选项:`)
		fs.PrintDefaults()
		fmt.Println(`
报告类型:
  --type screen       Screener初选结果
  --type lifecycle    Lifecycle生命周期分析结果
  --type tachibana    Tachibana立花提示结果
  --type all          所有结果 (默认)

示例:
  tradebuddy report --date 2026-06-15
  tradebuddy report --date 2026-06-15 --type lifecycle
  tradebuddy report --date 2026-06-15 --type all
`)
	}

	// 解析通用配置
	common, err := parseCommonFlags(fs, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数解析错误: %v\n", err)
		os.Exit(1)
	}

	var reportType string
	fs.StringVar(&reportType, "type", "all", "报告类型")
	fs.Parse(args)

	fmt.Printf("=== TradeBuddy Report ===\n")
	fmt.Printf("交易日期: %s\n", common.Date)
	fmt.Printf("报告类型: %s\n\n", reportType)

	// 打开数据库
	st, err := store.New(&store.Config{Path: common.DBPath})
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	// 检查Excel文件是否存在
	dateStr := formatDateForFile(common.Date)
	screenFile := filepath.Join(common.OutputPath, fmt.Sprintf("screen_result_%s.xlsx", dateStr))
	lifecycleFile := filepath.Join(common.OutputPath, fmt.Sprintf("lifecycle_result_%s.xlsx", dateStr))
	tachibanaFile := filepath.Join(common.OutputPath, fmt.Sprintf("tachibana_result_%s.xlsx", dateStr))

	// 显示 Screener 结果
	if reportType == "all" || reportType == "screen" {
		fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Screener - 强势股初选                               │\n")
		fmt.Printf("└─────────────────────────────────────────────────────┘\n")

		results, err := st.GetScreenResultsByDateRange(ctx, common.Date, common.Date)
		if err != nil {
			fmt.Printf("读取失败: %v\n\n", err)
		} else if len(results) == 0 {
			fmt.Printf("未找到数据\n\n")
		} else {
			fmt.Printf("符合条件股票: %d 只\n", len(results))

			// 统计涨幅分布
			avgGain := 0.0
			for _, r := range results {
				avgGain += r.PctChange
			}
			avgGain /= float64(len(results))

			fmt.Printf("  - 平均涨幅: %.2f%%\n", avgGain)

			// Excel文件
			if _, err := os.Stat(screenFile); err == nil {
				fmt.Printf("Excel报告: %s ✓\n\n", screenFile)
			} else {
				fmt.Printf("Excel报告: 未找到\n\n")
			}
		}
	}

	// 显示 Lifecycle 结果
	if reportType == "all" || reportType == "lifecycle" {
		fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Lifecycle - 生命周期分析                            │\n")
		fmt.Printf("└─────────────────────────────────────────────────────┘\n")

		metrics, err := st.GetLifecycleMetrics(ctx, common.Date)
		if err != nil {
			fmt.Printf("读取失败: %v\n\n", err)
		} else if len(metrics) == 0 {
			fmt.Printf("未找到数据\n\n")
		} else {
			fmt.Printf("分析股票: %d 只\n", len(metrics))

			// 统计评级分布
			gradeCount := make(map[string]int)
			for _, m := range metrics {
				gradeCount[m.Grade]++
			}

			fmt.Printf("评级分布:\n")
			for _, grade := range []string{"A", "B", "C", "D"} {
				count := gradeCount[grade]
				if count > 0 {
					fmt.Printf("  %s级: %d 只 (%.1f%%)\n", grade, count, float64(count)*100/float64(len(metrics)))
				}
			}

			// 前3名
			if len(metrics) > 0 {
				fmt.Printf("\n综合评分前3:\n")
				for i := 0; i < 3 && i < len(metrics); i++ {
					m := metrics[i]
					fmt.Printf("  %d. %s (%s) - 评分: %.2f, 评级: %s\n", i+1, m.Code, m.Name, m.LifecycleScore, m.Grade)
				}
			}

			// Excel文件
			if _, err := os.Stat(lifecycleFile); err == nil {
				fmt.Printf("\nExcel报告: %s ✓\n\n", lifecycleFile)
			} else {
				fmt.Printf("\nExcel报告: 未找到\n\n")
			}
		}
	}

	// 显示 Tachibana 结果
	if reportType == "all" || reportType == "tachibana" {
		fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Tachibana - 立花提示分析                            │\n")
		fmt.Printf("└─────────────────────────────────────────────────────┘\n")

		signals, err := st.GetTachibanaSignals(ctx, common.Date)
		if err != nil {
			fmt.Printf("读取失败: %v\n\n", err)
		} else if len(signals) == 0 {
			fmt.Printf("未找到数据\n\n")
		} else {
			fmt.Printf("生成信号: %d 个\n", len(signals))

			// 统计信号类型分布
			typeCount := make(map[string]int)
			for _, s := range signals {
				typeCount[s.SignalType]++
			}

			typeNames := map[string]string{
				"trend_probe_entry":      "试探建仓",
				"trend_confirmation_add": "同向加码",
				"distribution_reduce":    "分批减仓",
				"exit_on_rhythm_failure": "节奏失败",
				"wait_no_action":         "等待观望",
			}

			fmt.Printf("信号类型分布:\n")
			for signalType, name := range typeNames {
				count := typeCount[signalType]
				if count > 0 {
					fmt.Printf("  %s: %d 个 (%.1f%%)\n", name, count, float64(count)*100/float64(len(signals)))
				}
			}

			// 值得关注的信号
			actionableCount := typeCount["trend_probe_entry"] + typeCount["trend_confirmation_add"]
			if actionableCount > 0 {
				fmt.Printf("\n值得关注: %d 个 (试探建仓 + 同向加码)\n", actionableCount)
			}

			// Excel文件
			if _, err := os.Stat(tachibanaFile); err == nil {
				fmt.Printf("\nExcel报告: %s ✓\n\n", tachibanaFile)
			} else {
				fmt.Printf("\nExcel报告: 未找到\n\n")
			}
		}
	}
}
