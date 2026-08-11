package tests

import (
	"errors"
	"strings"
	"testing"

	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/services"
	"Linux-url-shortener/tests/mocks"
)

func TestGenerateCode(t *testing.T) {
	code := services.GenerateCode(6)

	if len(code) != 6 {
		t.Fatalf("expected code length 6, got %d", len(code))
	}
}

func TestGenerateCode_ZeroLength(t *testing.T) {
	code := services.GenerateCode(0)

	if code != "" {
		t.Fatalf("expected empty code, got %q", code)
	}
}

func TestGenerateCode_ValidCharacters(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := services.GenerateCode(100)

	for _, char := range code {
		if !strings.ContainsRune(charset, char) {
			t.Fatalf("generated invalid character %q", char)
		}
	}
}

func TestGenerateUniqueCode_Success(t *testing.T) {
	repo := &mocks.MockRepository{}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		10,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected generated shortcode")
	}

	if len(code) != 6 {
		t.Fatalf(
			"expected shortcode length 6, got %d",
			len(code),
		)
	}

	if !repo.SaveURLCalled {
		t.Fatal("expected SaveUrl to be called")
	}
}

func TestGenerateUniqueCode_SaveError(t *testing.T) {
	repo := &mocks.MockRepository{
		SaveURLErr: mocks.ErrDatabase,
	}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		10,
	)

	if !errors.Is(err, mocks.ErrDatabase) {
		t.Fatalf(
			"expected database error %v, got %v",
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

	if !repo.SaveURLCalled {
		t.Fatal("expected SaveUrl to be called")
	}
}

func TestGenerateUniqueCode_CollisionThenSuccess(t *testing.T) {
	attempts := 0

	repo := &mocks.MockRepository{}

	repo.SaveURLErr = nil

	originalSave := repo.SaveUrl

	_ = originalSave

	repo.SaveURLSequence = []error{
		database.ErrShortCodeExist,
		nil,
	}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		10,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code == "" {
		t.Fatal("expected generated shortcode")
	}

	if repo.SaveURLCalls != 2 {
		t.Fatalf(
			"expected 2 SaveUrl attempts, got %d",
			repo.SaveURLCalls,
		)
	}

	attempts = repo.SaveURLCalls

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestGenerateUniqueCode_MaxAttempts(t *testing.T) {
	repo := &mocks.MockRepository{
		SaveURLSequence: []error{
			database.ErrShortCodeExist,
			database.ErrShortCodeExist,
			database.ErrShortCodeExist,
		},
	}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		3,
	)

	if !errors.Is(
		err,
		services.ErrMaxCodeGenerationAttempts,
	) {
		t.Fatalf(
			"expected max attempts error, got %v",
			err,
		)
	}

	if code != "" {
		t.Fatalf(
			"expected empty code, got %q",
			code,
		)
	}

	if repo.SaveURLCalls != 3 {
		t.Fatalf(
			"expected 3 SaveUrl attempts, got %d",
			repo.SaveURLCalls,
		)
	}
}

func TestGenerateUniqueCode_DefaultMaxAttempts(t *testing.T) {
	repo := &mocks.MockRepository{}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		0,
	)

	if err != nil {
		t.Fatalf(
			"expected default max attempts behavior, got %v",
			err,
		)
	}

	if code == "" {
		t.Fatal("expected generated shortcode")
	}

	if repo.SaveURLCalls != 1 {
		t.Fatalf(
			"expected 1 SaveUrl attempt, got %d",
			repo.SaveURLCalls,
		)
	}
}

func TestGenerateUniqueCode_NegativeMaxAttempts(t *testing.T) {
	repo := &mocks.MockRepository{}

	code, err := services.GenerateUniqueCode(
		repo,
		"https://google.com",
		-5,
	)

	if err != nil {
		t.Fatalf(
			"expected default max attempts behavior, got %v",
			err,
		)
	}

	if code == "" {
		t.Fatal("expected generated shortcode")
	}

	if repo.SaveURLCalls != 1 {
		t.Fatalf(
			"expected 1 SaveUrl attempt, got %d",
			repo.SaveURLCalls,
		)
	}
}
