package profile

import (
	"path/filepath"
	"testing"
)

func TestFileStoreSavesProfile(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "profile.json"))
	profile := defaultProfile()
	profile.Name = "Saved Profile"
	profile.Email = "saved@example.com"

	if _, err := store.Save(profile); err != nil {
		t.Fatalf("save profile: %v", err)
	}

	got, err := store.Get()
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if got.Name != profile.Name || got.Email != profile.Email {
		t.Fatalf("expected saved profile, got %#v", got)
	}
}
