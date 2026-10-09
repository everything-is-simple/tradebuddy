package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// InsertTrackerPosition inserts a new tracking position
func (s *Store) InsertTrackerPosition(ctx context.Context, pos *model.TrackerPosition) error {
	query := `
		INSERT INTO tracker_positions (
			code, name, added_date, frozen_at,
			tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			tier1_shares, tier2_shares, tier3_shares,
			tier1_status, tier2_status, tier3_status,
			status, last_update, expire_date, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		pos.Code, pos.Name, pos.AddedDate, pos.FrozenAt,
		pos.Tier1Price, pos.Tier2Price, pos.Tier3Price, pos.StopPrice, pos.ProfitPrice,
		pos.Tier1Shares, pos.Tier2Shares, pos.Tier3Shares,
		pos.Tier1Status, pos.Tier2Status, pos.Tier3Status,
		pos.Status, pos.LastUpdate, pos.ExpireDate, pos.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to insert tracker position: %w", err)
	}

	id, _ := result.LastInsertId()
	pos.ID = id

	return nil
}

// UpdateTrackerPosition updates an existing position
func (s *Store) UpdateTrackerPosition(ctx context.Context, pos *model.TrackerPosition) error {
	query := `
		UPDATE tracker_positions SET
			tier1_status = ?, tier2_status = ?, tier3_status = ?,
			tier1_filled_price = ?, tier2_filled_price = ?, tier3_filled_price = ?,
			tier1_filled_date = ?, tier2_filled_date = ?, tier3_filled_date = ?,
			status = ?, last_update = ?, notes = ?
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		pos.Tier1Status, pos.Tier2Status, pos.Tier3Status,
		pos.Tier1FilledPrice, pos.Tier2FilledPrice, pos.Tier3FilledPrice,
		pos.Tier1FilledDate, pos.Tier2FilledDate, pos.Tier3FilledDate,
		pos.Status, pos.LastUpdate, pos.Notes,
		pos.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update tracker position: %w", err)
	}

	return nil
}

// GetTrackerPosition retrieves a position by ID
func (s *Store) GetTrackerPosition(ctx context.Context, id int64) (*model.TrackerPosition, error) {
	query := `
		SELECT id, code, name, added_date, frozen_at,
			   tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			   tier1_shares, tier2_shares, tier3_shares,
			   tier1_status, tier2_status, tier3_status,
			   tier1_filled_price, tier2_filled_price, tier3_filled_price,
			   tier1_filled_date, tier2_filled_date, tier3_filled_date,
			   status, last_update, expire_date, notes
		FROM tracker_positions
		WHERE id = ?
	`

	pos := &model.TrackerPosition{}
	var frozenAt, lastUpdate string
	var tier1FilledPrice, tier2FilledPrice, tier3FilledPrice sql.NullFloat64
	var tier1FilledDate, tier2FilledDate, tier3FilledDate sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&pos.ID, &pos.Code, &pos.Name, &pos.AddedDate, &frozenAt,
		&pos.Tier1Price, &pos.Tier2Price, &pos.Tier3Price, &pos.StopPrice, &pos.ProfitPrice,
		&pos.Tier1Shares, &pos.Tier2Shares, &pos.Tier3Shares,
		&pos.Tier1Status, &pos.Tier2Status, &pos.Tier3Status,
		&tier1FilledPrice, &tier2FilledPrice, &tier3FilledPrice,
		&tier1FilledDate, &tier2FilledDate, &tier3FilledDate,
		&pos.Status, &lastUpdate, &pos.ExpireDate, &pos.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tracker position not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tracker position: %w", err)
	}

	pos.FrozenAt = parseTime(frozenAt)
	pos.LastUpdate = parseTime(lastUpdate)
	pos.Tier1FilledPrice = tier1FilledPrice.Float64
	pos.Tier2FilledPrice = tier2FilledPrice.Float64
	pos.Tier3FilledPrice = tier3FilledPrice.Float64
	pos.Tier1FilledDate = tier1FilledDate.String
	pos.Tier2FilledDate = tier2FilledDate.String
	pos.Tier3FilledDate = tier3FilledDate.String

	return pos, nil
}

// GetTrackerPositionByCode retrieves a position by code and added date
func (s *Store) GetTrackerPositionByCode(ctx context.Context, code, addedDate string) (*model.TrackerPosition, error) {
	query := `
		SELECT id, code, name, added_date, frozen_at,
			   tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			   tier1_shares, tier2_shares, tier3_shares,
			   tier1_status, tier2_status, tier3_status,
			   tier1_filled_price, tier2_filled_price, tier3_filled_price,
			   tier1_filled_date, tier2_filled_date, tier3_filled_date,
			   status, last_update, expire_date, notes
		FROM tracker_positions
		WHERE code = ? AND added_date = ?
	`

	pos := &model.TrackerPosition{}
	var frozenAt, lastUpdate string
	var tier1FilledPrice, tier2FilledPrice, tier3FilledPrice sql.NullFloat64
	var tier1FilledDate, tier2FilledDate, tier3FilledDate sql.NullString

	err := s.db.QueryRowContext(ctx, query, code, addedDate).Scan(
		&pos.ID, &pos.Code, &pos.Name, &pos.AddedDate, &frozenAt,
		&pos.Tier1Price, &pos.Tier2Price, &pos.Tier3Price, &pos.StopPrice, &pos.ProfitPrice,
		&pos.Tier1Shares, &pos.Tier2Shares, &pos.Tier3Shares,
		&pos.Tier1Status, &pos.Tier2Status, &pos.Tier3Status,
		&tier1FilledPrice, &tier2FilledPrice, &tier3FilledPrice,
		&tier1FilledDate, &tier2FilledDate, &tier3FilledDate,
		&pos.Status, &lastUpdate, &pos.ExpireDate, &pos.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tracker position not found: %s on %s", code, addedDate)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tracker position: %w", err)
	}

	pos.FrozenAt = parseTime(frozenAt)
	pos.LastUpdate = parseTime(lastUpdate)
	pos.Tier1FilledPrice = tier1FilledPrice.Float64
	pos.Tier2FilledPrice = tier2FilledPrice.Float64
	pos.Tier3FilledPrice = tier3FilledPrice.Float64
	pos.Tier1FilledDate = tier1FilledDate.String
	pos.Tier2FilledDate = tier2FilledDate.String
	pos.Tier3FilledDate = tier3FilledDate.String

	return pos, nil
}

// ListTrackerPositions retrieves positions with optional status filter
func (s *Store) ListTrackerPositions(ctx context.Context, status string) ([]*model.TrackerPosition, error) {
	query := `
		SELECT id, code, name, added_date, frozen_at,
			   tier1_price, tier2_price, tier3_price, stop_price, profit_price,
			   tier1_shares, tier2_shares, tier3_shares,
			   tier1_status, tier2_status, tier3_status,
			   tier1_filled_price, tier2_filled_price, tier3_filled_price,
			   tier1_filled_date, tier2_filled_date, tier3_filled_date,
			   status, last_update, expire_date, notes
		FROM tracker_positions
	`

	args := []interface{}{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY added_date DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tracker positions: %w", err)
	}
	defer rows.Close()

	var positions []*model.TrackerPosition
	for rows.Next() {
		pos := &model.TrackerPosition{}
		var frozenAt, lastUpdate string
		var tier1FilledPrice, tier2FilledPrice, tier3FilledPrice sql.NullFloat64
		var tier1FilledDate, tier2FilledDate, tier3FilledDate sql.NullString

		if err := rows.Scan(
			&pos.ID, &pos.Code, &pos.Name, &pos.AddedDate, &frozenAt,
			&pos.Tier1Price, &pos.Tier2Price, &pos.Tier3Price, &pos.StopPrice, &pos.ProfitPrice,
			&pos.Tier1Shares, &pos.Tier2Shares, &pos.Tier3Shares,
			&pos.Tier1Status, &pos.Tier2Status, &pos.Tier3Status,
			&tier1FilledPrice, &tier2FilledPrice, &tier3FilledPrice,
			&tier1FilledDate, &tier2FilledDate, &tier3FilledDate,
			&pos.Status, &lastUpdate, &pos.ExpireDate, &pos.Notes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		pos.FrozenAt = parseTime(frozenAt)
		pos.LastUpdate = parseTime(lastUpdate)
		pos.Tier1FilledPrice = tier1FilledPrice.Float64
		pos.Tier2FilledPrice = tier2FilledPrice.Float64
		pos.Tier3FilledPrice = tier3FilledPrice.Float64
		pos.Tier1FilledDate = tier1FilledDate.String
		pos.Tier2FilledDate = tier2FilledDate.String
		pos.Tier3FilledDate = tier3FilledDate.String

		positions = append(positions, pos)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	return positions, nil
}

// CountTrackerPositions returns the number of positions by status
func (s *Store) CountTrackerPositions(ctx context.Context, status string) (int, error) {
	query := "SELECT COUNT(*) FROM tracker_positions"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tracker positions: %w", err)
	}

	return count, nil
}

// DeleteTrackerPosition deletes a position by ID
func (s *Store) DeleteTrackerPosition(ctx context.Context, id int64) error {
	query := "DELETE FROM tracker_positions WHERE id = ?"

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tracker position: %w", err)
	}

	return nil
}
