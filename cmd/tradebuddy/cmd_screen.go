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
	"tradebuddy/internal/screener"
	"tradebuddy/internal/store"
)

func runScreen(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("screen", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`用法: tradebuddy screen [options]

运行强势股初选引擎，筛选符合条件的股票。

选项:`)
		fs.PrintDefaults()
		fmt.Println(`
筛选条件:
  --min-gain FLOAT     最小涨幅百分比 (默认: 6.0)
  --lookback INT       回溯天数 (默认: 20)
  --dd52 FLOAT         距52周高点最大距离 (默认: 25.0)

示例:
  tradebuddy screen --date 2026-06-15
  tradebuddy screen --date 2026-06-15 --min-gain 8.0
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

	fmt.Printf("=== TradeBuddy Screener ===\n")
	fmt.Printf("交易日期: %s\n", common.Date)
	fmt.Printf("数据库: %s\n", common.DBPath)
	fmt.Printf("TDX路径: %s\n", common.TDXRoot)
	fmt.Printf("筛选条件: 涨幅≥%.1f%%, 回溯%d天, 距52周高点≤%.1f%%\n\n", minGain, lookback, dd52)

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

	// 2. 初始化数据源
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   common.TDXRoot,
		PreferTDX: true,
		EnableAPI: false,
	})

	// 3. 创建 Screener
	sc := screener.New(st, ds)
	cfg := &screener.Config{
		MinPctChange:   minGain,
		LookbackDays:   lookback,
		DD52Threshold:  dd52,
		BatchSize:      100,
		EnableParallel: false,
		MaxWorkers:     4,
	}

	// 4. 运行筛选（带进度显示）
	fmt.Println("开始筛选...")
	lastProgress := 0
	results, err := sc.RunWithProgress(ctx, common.Date, cfg, func(current, total int, elapsed time.Duration) {
		pct := current * 100 / total
		if pct >= lastProgress+10 || current == total {
			fmt.Printf("进度: %d/%d (%.1f%%) - 耗时: %v\n", current, total, float64(current)*100/float64(total), elapsed.Round(time.Second))
			lastProgress = pct
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "筛选失败: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)

	// 5. 生成报告
	if err := os.MkdirAll(common.OutputPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	reportFile := filepath.Join(common.OutputPath, fmt.Sprintf("screen_result_%s.xlsx", formatDateForFile(common.Date)))
	rep := reporter.NewScreenReporter()
	if err := rep.GenerateScreenReport(common.Date, results, reportFile); err != nil {
		fmt.Fprintf(os.Stderr, "生成报告失败: %v\n", err)
		os.Exit(1)
	}

	// 6. 输出结果
	fmt.Printf("\n=== 筛选完成 ===\n")
	fmt.Printf("符合条件股票: %d 只\n", len(results))
	fmt.Printf("总耗时: %v\n", elapsed.Round(time.Second))
	fmt.Printf("报告已保存: %s\n", reportFile)

	if len(results) > 0 {
		fmt.Printf("\n前10只股票:\n")
		for i, r := range results {
			if i >= 10 {
				break
			}
			fmt.Printf("  %d. %s (%s) - 涨幅: %.2f%%, 距52周高点: %.2f%%\n",
				i+1, r.Code, r.Name, r.PctChange, r.DD52)
		}
	}
}

// formatDateForFile 格式化日期用于文件名 (YYYYMMDD)
func formatDateForFile(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("20060102")
}
