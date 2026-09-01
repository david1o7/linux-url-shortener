package integration

import (
	"errors"
	"testing"

	"Linux-url-shortener/internal/database"
)

func TestIntegration_SaveAndGetUrl(t *testing.T) {
	resetState(t)

	err := shared.Repo.SaveUrl("abc123", "https://example.com/page")
	if err != nil {
		t.Fatalf("SaveUrl: %v", err)
	}

	got, err := shared.Repo.GetUrl("abc123")
	if err != nil {
		t.Fatalf("GetUrl: %v", err)
	}
	if got != "https://example.com/page" {
		t.Fatalf("got %q", got)
	}
}

func TestIntegration_DuplicateShortCode(t *testing.T) {
	resetState(t)

	if err := shared.Repo.SaveUrl("dup001", "https://a.com"); err != nil {
		t.Fatal(err)
	}
	err := shared.Repo.SaveUrl("dup001", "https://b.com")
	if !errors.Is(err, database.ErrShortCodeExist) {
		t.Fatalf("expected ErrShortCodeExist, got %v", err)
	}
}

func TestIntegration_ShortCodeExist(t *testing.T) {
	resetState(t)

	ok, err := shared.Repo.ShortCodeExist("nope00")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected not exists")
	}

	if err := shared.Repo.SaveUrl("yes001", "https://yes.com"); err != nil {
		t.Fatal(err)
	}
	ok, err = shared.Repo.ShortCodeExist("yes001")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected exists")
	}
}

func TestIntegration_GetByOriginal(t *testing.T) {
	resetState(t)

	if err := shared.Repo.SaveUrl("orig01", "https://original.example"); err != nil {
		t.Fatal(err)
	}

	row, err := shared.Repo.GetByOriginal("https://original.example")
	if err != nil {
		t.Fatal(err)
	}
	if row == nil || row.ShortCode != "orig01" {
		t.Fatalf("unexpected row: %#v", row)
	}

	missing, err := shared.Repo.GetByOriginal("https://missing.example")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatalf("expected nil, got %#v", missing)
	}
}

func TestIntegration_IncrementClicks(t *testing.T) {
	resetState(t)

	if err := shared.Repo.SaveUrl("clk001", "https://clicks.example"); err != nil {
		t.Fatal(err)
	}
	if err := shared.Repo.IncrementClicks("clk001"); err != nil {
		t.Fatal(err)
	}
	if err := shared.Repo.IncrementClicks("clk001"); err != nil {
		t.Fatal(err)
	}

	var clicks int
	err := shared.DB.QueryRow(`SELECT clicks FROM urls WHERE shortcode = $1`, "clk001").Scan(&clicks)
	if err != nil {
		t.Fatal(err)
	}
	if clicks != 2 {
		t.Fatalf("clicks = %d, want 2", clicks)
	}
}

func TestIntegration_GetUrlMissing(t *testing.T) {
	resetState(t)

	_, err := shared.Repo.GetUrl("missing")
	if err == nil {
		t.Fatal("expected error for missing short code")
	}
}
