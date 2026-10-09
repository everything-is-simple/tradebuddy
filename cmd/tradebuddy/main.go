package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()

	// 子命令
	command := os.Args[1]
	switch command {
	case "screen":
		runScreen(ctx, os.Args[2:])
	case "lifecycle":
		runLifecycle(ctx, os.Args[2:])
	case "tachibana":
		runTachibana(ctx, os.Args[2:])
	case "run":
		runAll(ctx, os.Args[2:])
	case "report":
		runReport(ctx, os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("tradebuddy version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`TradeBuddy - 个人量化交易系统

用法:
  tradebuddy <command> [options]

命令:
  screen      运行强势股初选
  lifecycle   运行生命周期分析
  tachibana   运行立花提示分析
  run         运行完整流程 (screen → lifecycle → tachibana)
  report      查看历史报告
  version     显示版本信息
  help        显示帮助信息

示例:
  tradebuddy screen --date 2026-06-15
  tradebuddy lifecycle --date 2026-06-15
  tradebuddy tachibana --date 2026-06-15
  tradebuddy run --date 2026-06-15
  tradebuddy report --date 2026-06-15 --type lifecycle

全局选项:
  --db PATH        数据库路径 (默认: data/tradebuddy.db)
  --tdx PATH       通达信数据根目录 (默认: H:\new_tdx64\vipdoc)
  --output PATH    报告输出目录 (默认: reports/)

详细帮助:
  tradebuddy <command> --help
`)
}

// 通用配置
type CommonConfig struct {
	Date       string
	DBPath     string
	TDXRoot    string
	OutputPath string
}

func parseCommonFlags(fs *flag.FlagSet, args []string) (*CommonConfig, error) {
	cfg := &CommonConfig{}

	// 默认日期为今天
	today := time.Now().Format("2006-01-02")

	fs.StringVar(&cfg.Date, "date", today, "交易日期 (YYYY-MM-DD)")
	fs.StringVar(&cfg.DBPath, "db", "data/tradebuddy.db", "数据库路径")
	fs.StringVar(&cfg.TDXRoot, "tdx", `H:\new_tdx64\vipdoc`, "通达信数据根目录")
	fs.StringVar(&cfg.OutputPath, "output", "reports", "报告输出目录")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// 验证日期格式
	if _, err := time.Parse("2006-01-02", cfg.Date); err != nil {
		return nil, fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD 格式: %w", err)
	}

	return cfg, nil
}

// 确保目录存在
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	return nil
}
