package tests

import (
	"strings"
	"testing"

	"Linux-url-shortener/internal/services"
	"Linux-url-shortener/internal/tests/mocks"
)

func TestGenerateCode(t *testing.T) {

	code := services.GenerateCode(6)

	if len(code) != 6 {
		t.Fatalf(
			"expected code length 6, got %d",
			len(code),
		)
	}
}

func TestGenerateCode_ZeroLength(t *testing.T) {
	code := services.GenerateCode(0)

	if code != "" {
		t.Fatalf(
			"expected empty code, got %q",
			code,
		)
	}
}

func TestGenerateCode_ValidCharacters(t *testing.T) {

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := services.GenerateCode(6)

	for _, char := range code {

		if !strings.ContainsRune(charset, char) {

			t.Fatalf(
				"generated invalid character %q",
				char,
			)
		}
	}
}

func TestGeneratedUniqueCode_AvailableCode(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsValue: false,
	}

	code, err := services.GenerateUniqueCode(repo)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected a generated code")
	}

	if len(code) != 6 {
		t.Fatalf(
			"expected code length 6, got %d",
			len(code),
		)
	}

	if !repo.ShortCodeExistsCalled {
		t.Fatal(
			"expected ShortCodeExist to be called",
		)
	}
}

func TestGeneratedUniqueCode_RepositoryError(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsErr: mocks.ErrDatabase,
	}

	code, err := services.GenerateUniqueCode(repo)

	if err == nil {
		t.Fatalf("expected repository error")
	}

	if code != "" {
		t.Fatalf("expected empty code, got %q", code)
	}

}

func TestGenerateUniqueCode_Collision(t *testing.T) {

	repo := &mocks.MockRepository{
		ShortCodeExistsSequence: []bool{true, true, false},
	}

	code, err := services.GenerateUniqueCode(repo)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatalf("expected a generated code")
	}

	if repo.ShortCodeExistsCalls != 3 {
		t.Fatalf(
			"expected 3 existence checks, got %d",
			repo.ShortCodeExistsCalls,
		)
	}
}

func TestGenerateUniqueCode_ReturnsRepositoryError(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsErr: mocks.ErrDatabase,
	}

	code, err := services.GenerateUniqueCode(repo)

	if err != mocks.ErrDatabase {
		t.Fatalf(
			"expected %v, got %v",
			mocks.ErrDatabase,
			err,
		)
	}

	if code != "" {
		t.Fatalf(
			"expected empty code, got %q",
			code,
		)
	}
}
