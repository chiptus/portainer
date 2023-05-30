package useractivity

import (
	"testing"
)

func mustSetup(t testing.TB) *Store {
	store, err := NewStore(t.TempDir(), 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed creating new store: %s", err)
	}

	t.Cleanup(func() { store.Close() })

	return store
}
