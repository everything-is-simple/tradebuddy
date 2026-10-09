package store

import (
	"context"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveTachibanaSignals saves trading signals
func (s *Store) SaveTachibanaSignals(ctx context.Context, tradeDate string, signals []*model.TachibanaSignal) error {
	if len(signals) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 删除该日期的旧数据
	_, err = tx.ExecContext(ctx, "DELETE FROM tachibana_signals WHERE trade_date = ?", tradeDate)
	if err != nil {
		return fmt.Errorf("failed to delete old signals: %w", err)
	}

	// 插入新数据
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO tachibana_signals (
			trade_date, code, name, decision, z_score, band_now,
			tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			shares_t1, shares_t2, shares_t3, notional, risk_amount, reason, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, signal := range signals {
		_, err := stmt.ExecContext(ctx,
			signal.TradeDate,
			signal.Code,
			signal.Name,
			signal.Decision,
			signal.ZScore,
			signal.BandNow,
			signal.Tier1Price,
			signal.Tier2Price,
			signal.Tier3Price,
			signal.StopPrice,
			signal.ProfitPrice,
			signal.SharesT1,
			signal.SharesT2,
			signal.SharesT3,
			signal.Notional,
			signal.RiskAmount,
			signal.Reason,
		)
		if err != nil {
			return fmt.Errorf("failed to insert signal %s: %w", signal.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LoadTachibanaSignals retrieves signals for a trade date
func (s *Store) LoadTachibanaSignals(ctx context.Context, tradeDate string) ([]*model.TachibanaSignal, error) {
	query := `
		SELECT trade_date, code, name, decision, z_score, band_now,
			   tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			   shares_t1, shares_t2, shares_t3, notional, risk_amount, reason, created_at
		FROM tachibana_signals
		WHERE trade_date = ?
		ORDER BY decision, code
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query tachibana signals: %w", err)
	}
	defer rows.Close()

	var signals []*model.TachibanaSignal
	for rows.Next() {
		sig := &model.TachibanaSignal{}
		var createdAt string
		if err := rows.Scan(
			&sig.TradeDate, &sig.Code, &sig.Name, &sig.Decision, &sig.ZScore, &sig.BandNow,
			&sig.Tier1Price, &sig.Tier2Price, &sig.Tier3Price, &sig.StopPrice, &sig.ProfitPrice,
			&sig.SharesT1, &sig.SharesT2, &sig.SharesT3, &sig.Notional, &sig.RiskAmount,
			&sig.Reason, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan signal: %w", err)
		}
		sig.CreatedAt = parseTime(createdAt)
		signals = append(signals, sig)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating signals: %w", err)
	}

	return signals, nil
}

// GetTachibanaSignalsByDecision retrieves signals filtered by decision
func (s *Store) GetTachibanaSignalsByDecision(ctx context.Context, tradeDate, decision string) ([]*model.TachibanaSignal, error) {
	query := `
		SELECT trade_date, code, name, decision, z_score, band_now,
			   tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			   shares_t1, shares_t2, shares_t3, notional, risk_amount, reason, created_at
		FROM tachibana_signals
		WHERE trade_date = ? AND decision = ?
		ORDER BY z_score
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate, decision)
	if err != nil {
		return nil, fmt.Errorf("failed to query tachibana signals: %w", err)
	}
	defer rows.Close()

	var signals []*model.TachibanaSignal
	for rows.Next() {
		sig := &model.TachibanaSignal{}
		var createdAt string
		if err := rows.Scan(
			&sig.TradeDate, &sig.Code, &sig.Name, &sig.Decision, &sig.ZScore, &sig.BandNow,
			&sig.Tier1Price, &sig.Tier2Price, &sig.Tier3Price, &sig.StopPrice, &sig.ProfitPrice,
			&sig.SharesT1, &sig.SharesT2, &sig.SharesT3, &sig.Notional, &sig.RiskAmount,
			&sig.Reason, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan signal: %w", err)
		}
		sig.CreatedAt = parseTime(createdAt)
		signals = append(signals, sig)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating signals: %w", err)
	}

	return signals, nil
}
