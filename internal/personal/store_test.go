package personal

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorePersistsPersonalWorkspace(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "personal.json"))
	now := time.Now().UTC()

	note, err := store.CreateNote(Note{
		ID:        "note-1",
		Title:     "Backend check",
		Body:      "Stored in file mode",
		Tags:      []string{"test"},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}
	if note.ID != "note-1" {
		t.Fatalf("unexpected note: %#v", note)
	}

	reminder, err := store.CreateReminder(Reminder{
		ID:        "reminder-1",
		Title:     "Ship backend",
		DueAt:     now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create reminder: %v", err)
	}

	transaction, err := store.CreateTransaction(Transaction{
		ID:         "tx-1",
		Type:       "income",
		Amount:     100,
		Category:   "Test",
		OccurredAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	habit, err := store.CreateHabit(Habit{
		ID:        "habit-1",
		Name:      "Review",
		Target:    "Daily backend check",
		Frequency: "daily",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	state, err := store.Summary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if len(state.Notes) != 1 || len(state.Reminders) != 1 || len(state.Transactions) != 1 || len(state.Habits) != 1 {
		t.Fatalf("expected all workspace records, got %#v", state)
	}

	toggled, found, err := store.ToggleReminder(reminder.ID)
	if err != nil {
		t.Fatalf("toggle reminder: %v", err)
	}
	if !found || !toggled.Done {
		t.Fatalf("expected toggled reminder, found=%v reminder=%#v", found, toggled)
	}

	checked, found, err := store.CheckInHabit(habit.ID)
	if err != nil {
		t.Fatalf("check in habit: %v", err)
	}
	if !found || checked.Streak != 1 {
		t.Fatalf("expected checked habit, found=%v habit=%#v", found, checked)
	}

	for label, remove := range map[string]func(string) (bool, error){
		"note":        store.DeleteNote,
		"reminder":    store.DeleteReminder,
		"transaction": store.DeleteTransaction,
		"habit":       store.DeleteHabit,
	} {
		id := map[string]string{
			"note":        note.ID,
			"reminder":    reminder.ID,
			"transaction": transaction.ID,
			"habit":       habit.ID,
		}[label]
		found, err := remove(id)
		if err != nil {
			t.Fatalf("delete %s: %v", label, err)
		}
		if !found {
			t.Fatalf("expected %s to be deleted", label)
		}
	}
}
