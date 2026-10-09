package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/lifecycle"
	"tradebuddy/internal/reporter"
	"tradebuddy/internal/screener"
	"tradebuddy/internal/store"
	"tradebuddy/internal/tachibana"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║      TradeBuddy 端到端集成测试                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Println()

	ctx := context.Background()
	tradeDate := "2026-06-15"
	dbPath := "data/integration_test.db"
	tdxRoot := `H:\new_tdx64\vipdoc`
	outputDir := "reports/integration_test"

	overallStart := time.Now()

	// 清理旧数据
	fmt.Println("准备测试环境...")
	os.RemoveAll(dbPath)
	os.RemoveAll(outputDir)
	os.MkdirAll("data", 0755)
	os.MkdirAll(outputDir, 0755)
	fmt.Println("✓ 环境准备完成\n")

	// 1. 初始化数据库
	fmt.Println("1. 初始化数据库...")
	st, err := store.New(&store.Config{Path: dbPath})
	if err != nil {
		log.Fatalf("❌ 创建数据库失败: %v", err)
	}
	defer st.Close()

	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}
	fmt.Println("✓ 数据库初始化完成\n")

	// 2. 初始化数据源
	fmt.Println("2. 初始化数据源...")
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   tdxRoot,
		PreferTDX: true,
		EnableAPI: false,
	})

	stocks, err := ds.ListAvailableStocks()
	if err != nil {
		log.Fatalf("❌ 列出股票失败: %v", err)
	}
	fmt.Printf("✓ 数据源初始化完成，共 %d 只股票\n\n", len(stocks))

	// ===== Stage 1: Screener =====
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│ [1/3] Screener - 强势股初选                         │")
	fmt.Println("└─────────────────────────────────────────────────────┘")

	stage1Start := time.Now()
	sc := screener.New(st, ds)
	screenCfg := &screener.Config{
		MinPctChange:   6.0,
		LookbackDays:   20,
		DD52Threshold:  25.0,
		BatchSize:      100,
		EnableParallel: false,
		MaxWorkers:     4,
	}

	lastProgress := 0
	screenResults, err := sc.RunWithProgress(ctx, tradeDate, screenCfg, func(current, total int, elapsed time.Duration) {
		pct := current * 100 / total
		if pct >= lastProgress+10 || current == total {
			fmt.Printf("  进度: %d/%d (%.1f%%) - 耗时: %v\n",
				current, total, float64(current)*100/float64(total), elapsed.Round(time.Second))
			lastProgress = pct
		}
	})

	if err != nil {
		log.Fatalf("❌ Screener 失败: %v", err)
	}

	stage1Elapsed := time.Since(stage1Start)

	// 生成 Screener 报告
	screenReportFile := fmt.Sprintf("%s/screen_result_%s.xlsx", outputDir, formatDate(tradeDate))
	screenRep := reporter.NewScreenReporter()
	if err := screenRep.GenerateScreenReport(tradeDate, screenResults, screenReportFile); err != nil {
		log.Fatalf("❌ 生成Screener报告失败: %v", err)
	}

	fmt.Printf("\n✓ Screener 完成: %d 只股票, 耗时: %v\n", len(screenResults), stage1Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n", screenReportFile)

	// 验证结果
	if len(screenResults) == 0 {
		log.Fatalf("❌ 验证失败: 未筛选到任何股票")
	}
	fmt.Printf("✓ 验证通过: 筛选到 %d 只股票\n\n", len(screenResults))

	// ===== Stage 2: Lifecycle =====
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│ [2/3] Lifecycle - 生命周期分析                      │")
	fmt.Println("└─────────────────────────────────────────────────────┘")

	stage2Start := time.Now()
	lc := lifecycle.New(st, ds)
	lifecycleCfg := &lifecycle.Config{
		MinHistoryDays: 60,
		MinSampleSize:  30,
		SpanWeight:     0.2,
		RangeWeight:    0.4,
		ATRWeight:      0.4,
	}

	metrics, err := lc.Run(ctx, tradeDate, screenResults, lifecycleCfg)
	if err != nil {
		log.Fatalf("❌ Lifecycle 失败: %v", err)
	}

	stage2Elapsed := time.Since(stage2Start)

	// 生成 Lifecycle 报告
	lifecycleReportFile := fmt.Sprintf("%s/lifecycle_result_%s.xlsx", outputDir, formatDate(tradeDate))
	lifecycleRep := reporter.NewLifecycleReporter()
	if err := lifecycleRep.GenerateLifecycleReport(tradeDate, metrics, lifecycleReportFile); err != nil {
		log.Fatalf("❌ 生成Lifecycle报告失败: %v", err)
	}

	fmt.Printf("\n✓ Lifecycle 完成: %d 只股票, 耗时: %v\n", len(metrics), stage2Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n", lifecycleReportFile)

	// 验证结果
	if len(metrics) == 0 {
		log.Fatalf("❌ 验证失败: 未生成任何生命周期指标")
	}

	gradeCount := make(map[string]int)
	for _, m := range metrics {
		gradeCount[m.Grade]++
	}

	fmt.Printf("✓ 验证通过: 分析了 %d 只股票\n", len(metrics))
	fmt.Printf("  评级分布: A=%d, B=%d, C=%d, D=%d\n\n",
		gradeCount["A"], gradeCount["B"], gradeCount["C"], gradeCount["D"])

	// ===== Stage 3: Tachibana =====
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│ [3/3] Tachibana - 立花提示分析                      │")
	fmt.Println("└─────────────────────────────────────────────────────┘")

	stage3Start := time.Now()
	tb := tachibana.New(st, ds)

	signals, err := tb.Run(ctx, tradeDate, metrics)
	if err != nil {
		log.Fatalf("❌ Tachibana 失败: %v", err)
	}

	stage3Elapsed := time.Since(stage3Start)

	// 生成 Tachibana 报告
	tachibanaReportFile := fmt.Sprintf("%s/tachibana_result_%s.xlsx", outputDir, formatDate(tradeDate))
	tachibanaRep := reporter.NewTachibanaReporter()
	if err := tachibanaRep.GenerateTachibanaReport(tradeDate, signals, tachibanaReportFile); err != nil {
		log.Fatalf("❌ 生成Tachibana报告失败: %v", err)
	}

	fmt.Printf("\n✓ Tachibana 完成: %d 个信号, 耗时: %v\n", len(signals), stage3Elapsed.Round(time.Second))
	fmt.Printf("✓ 报告已保存: %s\n", tachibanaReportFile)

	// 验证结果
	if len(signals) == 0 {
		log.Fatalf("❌ 验证失败: 未生成任何交易信号")
	}

	typeCount := make(map[string]int)
	for _, s := range signals {
		typeCount[s.SignalType]++
	}

	fmt.Printf("✓ 验证通过: 生成了 %d 个信号\n", len(signals))
	fmt.Printf("  信号分布: 试探建仓=%d, 同向加码=%d, 分批减仓=%d, 节奏失败=%d, 等待观望=%d\n\n",
		typeCount["trend_probe_entry"],
		typeCount["trend_confirmation_add"],
		typeCount["distribution_reduce"],
		typeCount["exit_on_rhythm_failure"],
		typeCount["wait_no_action"])

	// ===== 总结 =====
	overallElapsed := time.Since(overallStart)

	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║            集成测试全部通过！✓                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("阶段耗时:\n")
	fmt.Printf("  Screener:  %v\n", stage1Elapsed.Round(time.Second))
	fmt.Printf("  Lifecycle: %v\n", stage2Elapsed.Round(time.Second))
	fmt.Printf("  Tachibana: %v\n", stage3Elapsed.Round(time.Second))
	fmt.Printf("  总耗时:    %v\n\n", overallElapsed.Round(time.Second))

	fmt.Printf("输出文件:\n")
	fmt.Printf("  数据库: %s\n", dbPath)
	fmt.Printf("  报告目录: %s\n", outputDir)
	fmt.Printf("    - screen_result_%s.xlsx\n", formatDate(tradeDate))
	fmt.Printf("    - lifecycle_result_%s.xlsx\n", formatDate(tradeDate))
	fmt.Printf("    - tachibana_result_%s.xlsx\n\n", formatDate(tradeDate))

	fmt.Println("🎉 端到端测试成功！系统可以正常运行。")
}

func formatDate(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("20060102")
}
