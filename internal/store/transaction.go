package store

import (
	"context"
	"database/sql"
	"fmt"
	"tradebuddy/internal/model"
)

// InsertTransaction inserts a new transaction record
func (s *Store) InsertTransaction(ctx context.Context, txn *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			trade_date, code, name, direction, planned_price, actual_price,
			quantity, amount, commission, slippage_pct, note, source, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query,
		txn.TradeDate, txn.Code, txn.Name, txn.Direction,
		txn.PlannedPrice, txn.ActualPrice, txn.Quantity,
		txn.Amount, txn.Commission, txn.SlippagePct,
		txn.Note, txn.Source,
	)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	id, _ := result.LastInsertId()
	txn.ID = id

	return nil
}

// GetTransaction retrieves a transaction by ID
func (s *Store) GetTransaction(ctx context.Context, id int64) (*model.Transaction, error) {
	query := `
		SELECT id, trade_date, code, name, direction, planned_price, actual_price,
			   quantity, amount, commission, slippage_pct, note, source, created_at
		FROM transactions
		WHERE id = ?
	`

	txn := &model.Transaction{}
	var createdAt string
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&txn.ID, &txn.TradeDate, &txn.Code, &txn.Name, &txn.Direction,
		&txn.PlannedPrice, &txn.ActualPrice, &txn.Quantity,
		&txn.Amount, &txn.Commission, &txn.SlippagePct,
		&txn.Note, &txn.Source, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	txn.CreatedAt = parseTime(createdAt)
	return txn, nil
}

// GetTransactionsByDate retrieves all transactions for a trade date
func (s *Store) GetTransactionsByDate(ctx context.Context, tradeDate string) ([]*model.Transaction, error) {
	query := `
		SELECT id, trade_date, code, name, direction, planned_price, actual_price,
			   quantity, amount, commission, slippage_pct, note, source, created_at
		FROM transactions
		WHERE trade_date = ?
		ORDER BY created_at
	`

	rows, err := s.db.QueryContext(ctx, query, tradeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		txn := &model.Transaction{}
		var createdAt string
		if err := rows.Scan(
			&txn.ID, &txn.TradeDate, &txn.Code, &txn.Name, &txn.Direction,
			&txn.PlannedPrice, &txn.ActualPrice, &txn.Quantity,
			&txn.Amount, &txn.Commission, &txn.SlippagePct,
			&txn.Note, &txn.Source, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		txn.CreatedAt = parseTime(createdAt)
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetTransactionsByCode retrieves all transactions for a stock
func (s *Store) GetTransactionsByCode(ctx context.Context, code string) ([]*model.Transaction, error) {
	query := `
		SELECT id, trade_date, code, name, direction, planned_price, actual_price,
			   quantity, amount, commission, slippage_pct, note, source, created_at
		FROM transactions
		WHERE code = ?
		ORDER BY trade_date DESC, created_at
	`

	rows, err := s.db.QueryContext(ctx, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		txn := &model.Transaction{}
		var createdAt string
		if err := rows.Scan(
			&txn.ID, &txn.TradeDate, &txn.Code, &txn.Name, &txn.Direction,
			&txn.PlannedPrice, &txn.ActualPrice, &txn.Quantity,
			&txn.Amount, &txn.Commission, &txn.SlippagePct,
			&txn.Note, &txn.Source, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		txn.CreatedAt = parseTime(createdAt)
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetTransactionsByDateRange retrieves transactions within a date range
func (s *Store) GetTransactionsByDateRange(ctx context.Context, startDate, endDate string) ([]*model.Transaction, error) {
	query := `
		SELECT id, trade_date, code, name, direction, planned_price, actual_price,
			   quantity, amount, commission, slippage_pct, note, source, created_at
		FROM transactions
		WHERE trade_date >= ? AND trade_date <= ?
		ORDER BY trade_date DESC, created_at
	`

	rows, err := s.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		txn := &model.Transaction{}
		var createdAt string
		if err := rows.Scan(
			&txn.ID, &txn.TradeDate, &txn.Code, &txn.Name, &txn.Direction,
			&txn.PlannedPrice, &txn.ActualPrice, &txn.Quantity,
			&txn.Amount, &txn.Commission, &txn.SlippagePct,
			&txn.Note, &txn.Source, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		txn.CreatedAt = parseTime(createdAt)
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// DeleteTransaction deletes a transaction by ID
func (s *Store) DeleteTransaction(ctx context.Context, id int64) error {
	query := "DELETE FROM transactions WHERE id = ?"

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}
