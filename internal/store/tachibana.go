package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveTachibanaSignals 保存立花信号
func (s *Store) SaveTachibanaSignals(ctx context.Context, tradeDate string, signals []*model.TachibanaSignal) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO tachibana_signals (
			trade_date, code, name,
			signal_type, confidence,
			current_price, entry_zone_low, entry_zone_high, stop_loss,
			lifecycle_grade, lifecycle_score, span_days, dist_from_20dh, atr_normalized,
			title, description, risk, suggestion
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(trade_date, code) DO UPDATE SET
			name = excluded.name,
			signal_type = excluded.signal_type,
			confidence = excluded.confidence,
			current_price = excluded.current_price,
			entry_zone_low = excluded.entry_zone_low,
			entry_zone_high = excluded.entry_zone_high,
			stop_loss = excluded.stop_loss,
			lifecycle_grade = excluded.lifecycle_grade,
			lifecycle_score = excluded.lifecycle_score,
			span_days = excluded.span_days,
			dist_from_20dh = excluded.dist_from_20dh,
			atr_normalized = excluded.atr_normalized,
			title = excluded.title,
			description = excluded.description,
			risk = excluded.risk,
			suggestion = excluded.suggestion
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, sig := range signals {
		_, err := stmt.ExecContext(ctx,
			tradeDate, sig.Code, sig.Name,
			sig.SignalType, sig.Confidence,
			sig.CurrentPrice, sig.EntryZoneLow, sig.EntryZoneHigh, sig.StopLoss,
			sig.LifecycleGrade, sig.LifecycleScore, sig.SpanDays, sig.DistFrom20DH, sig.ATRNormalized,
			sig.Title, sig.Description, sig.Risk, sig.Suggestion,
		)
		if err != nil {
			return fmt.Errorf("failed to insert signal for %s: %w", sig.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTachibanaSignals 获取立花信号
func (s *Store) GetTachibanaSignals(ctx context.Context, tradeDate string) ([]*model.TachibanaSignal, error) {
	query := `
		SELECT
			code, name, trade_date,
			signal_type, confidence,
			current_price, entry_zone_low, entry_zone_high, stop_loss,
			lifecycle_grade, lifecycle_score, span_days, dist_from_20dh, atr_normalized,
			title, description, risk, suggestion
		FROM tachibana_signals
		WHERE trade_date = ?
		ORDER BY
			CASE signal_type
				WHEN 'trend_probe_entry' THEN 1
				WHEN 'trend_confirmation_add' THEN 2
				WHEN 'distribution_reduce' THEN 3
				WHEN 'exit_on_rhythm_failure' THEN 4
				ELSE 5
			END,
			lifecycle_score DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var signals []*model.TachibanaSignal

	for rows.Next() {
		sig := &model.TachibanaSignal{}
		err := rows.Scan(
			&sig.Code, &sig.Name, &sig.TradeDate,
			&sig.SignalType, &sig.Confidence,
			&sig.CurrentPrice, &sig.EntryZoneLow, &sig.EntryZoneHigh, &sig.StopLoss,
			&sig.LifecycleGrade, &sig.LifecycleScore, &sig.SpanDays, &sig.DistFrom20DH, &sig.ATRNormalized,
			&sig.Title, &sig.Description, &sig.Risk, &sig.Suggestion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		signals = append(signals, sig)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return signals, nil
}

// GetTachibanaSignalsByType 按类型获取信号
func (s *Store) GetTachibanaSignalsByType(ctx context.Context, tradeDate string, signalType string) ([]*model.TachibanaSignal, error) {
	query := `
		SELECT
			code, name, trade_date,
			signal_type, confidence,
			current_price, entry_zone_low, entry_zone_high, stop_loss,
			lifecycle_grade, lifecycle_score, span_days, dist_from_20dh, atr_normalized,
			title, description, risk, suggestion
		FROM tachibana_signals
		WHERE trade_date = ? AND signal_type = ?
		ORDER BY lifecycle_score DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate, signalType)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var signals []*model.TachibanaSignal

	for rows.Next() {
		sig := &model.TachibanaSignal{}
		err := rows.Scan(
			&sig.Code, &sig.Name, &sig.TradeDate,
			&sig.SignalType, &sig.Confidence,
			&sig.CurrentPrice, &sig.EntryZoneLow, &sig.EntryZoneHigh, &sig.StopLoss,
			&sig.LifecycleGrade, &sig.LifecycleScore, &sig.SpanDays, &sig.DistFrom20DH, &sig.ATRNormalized,
			&sig.Title, &sig.Description, &sig.Risk, &sig.Suggestion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		signals = append(signals, sig)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return signals, nil
}

// DeleteTachibanaSignals 删除指定日期的信号
func (s *Store) DeleteTachibanaSignals(ctx context.Context, tradeDate string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM tachibana_signals WHERE trade_date = ?", tradeDate)
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
