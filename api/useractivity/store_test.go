package useractivity

import "fmt"

func setup(path string) (*Store, error) {
	store, err := NewStore(path, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed creating new store: %w", err)
	}

	return store, nil
}
