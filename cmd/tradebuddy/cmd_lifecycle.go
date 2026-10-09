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
	"tradebuddy/internal/store"
)

func runLifecycle(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("lifecycle", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`用法: tradebuddy lifecycle [options]

运行生命周期分析引擎，分析股票的波段特征和强度。

选项:`)
		fs.PrintDefaults()
		fmt.Println(`
配置选项:
  --min-history INT    最少历史天数 (默认: 60)
  --min-sample INT     最小样本量 (默认: 30)

示例:
  tradebuddy lifecycle --date 2026-06-15
  tradebuddy lifecycle --date 2026-06-15 --min-sample 50
`)
	}

	// 解析通用配置
	common, err := parseCommonFlags(fs, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数解析错误: %v\n", err)
		os.Exit(1)
	}

	// Lifecycle 配置
	var minHistory int
	var minSample int
	fs.IntVar(&minHistory, "min-history", 60, "最少历史天数")
	fs.IntVar(&minSample, "min-sample", 30, "最小样本量")
	fs.Parse(args)

	fmt.Printf("=== TradeBuddy Lifecycle ===\n")
	fmt.Printf("交易日期: %s\n", common.Date)
	fmt.Printf("数据库: %s\n", common.DBPath)
	fmt.Printf("配置: 最少历史%d天, 最小样本量%d\n\n", minHistory, minSample)

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

	// 2. 读取 Screener 结果
	fmt.Println("读取初选结果...")
	screenResults, err := st.GetScreenResultsByDateRange(ctx, common.Date, common.Date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取初选结果失败: %v\n", err)
		os.Exit(1)
	}

	if len(screenResults) == 0 {
		fmt.Printf("警告: 未找到 %s 的初选结果，请先运行 'tradebuddy screen'\n", common.Date)
		os.Exit(1)
	}

	fmt.Printf("找到初选结果: %d 只股票\n\n", len(screenResults))

	// 3. 初始化数据源
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   common.TDXRoot,
		PreferTDX: true,
		EnableAPI: false,
	})

	// 4. 创建 Lifecycle 引擎
	lc := lifecycle.New(st, ds)
	cfg := &lifecycle.Config{
		MinHistoryDays: minHistory,
		MinSampleSize:  minSample,
		SpanWeight:     0.2,
		RangeWeight:    0.4,
		ATRWeight:      0.4,
	}

	// 5. 运行分析
	fmt.Println("开始生命周期分析...")
	metrics, err := lc.Run(ctx, common.Date, screenResults, cfg)
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

	reportFile := filepath.Join(common.OutputPath, fmt.Sprintf("lifecycle_result_%s.xlsx", formatDateForFile(common.Date)))
	rep := reporter.NewLifecycleReporter()
	if err := rep.GenerateLifecycleReport(common.Date, metrics, reportFile); err != nil {
		fmt.Fprintf(os.Stderr, "生成报告失败: %v\n", err)
		os.Exit(1)
	}

	// 7. 输出结果
	fmt.Printf("\n=== 分析完成 ===\n")
	fmt.Printf("分析股票: %d 只\n", len(metrics))
	fmt.Printf("总耗时: %v\n", elapsed.Round(time.Second))
	fmt.Printf("报告已保存: %s\n", reportFile)

	// 统计评级分布
	if len(metrics) > 0 {
		gradeCount := make(map[string]int)
		for _, m := range metrics {
			gradeCount[m.Grade]++
		}

		fmt.Printf("\n评级分布:\n")
		for _, grade := range []string{"A", "B", "C", "D"} {
			count := gradeCount[grade]
			if count > 0 {
				fmt.Printf("  %s级: %d 只 (%.1f%%)\n", grade, count, float64(count)*100/float64(len(metrics)))
			}
		}

		// 显示前5名
		fmt.Printf("\n综合评分前5:\n")
		for i, m := range metrics {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s (%s) - 评分: %.2f, 评级: %s\n",
				i+1, m.Code, m.Name, m.LifecycleScore, m.Grade)
		}
	}
}
