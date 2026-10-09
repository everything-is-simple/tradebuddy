package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// InsertDailyBar inserts a single daily bar
func (s *Store) InsertDailyBar(ctx context.Context, bar *model.DailyBar) error {
	query := `
		INSERT INTO bars_daily (code, date, open, high, low, close, volume, amount, adj_factor, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(code, date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume,
			amount = excluded.amount,
			adj_factor = excluded.adj_factor,
			source = excluded.source
	`

	_, err := s.db.ExecContext(ctx, query,
		bar.Code, bar.Date, bar.Open, bar.High, bar.Low, bar.Close,
		bar.Volume, bar.Amount, bar.AdjFactor, bar.Source,
	)

	if err != nil {
		return fmt.Errorf("failed to insert daily bar: %w", err)
	}

	return nil
}

// InsertDailyBars inserts multiple daily bars in a transaction
func (s *Store) InsertDailyBars(ctx context.Context, bars []*model.DailyBar) error {
	if len(bars) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO bars_daily (code, date, open, high, low, close, volume, amount, adj_factor, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(code, date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume,
			amount = excluded.amount,
			adj_factor = excluded.adj_factor,
			source = excluded.source
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, bar := range bars {
		_, err := stmt.ExecContext(ctx,
			bar.Code, bar.Date, bar.Open, bar.High, bar.Low, bar.Close,
			bar.Volume, bar.Amount, bar.AdjFactor, bar.Source,
		)
		if err != nil {
			return fmt.Errorf("failed to insert bar %s-%s: %w", bar.Code, bar.Date, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDailyBars retrieves recent daily bars for a stock (newest first)
func (s *Store) GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error) {
	query := `
		SELECT code, date, open, high, low, close, volume, amount, adj_factor, source
		FROM bars_daily
		WHERE code = ?
		ORDER BY date DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, code, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily bars: %w", err)
	}
	defer rows.Close()

	var bars []*model.DailyBar
	for rows.Next() {
		bar := &model.DailyBar{}
		if err := rows.Scan(
			&bar.Code, &bar.Date, &bar.Open, &bar.High, &bar.Low, &bar.Close,
			&bar.Volume, &bar.Amount, &bar.AdjFactor, &bar.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bar: %w", err)
		}
		bars = append(bars, bar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bars: %w", err)
	}

	return bars, nil
}

// GetDailyBarsRange retrieves daily bars within a date range (inclusive)
func (s *Store) GetDailyBarsRange(ctx context.Context, code string, startDate, endDate string) ([]*model.DailyBar, error) {
	query := `
		SELECT code, date, open, high, low, close, volume, amount, adj_factor, source
		FROM bars_daily
		WHERE code = ? AND date >= ? AND date <= ?
		ORDER BY date ASC
	`

	rows, err := s.db.QueryContext(ctx, query, code, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily bars: %w", err)
	}
	defer rows.Close()

	var bars []*model.DailyBar
	for rows.Next() {
		bar := &model.DailyBar{}
		if err := rows.Scan(
			&bar.Code, &bar.Date, &bar.Open, &bar.High, &bar.Low, &bar.Close,
			&bar.Volume, &bar.Amount, &bar.AdjFactor, &bar.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bar: %w", err)
		}
		bars = append(bars, bar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bars: %w", err)
	}

	return bars, nil
}

// GetLatestBar retrieves the most recent bar for a stock
func (s *Store) GetLatestBar(ctx context.Context, code string) (*model.DailyBar, error) {
	query := `
		SELECT code, date, open, high, low, close, volume, amount, adj_factor, source
		FROM bars_daily
		WHERE code = ?
		ORDER BY date DESC
		LIMIT 1
	`

	bar := &model.DailyBar{}
	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&bar.Code, &bar.Date, &bar.Open, &bar.High, &bar.Low, &bar.Close,
		&bar.Volume, &bar.Amount, &bar.AdjFactor, &bar.Source,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no bars found for %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest bar: %w", err)
	}

	return bar, nil
}

// GetBarByDate retrieves a specific bar by code and date
func (s *Store) GetBarByDate(ctx context.Context, code string, date string) (*model.DailyBar, error) {
	query := `
		SELECT code, date, open, high, low, close, volume, amount, adj_factor, source
		FROM bars_daily
		WHERE code = ? AND date = ?
	`

	bar := &model.DailyBar{}
	err := s.db.QueryRowContext(ctx, query, code, date).Scan(
		&bar.Code, &bar.Date, &bar.Open, &bar.High, &bar.Low, &bar.Close,
		&bar.Volume, &bar.Amount, &bar.AdjFactor, &bar.Source,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bar not found for %s on %s", code, date)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bar: %w", err)
	}

	return bar, nil
}

// GetBarsByDate retrieves all bars for a specific trade date
func (s *Store) GetBarsByDate(ctx context.Context, date string) ([]*model.DailyBar, error) {
	query := `
		SELECT code, date, open, high, low, close, volume, amount, adj_factor, source
		FROM bars_daily
		WHERE date = ?
		ORDER BY code
	`

	rows, err := s.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query bars by date: %w", err)
	}
	defer rows.Close()

	var bars []*model.DailyBar
	for rows.Next() {
		bar := &model.DailyBar{}
		if err := rows.Scan(
			&bar.Code, &bar.Date, &bar.Open, &bar.High, &bar.Low, &bar.Close,
			&bar.Volume, &bar.Amount, &bar.AdjFactor, &bar.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bar: %w", err)
		}
		bars = append(bars, bar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bars: %w", err)
	}

	return bars, nil
}

// CountBars returns the number of bars for a stock
func (s *Store) CountBars(ctx context.Context, code string) (int, error) {
	query := "SELECT COUNT(*) FROM bars_daily WHERE code = ?"

	var count int
	err := s.db.QueryRowContext(ctx, query, code).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bars: %w", err)
	}

	return count, nil
}

// GetDateRange returns the earliest and latest dates for a stock
func (s *Store) GetDateRange(ctx context.Context, code string) (start, end string, err error) {
	query := `
		SELECT MIN(date), MAX(date)
		FROM bars_daily
		WHERE code = ?
	`

	err = s.db.QueryRowContext(ctx, query, code).Scan(&start, &end)
	if err != nil {
		return "", "", fmt.Errorf("failed to get date range: %w", err)
	}

	return start, end, nil
}

// DeleteBarsBefore deletes bars older than the specified date
func (s *Store) DeleteBarsBefore(ctx context.Context, code string, beforeDate string) error {
	query := "DELETE FROM bars_daily WHERE code = ? AND date < ?"

	result, err := s.db.ExecContext(ctx, query, code, beforeDate)
	if err != nil {
		return fmt.Errorf("failed to delete bars: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		// Log deletion count if needed
	}

	return nil
}
