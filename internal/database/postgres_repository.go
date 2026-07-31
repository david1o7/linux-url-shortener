package database

import "database/sql"

type PostgresRepository struct{
	DB *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {

	return &PostgresRepository{
		DB: db,
	}
}
