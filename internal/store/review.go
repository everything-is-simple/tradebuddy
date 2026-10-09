package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveReview saves a daily review report
func (s *Store) SaveReview(ctx context.Context, review *model.Review) error {
	query := `
		INSERT INTO reviews (
			trade_date, filled_count, stopped_count, avg_slippage,
			daily_return, equity, cash, position_value, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(trade_date) DO UPDATE SET
			filled_count = excluded.filled_count,
			stopped_count = excluded.stopped_count,
			avg_slippage = excluded.avg_slippage,
			daily_return = excluded.daily_return,
			equity = excluded.equity,
			cash = excluded.cash,
			position_value = excluded.position_value,
			notes = excluded.notes,
			created_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.ExecContext(ctx, query,
		review.TradeDate,
		review.FilledCount,
		review.StoppedCount,
		review.AvgSlippage,
		review.DailyReturn,
		review.Equity,
		review.Cash,
		review.PositionValue,
		review.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to save review: %w", err)
	}

	return nil
}

// GetReview retrieves a review by trade date
func (s *Store) GetReview(ctx context.Context, tradeDate string) (*model.Review, error) {
	query := `
		SELECT trade_date, filled_count, stopped_count, avg_slippage,
			   daily_return, equity, cash, position_value, notes, created_at
		FROM reviews
		WHERE trade_date = ?
	`

	review := &model.Review{}
	var createdAt string
	err := s.db.QueryRowContext(ctx, query, tradeDate).Scan(
		&review.TradeDate,
		&review.FilledCount,
		&review.StoppedCount,
		&review.AvgSlippage,
		&review.DailyReturn,
		&review.Equity,
		&review.Cash,
		&review.PositionValue,
		&review.Notes,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("review not found for %s", tradeDate)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	review.CreatedAt = parseTime(createdAt)
	return review, nil
}

// ListReviews retrieves reviews within a date range
func (s *Store) ListReviews(ctx context.Context, startDate, endDate string) ([]*model.Review, error) {
	query := `
		SELECT trade_date, filled_count, stopped_count, avg_slippage,
			   daily_return, equity, cash, position_value, notes, created_at
		FROM reviews
		WHERE trade_date >= ? AND trade_date <= ?
		ORDER BY trade_date DESC
	`

	rows, err := s.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*model.Review
	for rows.Next() {
		review := &model.Review{}
		var createdAt string
		if err := rows.Scan(
			&review.TradeDate,
			&review.FilledCount,
			&review.StoppedCount,
			&review.AvgSlippage,
			&review.DailyReturn,
			&review.Equity,
			&review.Cash,
			&review.PositionValue,
			&review.Notes,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		review.CreatedAt = parseTime(createdAt)
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviews: %w", err)
	}

	return reviews, nil
}

// GetLatestReview retrieves the most recent review
func (s *Store) GetLatestReview(ctx context.Context) (*model.Review, error) {
	query := `
		SELECT trade_date, filled_count, stopped_count, avg_slippage,
			   daily_return, equity, cash, position_value, notes, created_at
		FROM reviews
		ORDER BY trade_date DESC
		LIMIT 1
	`

	review := &model.Review{}
	var createdAt string
	err := s.db.QueryRowContext(ctx, query).Scan(
		&review.TradeDate,
		&review.FilledCount,
		&review.StoppedCount,
		&review.AvgSlippage,
		&review.DailyReturn,
		&review.Equity,
		&review.Cash,
		&review.PositionValue,
		&review.Notes,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no reviews found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest review: %w", err)
	}

	review.CreatedAt = parseTime(createdAt)
	return review, nil
}
