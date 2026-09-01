package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"Linux-url-shortener/internal/logger"
)

func Migrate(db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}
	sort.Strings(files)

	for _, path := range files {
		version, name, err := parseMigrationName(filepath.Base(path))
		if err != nil {
			return err
		}
		if applied[version] {
			continue
		}

		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		logger.Log.Info(
			"Applying migration",
			"version", version,
			"name", name,
		)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute %s: %w", name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			version, name,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}

		logger.Log.Info("Migration applied", "version", version, "name", name)
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    BIGINT PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func appliedVersions(db *sql.DB) (map[int64]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			return
		}
	}()

	out := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func parseMigrationName(filename string) (version int64, name string, err error) {
	base := strings.TrimSuffix(filename, ".up.sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename %q (want NNNNNN_name.up.sql)", filename)
	}
	version, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration version in %q: %w", filename, err)
	}
	return version, parts[1], nil
}
