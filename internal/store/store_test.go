package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"tradebuddy/internal/model"
)

// setupTestStore creates a temporary test database
func setupTestStore(t *testing.T) *Store {
	t.Helper()

	// 创建临时目录
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// 创建Store
	store, err := New(&Config{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 执行迁移
	ctx := context.Background()
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return store
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := New(&Config{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// 验证连接
	if err := store.Ping(context.Background()); err != nil {
		t.Errorf("ping failed: %v", err)
	}

	// 验证数据库文件存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file not created")
	}
}

func TestMigrate(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// 验证schema版本
	version, err := store.GetSchemaVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}

	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}

	// 验证表是否存在
	tables := []string{
		"instruments",
		"bars_daily",
		"screen_results",
		"lifecycle_states",
		"tachibana_signals",
		"tracker_positions",
		"transactions",
		"reviews",
		"job_runs",
		"schema_migrations",
	}

	for _, table := range tables {
		var name string
		err := store.db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)

		if err != nil {
			t.Errorf("table %s does not exist: %v", table, err)
		}
	}
}

func TestInstrumentCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// Create
	inst := &model.Instrument{
		Code:      "sh600519",
		Name:      "贵州茅台",
		Board:     "主板",
		ListDate:  "2001-08-27",
		IsST:      false,
		IsActive:  true,
		UpdatedAt: time.Now(),
	}

	if err := store.UpsertInstrument(ctx, inst); err != nil {
		t.Fatalf("failed to insert instrument: %v", err)
	}

	// Read
	retrieved, err := store.GetInstrument(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to get instrument: %v", err)
	}

	if retrieved.Code != inst.Code {
		t.Errorf("expected code %s, got %s", inst.Code, retrieved.Code)
	}
	if retrieved.Name != inst.Name {
		t.Errorf("expected name %s, got %s", inst.Name, retrieved.Name)
	}

	// Update
	inst.Name = "贵州茅台(更新)"
	if err := store.UpsertInstrument(ctx, inst); err != nil {
		t.Fatalf("failed to update instrument: %v", err)
	}

	updated, err := store.GetInstrument(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to get updated instrument: %v", err)
	}

	if updated.Name != "贵州茅台(更新)" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}

	// List
	instruments, err := store.ListInstruments(ctx, InstrumentFilter{})
	if err != nil {
		t.Fatalf("failed to list instruments: %v", err)
	}

	if len(instruments) != 1 {
		t.Errorf("expected 1 instrument, got %d", len(instruments))
	}

	// Count
	count, err := store.CountInstruments(ctx, InstrumentFilter{})
	if err != nil {
		t.Fatalf("failed to count instruments: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestBatchUpsertInstruments(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	instruments := []*model.Instrument{
		{Code: "sh600519", Name: "贵州茅台", Board: "主板", ListDate: "2001-08-27", IsActive: true},
		{Code: "sz000001", Name: "平安银行", Board: "主板", ListDate: "1991-04-03", IsActive: true},
		{Code: "sz300750", Name: "宁德时代", Board: "创业板", ListDate: "2018-06-11", IsActive: true},
	}

	if err := store.BatchUpsertInstruments(ctx, instruments); err != nil {
		t.Fatalf("failed to batch insert: %v", err)
	}

	count, err := store.CountInstruments(ctx, InstrumentFilter{})
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 instruments, got %d", count)
	}
}

func TestDailyBarsCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// 先插入instrument
	inst := &model.Instrument{
		Code: "sh600519", Name: "贵州茅台", IsActive: true, UpdatedAt: time.Now(),
	}
	store.UpsertInstrument(ctx, inst)

	// Insert bars
	bars := []*model.DailyBar{
		{Code: "sh600519", Date: "2024-10-01", Open: 1800, High: 1850, Low: 1790, Close: 1840, Volume: 1000000, Amount: 1840000000, AdjFactor: 1.0, Source: "tdx"},
		{Code: "sh600519", Date: "2024-10-02", Open: 1840, High: 1870, Low: 1830, Close: 1860, Volume: 1100000, Amount: 2046000000, AdjFactor: 1.0, Source: "tdx"},
		{Code: "sh600519", Date: "2024-10-03", Open: 1860, High: 1880, Low: 1850, Close: 1870, Volume: 1200000, Amount: 2244000000, AdjFactor: 1.0, Source: "tdx"},
	}

	if err := store.InsertDailyBars(ctx, bars); err != nil {
		t.Fatalf("failed to insert bars: %v", err)
	}

	// Get latest bars
	retrieved, err := store.GetDailyBars(ctx, "sh600519", 10)
	if err != nil {
		t.Fatalf("failed to get bars: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 bars, got %d", len(retrieved))
	}

	// 验证顺序（应该是降序）
	if retrieved[0].Date != "2024-10-03" {
		t.Errorf("expected first bar date 2024-10-03, got %s", retrieved[0].Date)
	}

	// Get by date range
	ranged, err := store.GetDailyBarsRange(ctx, "sh600519", "2024-10-01", "2024-10-02")
	if err != nil {
		t.Fatalf("failed to get bars by range: %v", err)
	}

	if len(ranged) != 2 {
		t.Errorf("expected 2 bars, got %d", len(ranged))
	}

	// Get latest bar
	latest, err := store.GetLatestBar(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to get latest bar: %v", err)
	}

	if latest.Date != "2024-10-03" {
		t.Errorf("expected latest date 2024-10-03, got %s", latest.Date)
	}

	// Count bars
	count, err := store.CountBars(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to count bars: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 bars, got %d", count)
	}

	// Get date range
	start, end, err := store.GetDateRange(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to get date range: %v", err)
	}

	if start != "2024-10-01" || end != "2024-10-03" {
		t.Errorf("expected range 2024-10-01 to 2024-10-03, got %s to %s", start, end)
	}
}

func TestScreenResultsCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	results := []*model.ScreenResult{
		{TradeDate: "2024-10-08", Code: "605133", Name: "世名科技", PctChange: 8.5, ClosePrice: 45.6, DD52: -15.2, CreatedAt: time.Now()},
		{TradeDate: "2024-10-08", Code: "600825", Name: "新华传媒", PctChange: 7.2, ClosePrice: 12.3, DD52: -20.1, CreatedAt: time.Now()},
	}

	// Save
	if err := store.SaveScreenResults(ctx, "2024-10-08", results); err != nil {
		t.Fatalf("failed to save screen results: %v", err)
	}

	// Load
	loaded, err := store.LoadScreenResults(ctx, "2024-10-08")
	if err != nil {
		t.Fatalf("failed to load screen results: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 results, got %d", len(loaded))
	}

	// Count
	count, err := store.CountScreenResults(ctx, "2024-10-08")
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results, got %d", count)
	}
}

func TestTransactionCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	txn := &model.Transaction{
		TradeDate:    "2024-10-08",
		Code:         "sh600519",
		Name:         "贵州茅台",
		Direction:    "buy",
		PlannedPrice: 1800,
		ActualPrice:  1805,
		Quantity:     100,
		Amount:       180500,
		Commission:   5.0,
		SlippagePct:  0.28,
		Note:         "梯①",
		Source:       "manual",
		CreatedAt:    time.Now(),
	}

	// Insert
	if err := store.InsertTransaction(ctx, txn); err != nil {
		t.Fatalf("failed to insert transaction: %v", err)
	}

	if txn.ID == 0 {
		t.Error("expected ID to be set")
	}

	// Get by ID
	retrieved, err := store.GetTransaction(ctx, txn.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}

	if retrieved.Code != txn.Code {
		t.Errorf("expected code %s, got %s", txn.Code, retrieved.Code)
	}

	// Get by date
	byDate, err := store.GetTransactionsByDate(ctx, "2024-10-08")
	if err != nil {
		t.Fatalf("failed to get transactions by date: %v", err)
	}

	if len(byDate) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(byDate))
	}

	// Get by code
	byCode, err := store.GetTransactionsByCode(ctx, "sh600519")
	if err != nil {
		t.Fatalf("failed to get transactions by code: %v", err)
	}

	if len(byCode) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(byCode))
	}
}

func TestTrackerPositionCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	pos := &model.TrackerPosition{
		Code:        "sh600519",
		Name:        "贵州茅台",
		AddedDate:   "2024-10-08",
		FrozenAt:    time.Now(),
		Tier1Price:  1800,
		Tier2Price:  1750,
		Tier3Price:  1700,
		StopPrice:   1650,
		ProfitPrice: 1900,
		Tier1Shares: 400,
		Tier2Shares: 300,
		Tier3Shares: 300,
		Tier1Status: "pending",
		Tier2Status: "pending",
		Tier3Status: "pending",
		Status:      "active",
		LastUpdate:  time.Now(),
		ExpireDate:  "2024-11-17",
	}

	// Insert
	if err := store.InsertTrackerPosition(ctx, pos); err != nil {
		t.Fatalf("failed to insert position: %v", err)
	}

	if pos.ID == 0 {
		t.Error("expected ID to be set")
	}

	// Get by ID
	retrieved, err := store.GetTrackerPosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}

	if retrieved.Code != pos.Code {
		t.Errorf("expected code %s, got %s", pos.Code, retrieved.Code)
	}

	// Update
	pos.Tier1Status = "filled"
	pos.Tier1FilledPrice = 1805
	pos.Tier1FilledDate = "2024-10-09"

	if err := store.UpdateTrackerPosition(ctx, pos); err != nil {
		t.Fatalf("failed to update position: %v", err)
	}

	updated, err := store.GetTrackerPosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("failed to get updated position: %v", err)
	}

	if updated.Tier1Status != "filled" {
		t.Errorf("expected tier1 status filled, got %s", updated.Tier1Status)
	}

	// List
	positions, err := store.ListTrackerPositions(ctx, "active")
	if err != nil {
		t.Fatalf("failed to list positions: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("expected 1 position, got %d", len(positions))
	}

	// Count
	count, err := store.CountTrackerPositions(ctx, "active")
	if err != nil {
		t.Fatalf("failed to count positions: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 position, got %d", count)
	}
}

func TestReviewCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	review := &model.Review{
		TradeDate:     "2024-10-08",
		FilledCount:   2,
		StoppedCount:  1,
		AvgSlippage:   0.25,
		DailyReturn:   0.015,
		Equity:        1050000,
		Cash:          850000,
		PositionValue: 200000,
		Notes:         `{"detail":"test"}`,
		CreatedAt:     time.Now(),
	}

	// Save
	if err := store.SaveReview(ctx, review); err != nil {
		t.Fatalf("failed to save review: %v", err)
	}

	// Get
	retrieved, err := store.GetReview(ctx, "2024-10-08")
	if err != nil {
		t.Fatalf("failed to get review: %v", err)
	}

	if retrieved.FilledCount != 2 {
		t.Errorf("expected filled count 2, got %d", retrieved.FilledCount)
	}

	// Update (via SaveReview)
	review.FilledCount = 3
	if err := store.SaveReview(ctx, review); err != nil {
		t.Fatalf("failed to update review: %v", err)
	}

	updated, err := store.GetReview(ctx, "2024-10-08")
	if err != nil {
		t.Fatalf("failed to get updated review: %v", err)
	}

	if updated.FilledCount != 3 {
		t.Errorf("expected updated filled count 3, got %d", updated.FilledCount)
	}
}
