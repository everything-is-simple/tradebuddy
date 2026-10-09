package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/lifecycle"
	"tradebuddy/internal/reporter"
	"tradebuddy/internal/screener"
	"tradebuddy/internal/store"
	"tradebuddy/internal/tachibana"
)

func runAll(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`用法: tradebuddy run [options]

运行完整工作流程：
  1. Screener  - 强势股初选
  2. Lifecycle - 生命周期分析
  3. Tachibana - 立花提示分析
  4. 生成3份Excel报告

选项:`)
		fs.PrintDefaults()
		fmt.Println(`
筛选条件:
  --min-gain FLOAT     最小涨幅百分比 (默认: 6.0)
  --lookback INT       回溯天数 (默认: 20)
  --dd52 FLOAT         距52周高点最大距离 (默认: 25.0)

示例:
  tradebuddy run --date 2026-06-15
  tradebuddy run --date 2026-06-15 --min-gain 8.0
`)
	}

	// 解析通用配置
	common, err := parseCommonFlags(fs, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数解析错误: %v\n", err)
		os.Exit(1)
	}

	// 筛选配置
	var minGain float64
	var lookback int
	var dd52 float64
	fs.Float64Var(&minGain, "min-gain", 6.0, "最小涨幅百分比")
	fs.IntVar(&lookback, "lookback", 20, "回溯天数")
	fs.Float64Var(&dd52, "dd52", 25.0, "距52周高点最大距离")
	fs.Parse(args)

	fmt.Printf("╔═══════════════════════════════════════════════════════╗\n")
	fmt.Printf("║          TradeBuddy - 完整工作流程                   ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("交易日期: %s\n", common.Date)
	fmt.Printf("数据库: %s\n", common.DBPath)
	fmt.Printf("TDX路径: %s\n", common.TDXRoot)
	fmt.Printf("输出目录: %s\n", common.OutputPath)
	fmt.Printf("筛选条件: 涨幅≥%.1f%%, 回溯%d天, 距52周高点≤%.1f%%\n\n", minGain, lookback, dd52)

	overallStart := time.Now()

	// 1. 初始化数据库
	if err := ensureDir(common.DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 创建数据库目录失败: %v\n", err)
		os.Exit(1)
	}

	st, err := store.New(&store.Config{Path: common.DBPath})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	// 2. 初始化数据源
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   common.TDXRoot,
		PreferTDX: true,
		EnableAPI: false,
	})

	// 3. 确保输出目录存在
	if err := os.MkdirAll(common.OutputPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	// ===== Stage 1: Screener =====
	fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ [1/3] Screener - 强势股初选                         │\n")
	fmt.Printf("└─────────────────────────────────────────────────────┘\n")

	stage1Start := time.Now()
	sc := screener.New(st, ds)
	screenCfg := &screener.Config{
		MinPctChange:   minGain,
		LookbackDays:   lookback,
		DD52Threshold:  dd52,
		BatchSize:      100,
		EnableParallel: false,
		MaxWorkers:     4,
	}

	lastProgress := 0
	screenResults, err := sc.RunWithProgress(ctx, common.Date, screenCfg, func(current, total int, elapsed time.Duration) {
		pct := current * 100 / total
		if pct >= lastProgress+10 || current == total {
			fmt.Printf("  进度: %d/%d (%.1f%%) - 耗时: %v\n", current, total, float64(current)*100/float64(total), elapsed.Round(time.Second))
			lastProgress = pct
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Screener 失败: %v\n", err)
		os.Exit(1)
	}

	stage1Elapsed := time.Since(stage1Start)

	// 生成 Screener 报告
	screenReportFile := filepath.Join(common.OutputPath, fmt.Sprintf("screen_result_%s.xlsx", formatDateForFile(common.Date)))
	screenRep := reporter.NewScreenReporter()
	if err := screenRep.GenerateScreenReport(common.Date, screenResults, screenReportFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成Screener报告失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Screener 完成: %d 只股票, 耗时: %v\n", len(screenResults), stage1Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n\n", screenReportFile)

	if len(screenResults) == 0 {
		fmt.Printf("⚠ 未筛选到符合条件的股票，流程结束\n")
		os.Exit(0)
	}

	// ===== Stage 2: Lifecycle =====
	fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ [2/3] Lifecycle - 生命周期分析                      │\n")
	fmt.Printf("└─────────────────────────────────────────────────────┘\n")

	stage2Start := time.Now()
	lc := lifecycle.New(st, ds)
	lifecycleCfg := &lifecycle.Config{
		MinHistoryDays: 60,
		MinSampleSize:  30,
		SpanWeight:     0.2,
		RangeWeight:    0.4,
		ATRWeight:      0.4,
	}

	metrics, err := lc.Run(ctx, common.Date, screenResults, lifecycleCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Lifecycle 失败: %v\n", err)
		os.Exit(1)
	}

	stage2Elapsed := time.Since(stage2Start)

	// 生成 Lifecycle 报告
	lifecycleReportFile := filepath.Join(common.OutputPath, fmt.Sprintf("lifecycle_result_%s.xlsx", formatDateForFile(common.Date)))
	lifecycleRep := reporter.NewLifecycleReporter()
	if err := lifecycleRep.GenerateLifecycleReport(common.Date, metrics, lifecycleReportFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成Lifecycle报告失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Lifecycle 完成: %d 只股票, 耗时: %v\n", len(metrics), stage2Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n\n", lifecycleReportFile)

	// ===== Stage 3: Tachibana =====
	fmt.Printf("┌─────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ [3/3] Tachibana - 立花提示分析                      │\n")
	fmt.Printf("└─────────────────────────────────────────────────────┘\n")

	stage3Start := time.Now()
	tb := tachibana.New(st, ds)

	signals, err := tb.Run(ctx, common.Date, metrics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Tachibana 失败: %v\n", err)
		os.Exit(1)
	}

	stage3Elapsed := time.Since(stage3Start)

	// 生成 Tachibana 报告
	tachibanaReportFile := filepath.Join(common.OutputPath, fmt.Sprintf("tachibana_result_%s.xlsx", formatDateForFile(common.Date)))
	tachibanaRep := reporter.NewTachibanaReporter()
	if err := tachibanaRep.GenerateTachibanaReport(common.Date, signals, tachibanaReportFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 生成Tachibana报告失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Tachibana 完成: %d 个信号, 耗时: %v\n", len(signals), stage3Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n\n", tachibanaReportFile)

	// ===== 总结 =====
	overallElapsed := time.Since(overallStart)

	fmt.Printf("╔═══════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                  工作流程完成                         ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("✓ Screener:  %d 只股票 (%v)\n", len(screenResults), stage1Elapsed.Round(time.Second))
	fmt.Printf("✓ Lifecycle: %d 只股票 (%v)\n", len(metrics), stage2Elapsed.Round(time.Second))
	fmt.Printf("✓ Tachibana: %d 个信号 (%v)\n", len(signals), stage3Elapsed.Round(time.Second))
	fmt.Printf("✓ 总耗时: %v\n\n", overallElapsed.Round(time.Second))

	fmt.Printf("报告文件:\n")
	fmt.Printf("  1. %s\n", screenReportFile)
	fmt.Printf("  2. %s\n", lifecycleReportFile)
	fmt.Printf("  3. %s\n", tachibanaReportFile)

	// 统计摘要
	if len(signals) > 0 {
		typeCount := make(map[string]int)
		for _, s := range signals {
			typeCount[s.SignalType]++
		}

		fmt.Printf("\n信号类型分布:\n")
		typeNames := map[string]string{
			"trend_probe_entry":      "试探建仓",
			"trend_confirmation_add": "同向加码",
			"distribution_reduce":    "分批减仓",
			"exit_on_rhythm_failure": "节奏失败",
			"wait_no_action":         "等待观望",
		}

		for signalType, name := range typeNames {
			count := typeCount[signalType]
			if count > 0 {
				fmt.Printf("  %s: %d 个\n", name, count)
			}
		}
	}

	fmt.Printf("\n🎉 所有任务完成！\n")
}
