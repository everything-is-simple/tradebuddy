package store

import (
	"context"
	"fmt"
	"strings"
)

// Migrate executes database migrations
func (s *Store) Migrate(ctx context.Context) error {
	// 读取迁移文件
	migrationFiles := []string{
		"migrations/0001_init.sql",
	}

	for _, file := range migrationFiles {
		content, err := migrationsFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		// 执行迁移SQL
		if err := s.executeMigration(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

// executeMigration executes a migration SQL script
func (s *Store) executeMigration(ctx context.Context, sql string) error {
	// 分割SQL语句（以分号分隔）
	statements := splitSQL(sql)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nSQL: %s", err, stmt)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// splitSQL splits SQL script into individual statements
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	inComment := false

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 跳过注释行
		if strings.HasPrefix(trimmed, "--") {
			inComment = false
			continue
		}

		// 多行注释处理
		if strings.Contains(trimmed, "/*") {
			inComment = true
		}
		if strings.Contains(trimmed, "*/") {
			inComment = false
			continue
		}
		if inComment {
			continue
		}

		// 累积语句
		if trimmed != "" {
			current.WriteString(line)
			current.WriteString("\n")

			// 遇到分号，保存语句
			if strings.HasSuffix(trimmed, ";") {
				statements = append(statements, current.String())
				current.Reset()
			}
		}
	}

	// 最后一个语句（如果没有分号结尾）
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

// GetSchemaVersion returns current schema version
func (s *Store) GetSchemaVersion(ctx context.Context) (int, error) {
	var version int
	err := s.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// ResetDatabase drops all tables (for testing)
func (s *Store) ResetDatabase(ctx context.Context) error {
	tables := []string{
		"job_runs",
		"reviews",
		"transactions",
		"tracker_positions",
		"tachibana_signals",
		"lifecycle_states",
		"screen_results",
		"bars_daily",
		"instruments",
		"schema_migrations",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reset: %w", err)
	}

	return nil
}
