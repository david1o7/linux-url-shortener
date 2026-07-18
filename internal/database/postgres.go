package database

import (
	"Linux-url-shortener/internal/logger"
	"Linux-url-shortener/internal/models"
	"database/sql"

	_ "github.com/lib/pq"
)

func Connect(host, port, user, password, dbname, sslmode string) (*sql.DB, error) {

	connStr := string("host="+host+" port="+port +" user=" +user+" password="+ password+ " dbname="+ dbname+ " sslmode="+ sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func SaveUrl(db *sql.DB, shortCode string, OriginalCode string) error {
	query := `INSERT INTO urls(originalurl, shortcode) VALUES($1,$2)`

	_, err := db.Exec(query, OriginalCode,shortCode)


	if err != nil{
	logger.Log.Error(
		"Database query Error",
		"Error", err,
	)
	}
	return err
}

func GetUrl(db *sql.DB, shortcode string) (string, error){
	var original string
	query := `SELECT originalurl FROM urls WHERE shortcode = $1`

	err := db.QueryRow(query, shortcode).Scan(&original)

	if err != nil{
		return " ", err
	}
	return original, err
}

func GetByOriginal(db *sql.DB, original string) (*models.Url, error){
	query := `SELECT id, originalurl, shortcode, created_at from urls WHERE originalurl = $1`

	row := db.QueryRow(query, original)

	var url models.Url

	err := row.Scan(
		&url.ID,
		&url.Original,
		&url.ShortCode,
		&url.CreatedAt,
	)
	if err == sql.ErrNoRows{
		return nil, nil	
	}

	if err != nil{
		return nil, err
	}

	return &url, nil
}

func ShortCodeExist(db *sql.DB, shortCode string) (bool, error){
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE shortcode = $1)`

	var exists bool

	err := db.QueryRow(query, shortCode).Scan(&exists)

	return exists, err
}

func IncrementClicks(db *sql.DB, shortcode string) error {
	query := `UPDATE urls
	 SET 
	 clicks = clicks + 1, 
	 last_accessed = Now() 
	 WHERE shortcode = $1`

	 _, err := db.Exec(query, shortcode)

	if err != nil{
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
