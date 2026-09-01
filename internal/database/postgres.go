package database

import (
	"Linux-url-shortener/internal/config"
	"Linux-url-shortener/internal/logger"
	"Linux-url-shortener/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PostgresDSN())
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	logger.Log.Info(
		"Database pool configured",
		"max_open", cfg.DBMaxOpenConns,
		"max_idle", cfg.DBMaxIdleConns,
		"max_lifetime", cfg.DBConnMaxLifetime.String(),
		"max_idle_time", cfg.DBConnMaxIdleTime.String(),
	)

	return db, nil
}

func WaitForDB(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if lastErr = db.Ping(); lastErr == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("database not ready after %s: %w", timeout, lastErr)
}

func (p *PostgresRepository) SaveUrl(shortCode string, OriginalCode string) error {
	query := `INSERT INTO urls(originalurl, shortcode) VALUES($1,$2)`

	_, err := p.DB.Exec(query, OriginalCode, shortCode)

	if err != nil {
		var pgErr *pq.Error

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrShortCodeExist
		}

		logger.Log.Error(
			"Database query Error",
			"Error", err,
		)
	}

	return err
}

func (p *PostgresRepository) GetUrl(shortcode string) (string, error) {
	var original string
	query := `SELECT originalurl FROM urls WHERE shortcode = $1`

	err := p.DB.QueryRow(query, shortcode).Scan(&original)

	if err != nil {
		return " ", err
	}
	return original, err
}

func (p *PostgresRepository) GetByOriginal(original string) (*models.Url, error) {

	query := `SELECT id, originalurl, shortcode, created_at from urls WHERE originalurl = $1`

	row := p.DB.QueryRow(query, original)

	var url models.Url

	err := row.Scan(
		&url.ID,
		&url.Original,
		&url.ShortCode,
		&url.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (p *PostgresRepository) ShortCodeExist(shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE shortcode = $1)`

	var exists bool

	err := p.DB.QueryRow(query, shortCode).Scan(&exists)

	return exists, err
}

func (p *PostgresRepository) IncrementClicks(shortcode string) error {
	query := `UPDATE urls
	 SET 
	 clicks = clicks + 1, 
	 last_accessed = Now() 
	 WHERE shortcode = $1`

	_, err := p.DB.Exec(query, shortcode)

	if err != nil {
		logger.Log.Error(
			"Database query Error",
			"Error", err,
		)
	}

	logger.Log.Info(
		"URL clicked",
		"Short code", shortcode,
	)

	return err
}
