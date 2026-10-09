package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store provides data access layer for TradeBuddy
type Store struct {
	db *sql.DB
}

// Config holds database configuration
type Config struct {
	Path string // 数据库文件路径
}

// New creates a new Store instance
func New(cfg *Config) (*Store, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	// 确保目录存在
	dir := filepath.Dir(cfg.Path)
	if dir != "" && dir != "." {
		// 目录创建由外部处理，这里只验证
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 优化SQLite设置
	pragmas := []string{
		"PRAGMA journal_mode=WAL",        // Write-Ahead Logging
		"PRAGMA synchronous=NORMAL",      // 平衡性能和安全
		"PRAGMA foreign_keys=ON",         // 启用外键约束
		"PRAGMA busy_timeout=5000",       // 5秒超时
		"PRAGMA cache_size=-64000",       // 64MB缓存
		"PRAGMA temp_store=MEMORY",       // 临时表存内存
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	// 验证连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying database connection for advanced usage
func (s *Store) DB() *sql.DB {
	return s.db
}

// BeginTx starts a new transaction
func (s *Store) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

// Ping verifies database connection
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Stats returns database statistics
func (s *Store) Stats() sql.DBStats {
	return s.db.Stats()
}
