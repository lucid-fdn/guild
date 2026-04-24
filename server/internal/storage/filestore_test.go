package storage

import "testing"

func TestPutRejectsUnsafeID(t *testing.T) {
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Put("taskpacks", "../escape", map[string]string{"ok": "no"}); err == nil {
		t.Fatal("expected unsafe id to be rejected")
	}
}

func TestPutRejectsUnknownCollection(t *testing.T) {
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Put("unknown", "4e1fe00c-6303-453c-8cb6-2c34f84896e4", map[string]string{"ok": "no"}); err == nil {
		t.Fatal("expected unknown collection to be rejected")
	}
}
