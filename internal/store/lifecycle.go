package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveLifecycleMetrics 保存生命周期指标
func (s *Store) SaveLifecycleMetrics(ctx context.Context, tradeDate string, metrics []*model.LifecycleMetrics) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO lifecycle_analysis (
			trade_date, code, name,
			span_days, days_since_20dh,
			price_range, price_range_pct, atr_normalized,
			dist_from_20dh, dist_from_52wh,
			span_rank, range_rank, atr_rank,
			lifecycle_score, grade
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(trade_date, code) DO UPDATE SET
			name = excluded.name,
			span_days = excluded.span_days,
			days_since_20dh = excluded.days_since_20dh,
			price_range = excluded.price_range,
			price_range_pct = excluded.price_range_pct,
			atr_normalized = excluded.atr_normalized,
			dist_from_20dh = excluded.dist_from_20dh,
			dist_from_52wh = excluded.dist_from_52wh,
			span_rank = excluded.span_rank,
			range_rank = excluded.range_rank,
			atr_rank = excluded.atr_rank,
			lifecycle_score = excluded.lifecycle_score,
			grade = excluded.grade
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range metrics {
		_, err := stmt.ExecContext(ctx,
			tradeDate, m.Code, m.Name,
			m.SpanDays, m.DaysSince20DH,
			m.PriceRange, m.PriceRangePct, m.ATRNormalized,
			m.DistFrom20DH, m.DistFrom52WH,
			m.SpanRank, m.RangeRank, m.ATRRank,
			m.LifecycleScore, m.Grade,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric for %s: %w", m.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLifecycleMetrics 获取生命周期指标
func (s *Store) GetLifecycleMetrics(ctx context.Context, tradeDate string) ([]*model.LifecycleMetrics, error) {
	query := `
		SELECT
			code, name, trade_date,
			span_days, days_since_20dh,
			price_range, price_range_pct, atr_normalized,
			dist_from_20dh, dist_from_52wh,
			span_rank, range_rank, atr_rank,
			lifecycle_score, grade
		FROM lifecycle_analysis
		WHERE trade_date = ?
		ORDER BY lifecycle_score DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var metrics []*model.LifecycleMetrics

	for rows.Next() {
		m := &model.LifecycleMetrics{}
		err := rows.Scan(
			&m.Code, &m.Name, &m.TradeDate,
			&m.SpanDays, &m.DaysSince20DH,
			&m.PriceRange, &m.PriceRangePct, &m.ATRNormalized,
			&m.DistFrom20DH, &m.DistFrom52WH,
			&m.SpanRank, &m.RangeRank, &m.ATRRank,
			&m.LifecycleScore, &m.Grade,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return metrics, nil
}

// GetLifecycleMetricsByGrade 按评级获取
func (s *Store) GetLifecycleMetricsByGrade(ctx context.Context, tradeDate string, grade string) ([]*model.LifecycleMetrics, error) {
	query := `
		SELECT
			code, name, trade_date,
			span_days, days_since_20dh,
			price_range, price_range_pct, atr_normalized,
			dist_from_20dh, dist_from_52wh,
			span_rank, range_rank, atr_rank,
			lifecycle_score, grade
		FROM lifecycle_analysis
		WHERE trade_date = ? AND grade = ?
		ORDER BY lifecycle_score DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate, grade)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var metrics []*model.LifecycleMetrics

	for rows.Next() {
		m := &model.LifecycleMetrics{}
		err := rows.Scan(
			&m.Code, &m.Name, &m.TradeDate,
			&m.SpanDays, &m.DaysSince20DH,
			&m.PriceRange, &m.PriceRangePct, &m.ATRNormalized,
			&m.DistFrom20DH, &m.DistFrom52WH,
			&m.SpanRank, &m.RangeRank, &m.ATRRank,
			&m.LifecycleScore, &m.Grade,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return metrics, nil
}

// DeleteLifecycleMetrics 删除指定日期的生命周期指标
func (s *Store) DeleteLifecycleMetrics(ctx context.Context, tradeDate string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM lifecycle_analysis WHERE trade_date = ?", tradeDate)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
