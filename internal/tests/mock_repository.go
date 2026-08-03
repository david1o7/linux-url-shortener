package tests

import (
	"errors"

	"Linux-url-shortener/internal/models"
)

type MockRepository struct {
	SaveCalled            bool
	GetCalled             bool
	IncrementCalled       bool
	ShortCodeExistsCalled bool
	GetByOriginalCalled   bool

	SaveErr            error
	GetErr             error
	IncrementErr       error
	ShortCodeExistsErr error
	GetByOriginalErr   error

	URL                  string
	ExistingURL          *models.Url
	ShortCodeExistsValue bool
}

func (m *MockRepository) SaveURL(shortCode, originalURL string) error {
	m.SaveCalled = true
	return m.SaveErr
}

func (m *MockRepository) GetURL(shortCode string) (string, error) {
	m.GetCalled = true

	if m.GetErr != nil {
		return "", m.GetErr
	}

	return m.URL, nil
}

func (m *MockRepository) IncrementClicks(shortCode string) error {
	m.IncrementCalled = true
	return m.IncrementErr
}

func (m *MockRepository) GetByOriginal(original string) (*models.Url, error) {
	m.GetByOriginalCalled = true

	if m.GetByOriginalErr != nil {
		return nil, m.GetByOriginalErr
	}

	return m.ExistingURL, nil
}

func (m *MockRepository) ShortCodeExists(code string) (bool, error) {
	m.ShortCodeExistsCalled = true

	if m.ShortCodeExistsErr != nil {
		return false, m.ShortCodeExistsErr
	}

	return m.ShortCodeExistsValue, nil
}

var ErrDatabase = errors.New("database error")
