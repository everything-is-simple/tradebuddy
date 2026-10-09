package store

import (
	"context"
	"fmt"
	"tradebuddy/internal/model"
)

// SaveLifecycleStates saves lifecycle analysis results
func (s *Store) SaveLifecycleStates(ctx context.Context, tradeDate string, states []*model.LifecycleState) error {
	if len(states) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 删除该日期的旧数据
	_, err = tx.ExecContext(ctx, "DELETE FROM lifecycle_states WHERE trade_date = ?", tradeDate)
	if err != nil {
		return fmt.Errorf("failed to delete old states: %w", err)
	}

	// 插入新数据
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO lifecycle_states (
			trade_date, code, stage, stage_desc, vol_class, vr13, pos52, gain_lo52, dd_hi52,
			survive_prob, ci_low, ci_high, sample_size, raw_view, evidence, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, state := range states {
		_, err := stmt.ExecContext(ctx,
			state.TradeDate,
			state.Code,
			state.Stage,
			state.StageDesc,
			state.VolClass,
			state.VR13,
			state.Pos52,
			state.GainLo52,
			state.DDHI52,
			state.SurviveProb,
			state.CILow,
			state.CIHigh,
			state.SampleSize,
			state.RawView,
			state.Evidence,
		)
		if err != nil {
			return fmt.Errorf("failed to insert state %s: %w", state.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LoadLifecycleStates retrieves lifecycle states for a trade date
func (s *Store) LoadLifecycleStates(ctx context.Context, tradeDate string) ([]*model.LifecycleState, error) {
	query := `
		SELECT trade_date, code, stage, stage_desc, vol_class, vr13, pos52, gain_lo52, dd_hi52,
			   survive_prob, ci_low, ci_high, sample_size, raw_view, evidence, created_at
		FROM lifecycle_states
		WHERE trade_date = ?
		ORDER BY code
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle states: %w", err)
	}
	defer rows.Close()

	var states []*model.LifecycleState
	for rows.Next() {
		state := &model.LifecycleState{}
		var createdAt string
		if err := rows.Scan(
			&state.TradeDate, &state.Code, &state.Stage, &state.StageDesc, &state.VolClass,
			&state.VR13, &state.Pos52, &state.GainLo52, &state.DDHI52,
			&state.SurviveProb, &state.CILow, &state.CIHigh, &state.SampleSize,
			&state.RawView, &state.Evidence, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		state.CreatedAt = parseTime(createdAt)
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating states: %w", err)
	}

	return states, nil
}

// GetLifecycleState retrieves a single lifecycle state
func (s *Store) GetLifecycleState(ctx context.Context, tradeDate, code string) (*model.LifecycleState, error) {
	query := `
		SELECT trade_date, code, stage, stage_desc, vol_class, vr13, pos52, gain_lo52, dd_hi52,
			   survive_prob, ci_low, ci_high, sample_size, raw_view, evidence, created_at
		FROM lifecycle_states
		WHERE trade_date = ? AND code = ?
	`

	state := &model.LifecycleState{}
	var createdAt string
	err := s.db.QueryRowContext(ctx, query, tradeDate, code).Scan(
		&state.TradeDate, &state.Code, &state.Stage, &state.StageDesc, &state.VolClass,
		&state.VR13, &state.Pos52, &state.GainLo52, &state.DDHI52,
		&state.SurviveProb, &state.CILow, &state.CIHigh, &state.SampleSize,
		&state.RawView, &state.Evidence, &createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle state: %w", err)
	}

	state.CreatedAt = parseTime(createdAt)
	return state, nil
}

// GetLifecycleStatesByStage retrieves states filtered by stage
func (s *Store) GetLifecycleStatesByStage(ctx context.Context, tradeDate, stage string) ([]*model.LifecycleState, error) {
	query := `
		SELECT trade_date, code, stage, stage_desc, vol_class, vr13, pos52, gain_lo52, dd_hi52,
			   survive_prob, ci_low, ci_high, sample_size, raw_view, evidence, created_at
		FROM lifecycle_states
		WHERE trade_date = ? AND stage = ?
		ORDER BY survive_prob DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate, stage)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle states: %w", err)
	}
	defer rows.Close()

	var states []*model.LifecycleState
	for rows.Next() {
		state := &model.LifecycleState{}
		var createdAt string
		if err := rows.Scan(
			&state.TradeDate, &state.Code, &state.Stage, &state.StageDesc, &state.VolClass,
			&state.VR13, &state.Pos52, &state.GainLo52, &state.DDHI52,
			&state.SurviveProb, &state.CILow, &state.CIHigh, &state.SampleSize,
			&state.RawView, &state.Evidence, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		state.CreatedAt = parseTime(createdAt)
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating states: %w", err)
	}

	return states, nil
}
