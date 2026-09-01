package integration

import (
	"testing"

	"Linux-url-shortener/internal/database"
)

func TestIntegration_MigrateIdempotent(t *testing.T) {
	resetState(t)

	// Second run must be a no-op (already applied).
	if err := database.Migrate(shared.DB, migrationsDir()); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var n int
	err := shared.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected at least 1 migration row, got %d", n)
	}

	// urls table exists
	var exists bool
	err = shared.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'urls'
		)`).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("urls table missing after migrate")
	}
}
