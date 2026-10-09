package store

import (
	"context"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveScreenResults saves screening results for a trade date
func (s *Store) SaveScreenResults(ctx context.Context, tradeDate string, results []*model.ScreenResult) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 先删除该日期的旧数据
	_, err = tx.ExecContext(ctx, "DELETE FROM screen_results WHERE trade_date = ?", tradeDate)
	if err != nil {
		return fmt.Errorf("failed to delete old results: %w", err)
	}

	// 插入新数据
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO screen_results (
			trade_date, code, name, pct_change, close_price, high_price,
			volume, amount, turnover, dd52, weekly_ma, monthly_ma, data_source, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, result := range results {
		_, err := stmt.ExecContext(ctx,
			result.TradeDate,
			result.Code,
			result.Name,
			result.PctChange,
			result.ClosePrice,
			result.HighPrice,
			result.Volume,
			result.Amount,
			result.Turnover,
			result.DD52,
			result.WeeklyMA,
			result.MonthlyMA,
			result.DataSource,
		)
		if err != nil {
			return fmt.Errorf("failed to insert result %s: %w", result.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LoadScreenResults retrieves screening results for a trade date
func (s *Store) LoadScreenResults(ctx context.Context, tradeDate string) ([]*model.ScreenResult, error) {
	query := `
		SELECT trade_date, code, name, pct_change, close_price, high_price,
			   volume, amount, turnover, dd52, weekly_ma, monthly_ma, data_source, created_at
		FROM screen_results
		WHERE trade_date = ?
		ORDER BY pct_change DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query screen results: %w", err)
	}
	defer rows.Close()

	var results []*model.ScreenResult
	for rows.Next() {
		r := &model.ScreenResult{}
		var createdAt string
		if err := rows.Scan(
			&r.TradeDate, &r.Code, &r.Name, &r.PctChange, &r.ClosePrice, &r.HighPrice,
			&r.Volume, &r.Amount, &r.Turnover, &r.DD52, &r.WeeklyMA, &r.MonthlyMA,
			&r.DataSource, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		r.CreatedAt = parseTime(createdAt)
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	return results, nil
}

// GetScreenResultsByDateRange retrieves results within a date range
func (s *Store) GetScreenResultsByDateRange(ctx context.Context, startDate, endDate string) ([]*model.ScreenResult, error) {
	query := `
		SELECT trade_date, code, name, pct_change, close_price, high_price,
			   volume, amount, turnover, dd52, weekly_ma, monthly_ma, data_source, created_at
		FROM screen_results
		WHERE trade_date >= ? AND trade_date <= ?
		ORDER BY trade_date DESC, pct_change DESC
	`

	rows, err := s.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query screen results: %w", err)
	}
	defer rows.Close()

	var results []*model.ScreenResult
	for rows.Next() {
		r := &model.ScreenResult{}
		var createdAt string
		if err := rows.Scan(
			&r.TradeDate, &r.Code, &r.Name, &r.PctChange, &r.ClosePrice, &r.HighPrice,
			&r.Volume, &r.Amount, &r.Turnover, &r.DD52, &r.WeeklyMA, &r.MonthlyMA,
			&r.DataSource, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		r.CreatedAt = parseTime(createdAt)
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	return results, nil
}

// CountScreenResults returns the number of results for a trade date
func (s *Store) CountScreenResults(ctx context.Context, tradeDate string) (int, error) {
	query := "SELECT COUNT(*) FROM screen_results WHERE trade_date = ?"

	var count int
	err := s.db.QueryRowContext(ctx, query, tradeDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count screen results: %w", err)
	}

	return count, nil
}

// GetLatestScreenDate returns the most recent screening date
func (s *Store) GetLatestScreenDate(ctx context.Context) (string, error) {
	query := "SELECT MAX(trade_date) FROM screen_results"

	var date string
	err := s.db.QueryRowContext(ctx, query).Scan(&date)
	if err != nil {
		return "", fmt.Errorf("failed to get latest screen date: %w", err)
	}

	return date, nil
}
