package tests

import (
	"Linux-url-shortener/internal/models"
	"testing"
)

func TestMockRepository_SaveURL(t *testing.T) {

	repo := &MockRepository{}

	err := repo.SaveURL("abc123", "https://google.com")

	if err != nil {
		t.Fatal(err)
	}

	if !repo.SaveCalled {
		t.Fatal("SaveURL was never called")
	}
}

func TestMockRepository_SaveURL_Error(t *testing.T) {

	repo := &MockRepository{
		SaveErr: ErrDatabase,
	}

	err := repo.SaveURL("abc", "https://google.com")

	if err == nil {
		t.Fatal("expected database error")
	}
}

func TestMockRepository_GetURL(t *testing.T) {

	repo := &MockRepository{
		URL: "https://google.com",
	}

	url, err := repo.GetURL("abc")

	if err != nil {
		t.Fatal(err)
	}

	if url != "https://google.com" {
		t.Fatal("wrong url returned")
	}

	if !repo.GetCalled {
		t.Fatal("GetURL wasn't called")
	}
}

func TestMockRepository_IncrementClicks(t *testing.T) {

	repo := &MockRepository{}

	err := repo.IncrementClicks("abc")

	if err != nil {
		t.Fatal(err)
	}

	if !repo.IncrementCalled {
		t.Fatal("IncrementClicks wasn't called")
	}
}

func TestMockRepository_ShortCodeExists(t *testing.T) {

	repo := &MockRepository{
		ShortCodeExistsValue: true,
	}

	ok, err := repo.ShortCodeExists("abc")

	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("expected shortcode to exist")
	}

	if !repo.ShortCodeExistsCalled {
		t.Fatal("method wasn't called")
	}
}

func TestMockRepository_GetByOriginal(t *testing.T) {

	expected := &models.Url{
		Original: "https://google.com",
	}

	repo := &MockRepository{
		ExistingURL: expected,
	}

	url, err := repo.GetByOriginal("https://google.com")

	if err != nil {
		t.Fatal(err)
	}

	if url == nil {
		t.Fatal("expected url")
	}

	if !repo.GetByOriginalCalled {
		t.Fatal("method not called")
	}
}