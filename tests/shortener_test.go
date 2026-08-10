package tests

import (
	"errors"
	"strings"
	"testing"

	"Linux-url-shortener/internal/services"
	"Linux-url-shortener/tests/mocks"
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

	code, err := services.GenerateUniqueCode(repo, 10)

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

	code, err := services.GenerateUniqueCode(repo, 10)

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

	code, err := services.GenerateUniqueCode(repo, 10)

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

	code, err := services.GenerateUniqueCode(repo, 10)

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

func TestGenerateUniqueCode_ReturnsUniqueCode(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsValue: false,
	}

	code, err := services.GenerateUniqueCode(repo, 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected generated code, got empty string")
	}

	if len(code) != 6 {
		t.Fatalf(
			"expected code length 6, got %d",
			len(code),
		)
	}
}

func TestGenerateUniqueCode_RepositoryError(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsErr: mocks.ErrDatabase,
	}

	code, err := services.GenerateUniqueCode(repo, 10)

	if !errors.Is(err, mocks.ErrDatabase) {
		t.Fatalf(
			"expected repository error %v, got %v",
			mocks.ErrDatabase,
			err,
		)
	}

	if code != "" {
		t.Fatalf(
			"expected empty code on error, got %q",
			code,
		)
	}
}

func TestGenerateUniqueCode_InvalidMaxAttempts(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsValue: false,
	}

	code, err := services.GenerateUniqueCode(repo, 0)

	if err != nil {
		t.Fatalf(
			"expected default behavior, got %v",
			err,
		)
	}

	if code == "" {
		t.Fatal("expected generated code")
	}
}

func TestGenerateUniqueCode_MaxAttempts(t *testing.T) {
	repo := &mocks.MockRepository{
		ShortCodeExistsValue: true,
	}

	_, err := services.GenerateUniqueCode(repo, 3)

	if !errors.Is(err, services.ErrMaxCodeGenerationAttempts) {
		t.Fatalf(
			"expected max attempts error, got %v",
			err,
		)
	}
}

func TestGenerateUniqueCode_CollisionThenSuccess(t *testing.T) {
	attempts := 0

	repo := &mocks.MockRepository{
		ShortCodeExistFunc: func(code string) (bool, error) {
			attempts++

			if attempts < 3 {
				return true, nil
			}

			return false, nil
		},
	}

	code, err := services.GenerateUniqueCode(repo, 3)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if code == "" {
		t.Fatal("expected generated code")
	}

	if attempts != 3 {
		t.Fatalf(
			"expected 3 attempts, got %d",
			attempts,
		)
	}
}
