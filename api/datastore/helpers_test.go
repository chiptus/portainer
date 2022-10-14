package datastore

import (
	"path/filepath"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/rs/zerolog/log"
)

// isFileExist is helper function to check for file existence
func isFileExist(path string) bool {
	matches, err := filepath.Glob(path)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func updateVersion(store *Store, v int) {
	err := store.VersionService.StoreDBVersion(v)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func updateEdition(store *Store, edition portaineree.SoftwareEdition) {
	err := store.VersionService.StoreEdition(edition)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func testVersion(store *Store, versionWant int, t *testing.T) {
	if v, _ := store.version(); v != versionWant {
		t.Errorf("Expect store version to be %d but was %d", versionWant, v)
	}
}
