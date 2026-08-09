package tests

import (
	"Linux-url-shortener/internal/models"
	"Linux-url-shortener/internal/tests/mocks"
	"testing"
)

func TestMockRepository_SaveURL(t *testing.T) {

	repo := &mocks.MockRepository{}

	err := repo.SaveUrl("abc123", "https://google.com")

	if err != nil {
		t.Fatal(err)
	}

	if !repo.SaveCalled {
		t.Fatal("SaveURL was never called")
	}
}

func TestMockRepository_SaveURL_Error(t *testing.T) {

	repo := &mocks.MockRepository{
		SaveErr: mocks.ErrDatabase,
	}

	err := repo.SaveUrl("abc", "https://google.com")

	if err == nil {
		t.Fatal("expected database error")
	}
}

func TestMockRepository_GetURL(t *testing.T) {

	repo := &mocks.MockRepository{
		URL: "https://google.com",
	}

	url, err := repo.GetUrl("abc")

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

	repo := &mocks.MockRepository{}

	err := repo.IncrementClicks("abc")

	if err != nil {
		t.Fatal(err)
	}

	if !repo.IncrementCalled {
		t.Fatal("IncrementClicks wasn't called")
	}
}

func TestMockRepository_ShortCodeExists(t *testing.T) {

	repo := &mocks.MockRepository{
		ShortCodeExistsValue: true,
	}

	ok, err := repo.ShortCodeExist("abc")

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

	repo := &mocks.MockRepository{
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

func TestMockRepository_GetURL_Error(t *testing.T) {
	repo := &mocks.MockRepository{
		GetErr: mocks.ErrDatabase,
	}

	_, err := repo.GetUrl("abc")

	if err == nil {
		t.Fatal("expected database error")
	}

	if !repo.GetCalled {
		t.Fatal("expected GetUrl to be called")
	}
}

func TestMockRepository_IncrementClicks_Error(t *testing.T) {
	repo := &mocks.MockRepository{
		IncrementErr: mocks.ErrDatabase,
	}

	err := repo.IncrementClicks("abc")

	if err == nil {
		t.Fatal("expected database error")
	}

	if !repo.IncrementCalled {
		t.Fatal("expected IncrementClicks to be called")
	}
}

func TestMockRepository_GetByOriginal_Error(t *testing.T) {
	repo := &mocks.MockRepository{
		GetByOriginalErr: mocks.ErrDatabase,
	}

	_, err := repo.GetByOriginal("https://google.com")

	if err == nil {
		t.Fatal("expected database error")
	}

	if !repo.GetByOriginalCalled {
		t.Fatal("expected GetByOriginal to be called")
	}
}

func TestMockRepository_ShortCodeExists_Error(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsErr: mocks.ErrDatabase,
	}

	_, err := repo.ShortCodeExist("abc")

	if err == nil {
		t.Fatal("expected database error")
	}

	if !repo.ShortCodeExistsCalled {
		t.Fatal("expected ShortCodeExist to be called")
	}
}
