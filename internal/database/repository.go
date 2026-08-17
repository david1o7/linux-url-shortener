package database

import (
	"Linux-url-shortener/internal/models"
	"errors"
)

type Repository interface {
	SaveUrl(shortCode string, originalURL string) error

	GetUrl(shortCode string) (string, error)

	GetByOriginal(originalURL string) (*models.Url, error)

	IncrementClicks(shortCode string) error

	ShortCodeExist(shortCode string) (bool, error)
}

var ErrShortCodeExist = errors.New("short code already exists")
