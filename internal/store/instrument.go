package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"tradebuddy/internal/model"
)

// UpsertInstrument inserts or updates an instrument
func (s *Store) UpsertInstrument(ctx context.Context, inst *model.Instrument) error {
	query := `
		INSERT INTO instruments (code, name, board, list_date, is_st, is_active, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(code) DO UPDATE SET
			name = excluded.name,
			board = excluded.board,
			list_date = excluded.list_date,
			is_st = excluded.is_st,
			is_active = excluded.is_active,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.ExecContext(ctx, query,
		inst.Code,
		inst.Name,
		inst.Board,
		inst.ListDate,
		boolToInt(inst.IsST),
		boolToInt(inst.IsActive),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert instrument: %w", err)
	}

	return nil
}

// GetInstrument retrieves a single instrument by code
func (s *Store) GetInstrument(ctx context.Context, code string) (*model.Instrument, error) {
	query := `
		SELECT code, name, board, list_date, is_st, is_active, updated_at
		FROM instruments
		WHERE code = ?
	`

	inst := &model.Instrument{}
	var isST, isActive int
	var updatedAt string

	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&inst.Code,
		&inst.Name,
		&inst.Board,
		&inst.ListDate,
		&isST,
		&isActive,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instrument not found: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument: %w", err)
	}

	inst.IsST = intToBool(isST)
	inst.IsActive = intToBool(isActive)
	inst.UpdatedAt = parseTime(updatedAt)

	return inst, nil
}

// InstrumentFilter defines filter criteria for listing instruments
type InstrumentFilter struct {
	ExcludeST   bool
	ActiveOnly  bool
	Codes       []string
	Limit       int
}

// ListInstruments retrieves instruments with optional filters
func (s *Store) ListInstruments(ctx context.Context, filter InstrumentFilter) ([]*model.Instrument, error) {
	query := "SELECT code, name, board, list_date, is_st, is_active, updated_at FROM instruments WHERE 1=1"
	args := []interface{}{}

	if filter.ExcludeST {
		query += " AND is_st = 0"
	}

	if filter.ActiveOnly {
		query += " AND is_active = 1"
	}

	if len(filter.Codes) > 0 {
		placeholders := make([]string, len(filter.Codes))
		for i, code := range filter.Codes {
			placeholders[i] = "?"
			args = append(args, code)
		}
		query += fmt.Sprintf(" AND code IN (%s)", joinStrings(placeholders, ","))
	}

	query += " ORDER BY code"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list instruments: %w", err)
	}
	defer rows.Close()

	var instruments []*model.Instrument
	for rows.Next() {
		inst := &model.Instrument{}
		var isST, isActive int
		var updatedAt string

		if err := rows.Scan(
			&inst.Code,
			&inst.Name,
			&inst.Board,
			&inst.ListDate,
			&isST,
			&isActive,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan instrument: %w", err)
		}

		inst.IsST = intToBool(isST)
		inst.IsActive = intToBool(isActive)
		inst.UpdatedAt = parseTime(updatedAt)
		instruments = append(instruments, inst)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating instruments: %w", err)
	}

	return instruments, nil
}

// BatchUpsertInstruments inserts or updates multiple instruments in a transaction
func (s *Store) BatchUpsertInstruments(ctx context.Context, instruments []*model.Instrument) error {
	if len(instruments) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO instruments (code, name, board, list_date, is_st, is_active, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(code) DO UPDATE SET
			name = excluded.name,
			board = excluded.board,
			list_date = excluded.list_date,
			is_st = excluded.is_st,
			is_active = excluded.is_active,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, inst := range instruments {
		_, err := stmt.ExecContext(ctx,
			inst.Code,
			inst.Name,
			inst.Board,
			inst.ListDate,
			boolToInt(inst.IsST),
			boolToInt(inst.IsActive),
		)
		if err != nil {
			return fmt.Errorf("failed to insert instrument %s: %w", inst.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CountInstruments returns total number of instruments
func (s *Store) CountInstruments(ctx context.Context, filter InstrumentFilter) (int, error) {
	query := "SELECT COUNT(*) FROM instruments WHERE 1=1"
	args := []interface{}{}

	if filter.ExcludeST {
		query += " AND is_st = 0"
	}

	if filter.ActiveOnly {
		query += " AND is_active = 1"
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count instruments: %w", err)
	}

	return count, nil
}

// Helper functions

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}
