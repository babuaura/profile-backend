package contact

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorePersistsContactMessages(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "messages.jsonl"))
	message := Message{
		ID:        "lead-1",
		Name:      "Babu",
		Email:     "babu@example.com",
		Message:   "Hello from the portfolio",
		Source:    "test",
		Status:    "new",
		CreatedAt: time.Now().UTC(),
	}

	if err := store.Save(message); err != nil {
		t.Fatalf("save message: %v", err)
	}

	messages, err := store.List()
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 1 || messages[0].ID != message.ID {
		t.Fatalf("expected saved message, got %#v", messages)
	}

	updated, found, err := store.UpdateStatus(message.ID, "read")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if !found || updated.Status != "read" {
		t.Fatalf("expected read message, found=%v updated=%#v", found, updated)
	}

	found, err = store.Delete(message.ID)
	if err != nil {
		t.Fatalf("delete message: %v", err)
	}
	if !found {
		t.Fatal("expected message to be deleted")
	}
}
