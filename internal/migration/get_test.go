//go:build !migrationtest

package migration

import (
	"fmt"
	"testing"
)

func compareMigrations(t *testing.T, actual, expected []Migration) {
	if len(actual) != len(expected) {
		t.Fatalf("ERROR - mismatching lengths. Is %d, but should be %d", len(actual), len(expected))
	}

	for i := 0; i < len(actual); i++ {
		migrationActual := actual[i]
		migrationExpected := expected[i]

		if migrationActual.Version != migrationExpected.Version {
			fmt.Println(i, migrationActual, migrationExpected)
			t.Fatal("Migrations are not sorted the same")
		}
	}
}

func TestGetMigrationsSorted(t *testing.T) {
	const currentVersion = 0
	input := []Migration{
		{
			Version: 2,
		},
		{
			Version: 1,
		},
	}
	expected := []Migration{
		{
			Version: 1,
		},
		{
			Version: 2,
		},
	}

	actual := getMigrations(input, currentVersion)

	compareMigrations(t, actual, expected)
}

func TestGetMigrationsFilter(t *testing.T) {
	const currentVersion = 1
	input := []Migration{
		{
			Version: 2,
		},
		{
			Version: 1,
		},
		{
			Version: 3,
		},
	}
	expected := []Migration{
		{
			Version: 2,
		},
		{
			Version: 3,
		},
	}

	actual := getMigrations(input, currentVersion)

	compareMigrations(t, actual, expected)
}
