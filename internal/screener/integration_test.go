package screener

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRunAndExport(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()
	cfg := &Config{
		MinPctChange:   5.0,  // 宽松条件
		LookbackDays:   20,
		DD52Threshold:  30.0,
		BatchSize:      50,
		EnableParallel: false,
	}

	outputPath := "test_screen_output.xlsx"
	defer os.Remove(outputPath)

	// 由于完整筛选耗时较长，这里跳过
	t.Skip("跳过完整筛选测试，需要完整数据和较长时间")

	err := screener.RunAndExport(ctx, "2026-06-15", cfg, outputPath)
	if err != nil {
		t.Fatalf("RunAndExport failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	t.Log("✓ Screen and export completed")
}

func TestRunWithProgress(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()

	// 进度回调
	progressCalled := false
	progressFn := func(current, total int, elapsed time.Duration) {
		progressCalled = true
		t.Logf("Progress: %d/%d (%.1f%%), elapsed: %v",
			current, total, float64(current)/float64(total)*100, elapsed)
	}

	// 由于完整筛选耗时较长，这里跳过
	t.Skip("跳过进度测试，需要完整数据")

	_, err := screener.RunWithProgress(ctx, "2026-06-15", cfg, progressFn)
	if err != nil {
		t.Fatalf("RunWithProgress failed: %v", err)
	}

	if !progressCalled {
		t.Error("Progress callback was not called")
	}
}

func TestParallelProcessing(t *testing.T) {
	screener, st := setupTestScreener(t)
	defer st.Close()

	ctx := context.Background()
	cfg := &Config{
		MinPctChange:   5.0,
		LookbackDays:   20,
		DD52Threshold:  30.0,
		BatchSize:      50,
		EnableParallel: true,  // 启用并行
		MaxWorkers:     4,
	}

	t.Skip("跳过并行处理测试，需要完整数据")

	startTime := time.Now()
	results, err := screener.Run(ctx, "2026-06-15", cfg)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Parallel run failed: %v", err)
	}

	t.Logf("✓ Parallel processing completed")
	t.Logf("  Results: %d", len(results))
	t.Logf("  Elapsed: %v", elapsed)
}

func BenchmarkScreenSerial(b *testing.B) {
	screener, st := setupTestScreener(&testing.T{})
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.EnableParallel = false

	// 只测试少量股票
	codes := []string{"sh600519", "sz000001", "sz300750"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range codes {
			screener.screenOne(ctx, code, "2026-06-15", cfg)
		}
	}
}

func BenchmarkScreenParallel(b *testing.B) {
	screener, st := setupTestScreener(&testing.T{})
	defer st.Close()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.EnableParallel = true
	cfg.MaxWorkers = 4

	codes := []string{"sh600519", "sz000001", "sz300750"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screener.runParallel(ctx, codes, "2026-06-15", cfg, nil)
	}
}
