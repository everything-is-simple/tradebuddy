package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/reporter"
	"tradebuddy/internal/store"
	"tradebuddy/internal/tachibana"
)

func runTachibana(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("tachibana", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`用法: tradebuddy tachibana [options]

运行立花提示分析引擎，生成交易建议信号。

选项:`)
		fs.PrintDefaults()
		fmt.Println(`
示例:
  tradebuddy tachibana --date 2026-06-15
`)
	}

	// 解析通用配置
	common, err := parseCommonFlags(fs, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数解析错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== TradeBuddy Tachibana ===\n")
	fmt.Printf("交易日期: %s\n", common.Date)
	fmt.Printf("数据库: %s\n\n", common.DBPath)

	startTime := time.Now()

	// 1. 初始化数据库
	if err := ensureDir(common.DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "创建数据库目录失败: %v\n", err)
		os.Exit(1)
	}

	st, err := store.New(&store.Config{Path: common.DBPath})
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	// 2. 读取 Lifecycle 结果
	fmt.Println("读取生命周期分析结果...")
	metrics, err := st.GetLifecycleMetrics(ctx, common.Date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取生命周期结果失败: %v\n", err)
		os.Exit(1)
	}

	if len(metrics) == 0 {
		fmt.Printf("警告: 未找到 %s 的生命周期分析结果，请先运行 'tradebuddy lifecycle'\n", common.Date)
		os.Exit(1)
	}

	fmt.Printf("找到生命周期结果: %d 只股票\n\n", len(metrics))

	// 3. 初始化数据源
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   common.TDXRoot,
		PreferTDX: true,
		EnableAPI: false,
	})

	// 4. 创建 Tachibana 引擎
	tb := tachibana.New(st, ds)

	// 5. 运行分析
	fmt.Println("开始立花提示分析...")
	signals, err := tb.Run(ctx, common.Date, metrics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "分析失败: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)

	// 6. 生成报告
	if err := os.MkdirAll(common.OutputPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	reportFile := filepath.Join(common.OutputPath, fmt.Sprintf("tachibana_result_%s.xlsx", formatDateForFile(common.Date)))
	rep := reporter.NewTachibanaReporter()
	if err := rep.GenerateTachibanaReport(common.Date, signals, reportFile); err != nil {
		fmt.Fprintf(os.Stderr, "生成报告失败: %v\n", err)
		os.Exit(1)
	}

	// 7. 输出结果
	fmt.Printf("\n=== 分析完成 ===\n")
	fmt.Printf("生成信号: %d 个\n", len(signals))
	fmt.Printf("总耗时: %v\n", elapsed.Round(time.Second))
	fmt.Printf("报告已保存: %s\n", reportFile)

	// 统计信号类型分布
	if len(signals) > 0 {
		typeCount := make(map[string]int)
		for _, s := range signals {
			typeCount[s.SignalType]++
		}

		fmt.Printf("\n信号类型分布:\n")
		typeNames := map[string]string{
			"trend_probe_entry":       "试探建仓",
			"trend_confirmation_add":  "同向加码",
			"distribution_reduce":     "分批减仓",
			"exit_on_rhythm_failure":  "节奏失败",
			"wait_no_action":          "等待观望",
		}

		for signalType, name := range typeNames {
			count := typeCount[signalType]
			if count > 0 {
				fmt.Printf("  %s: %d 个 (%.1f%%)\n", name, count, float64(count)*100/float64(len(signals)))
			}
		}

		// 显示值得关注的信号
		fmt.Printf("\n值得关注的信号:\n")
		count := 0
		for _, s := range signals {
			if s.SignalType == "trend_probe_entry" || s.SignalType == "trend_confirmation_add" {
				count++
				if count <= 5 {
					actionName := typeNames[s.SignalType]
					fmt.Printf("  %d. %s (%s) - %s\n", count, s.Code, s.Name, actionName)
					fmt.Printf("     价格区间: %.2f - %.2f, 止损: %.2f\n", s.EntryZoneLow, s.EntryZoneHigh, s.StopLoss)
				}
			}
		}
		if count > 5 {
			fmt.Printf("  ... 还有 %d 个信号\n", count-5)
		}
	}
}
