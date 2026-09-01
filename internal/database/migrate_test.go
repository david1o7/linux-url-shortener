package database

import "testing"

func TestParseMigrationName_OK(t *testing.T) {
	v, name, err := parseMigrationName("000001_init_schema.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 || name != "init_schema" {
		t.Fatalf("got %d %q", v, name)
	}
}

func TestParseMigrationName_Invalid(t *testing.T) {
	cases := []string{
		"bad.up.sql",
		"init_schema.up.sql",
		"abc_name.up.sql",
	}
	for _, c := range cases {
		if _, _, err := parseMigrationName(c); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}
