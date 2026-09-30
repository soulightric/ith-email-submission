package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"monitoring-email-ith/config"
)

// Connect membuka connection pool ke PostgreSQL berdasarkan Config.
func Connect(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, err := conn.Exec(`ALTER TABLE email_requests ADD COLUMN IF NOT EXISTS formulir_path TEXT`); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("gagal menyiapkan kolom formulir: %w", err)
	}
	if _, err := conn.Exec(`ALTER TABLE email_requests ADD COLUMN IF NOT EXISTS contact_email VARCHAR(254) NOT NULL DEFAULT ''`); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("gagal menyiapkan kolom email kontak: %w", err)
	}
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(5)
	return conn, nil
}
