package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"tradebuddy/internal/datasource"
	"tradebuddy/internal/screener"
	"tradebuddy/internal/store"
)

func main() {
	fmt.Println("=== TradeBuddy 完整筛选测试 ===")
	fmt.Println("使用数据：2026-06-15（TDX历史数据）")
	fmt.Println()

	// 1. 初始化数据库
	fmt.Println("1. 初始化数据库...")
	st, err := store.New(&store.Config{Path: "test_screen.db"})
	if err != nil {
		log.Fatalf("创建数据库失败: %v", err)
	}
	defer st.Close()

	// 尝试迁移，如果已存在则忽略错误
	if err := st.Migrate(context.Background()); err != nil {
		fmt.Printf("迁移警告（可能已存在）: %v\n", err)
		// 不退出，继续使用现有数据库
	}
	fmt.Println("✓ 数据库初始化完成")
	fmt.Println()

	// 2. 初始化数据源
	fmt.Println("2. 初始化数据源...")
	ds := datasource.NewManager(&datasource.Config{
		TDXRoot:   "H:\\new_tdx64\\vipdoc",
		PreferTDX: true,
		EnableAPI: false,
	})

	stocks, err := ds.ListAvailableStocks()
	if err != nil {
		log.Fatalf("列出股票失败: %v", err)
	}
	fmt.Printf("✓ 数据源初始化完成，共 %d 只股票\n", len(stocks))
	fmt.Println()

	// 3. 创建筛选引擎
	fmt.Println("3. 创建筛选引擎...")
	screen := screener.New(st, ds)
	fmt.Println("✓ 筛选引擎创建完成")
	fmt.Println()

	// 4. 配置筛选参数
	cfg := &screener.Config{
		MinPctChange:   6.0,  // 涨幅 ≥ 6%
		LookbackDays:   20,   // 20日高点
		DD52Threshold:  25.0, // 距52周高点 ≤ 25%
		BatchSize:      100,
		EnableParallel: false, // 串行处理，便于观察
		MaxWorkers:     4,
	}

	fmt.Println("4. 筛选配置:")
	fmt.Printf("   - 最小涨幅: %.1f%%\n", cfg.MinPctChange)
	fmt.Printf("   - 回溯天数: %d日\n", cfg.LookbackDays)
	fmt.Printf("   - 距52周高点阈值: %.1f%%\n", cfg.DD52Threshold)
	fmt.Println()

	// 5. 运行筛选（带进度回调）
	fmt.Println("5. 开始筛选...")
	fmt.Println("----------------------------------------")

	startTime := time.Now()
	ctx := context.Background()

	progressFn := func(current, total int, elapsed time.Duration) {
		if current%500 == 0 {
			percent := float64(current) / float64(total) * 100
			speed := float64(current) / elapsed.Seconds()
			remaining := time.Duration(float64(total-current)/speed) * time.Second
			fmt.Printf("进度: %d/%d (%.1f%%) | 速度: %.0f股/秒 | 预计剩余: %v\n",
				current, total, percent, speed, remaining.Round(time.Second))
		}
	}

	results, err := screen.RunWithProgress(ctx, "2026-06-15", cfg, progressFn)
	if err != nil {
		log.Fatalf("筛选失败: %v", err)
	}

	elapsed := time.Since(startTime)
	fmt.Println("----------------------------------------")
	fmt.Printf("✓ 筛选完成，耗时: %v\n", elapsed)
	fmt.Println()

	// 6. 显示结果
	fmt.Println("6. 筛选结果:")
	fmt.Printf("   符合条件: %d 只\n", len(results))
	fmt.Println()

	if len(results) > 0 {
		fmt.Println("   前10只股票详情:")
		fmt.Println("   ┌──────────┬────────────┬──────┬──────┬──────────┬──────────┐")
		fmt.Println("   │ 代码     │ 名称       │ 涨幅 │ 收盘 │ 成交量   │ 距52周高 │")
		fmt.Println("   ├──────────┼────────────┼──────┼──────┼──────────┼──────────┤")

		for i, r := range results {
			if i >= 10 {
				break
			}
			fmt.Printf("   │ %-8s │ %-10s │ %5.2f│ %5.2f│ %8d │ %7.2f  │\n",
				r.Code, r.Name, r.PctChange, r.ClosePrice, r.Volume/10000, r.DD52)
		}
		fmt.Println("   └──────────┴────────────┴──────┴──────┴──────────┴──────────┘")
		fmt.Println()
	}

	// 7. 生成Excel报告
	fmt.Println("7. 生成Excel报告...")
	outputPath := "screen_result_20260615.xlsx"
	err = screen.RunAndExport(ctx, "2026-06-15", cfg, outputPath)
	if err != nil {
		log.Printf("生成报告失败: %v", err)
	} else {
		fmt.Printf("✓ Excel报告已生成: %s\n", outputPath)
	}
	fmt.Println()

	// 8. 统计信息
	fmt.Println("8. 统计信息:")
	fmt.Printf("   - 总股票数: %d\n", len(stocks))
	fmt.Printf("   - 符合条件: %d\n", len(results))
	fmt.Printf("   - 筛选率: %.2f%%\n", float64(len(results))/float64(len(stocks))*100)
	fmt.Printf("   - 总耗时: %v\n", elapsed)
	fmt.Printf("   - 平均速度: %.0f 股/秒\n", float64(len(stocks))/elapsed.Seconds())
	fmt.Println()

	fmt.Println("=== 筛选测试完成 ===")
}
