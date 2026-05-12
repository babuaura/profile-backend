package storage

import (
	"strings"
	"testing"
)

func TestPostgresSchemaIncludesAllRuntimeTables(t *testing.T) {
	for _, table := range []string{
		"profiles",
		"contact_messages",
		"personal_notes",
		"personal_reminders",
		"personal_transactions",
		"personal_habits",
		"notification_tokens",
	} {
		if !strings.Contains(postgresSchema, "create table if not exists "+table) {
			t.Fatalf("postgres schema missing %s", table)
		}
	}
}
