package datastore

import (
	"fmt"
	"os"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database/models"

	"github.com/rs/zerolog/log"
)

func TestStoreCreation(t *testing.T) {
	_, store := MustNewTestStore(t, true, true)
	if store == nil {
		t.Fatal("Expect to create a store")
	}

	v, err := store.VersionService.Version()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	if portainer.SoftwareEdition(v.Edition) != portaineree.PortainerEE {
		t.Error("Expect to get EE Edition")
	}

	if v.SchemaVersion != portaineree.APIVersion {
		t.Error("Expect to get APIVersion")
	}
}

func TestBackup(t *testing.T) {
	_, store := MustNewTestStore(t, true, true)

	tests := []struct {
		edition portainer.SoftwareEdition
		version string
	}{
		{edition: portaineree.PortainerCE, version: portaineree.APIVersion},
		{edition: portaineree.PortainerEE, version: portaineree.APIVersion},
	}

	for _, tc := range tests {
		backupFileName := store.backupFilename()
		t.Run(fmt.Sprintf("Backup should create %s", backupFileName), func(t *testing.T) {
			v := models.Version{
				Edition:       int(tc.edition),
				SchemaVersion: tc.version,
			}
			store.VersionService.UpdateVersion(&v)
			store.Backup()

			if !isFileExist(backupFileName) {
				t.Errorf("Expect backup file to be created %s", backupFileName)
			}

			// delete the backup file for the next test
			err := os.Remove(backupFileName)
			if err != nil {
				t.Errorf("Failed to remove backup file %s", backupFileName)
			}
		})
	}
}

func TestRestore(t *testing.T) {
	editions := []portainer.SoftwareEdition{portaineree.PortainerCE, portaineree.PortainerEE}

	_, store := MustNewTestStore(t, true, false)

	for _, e := range editions {
		editionLabel := e.GetEditionLabel()

		t.Run(fmt.Sprintf("Basic Restore for %s", editionLabel), func(t *testing.T) {
			// override and set initial db version and edition
			updateEdition(store, e)
			updateVersion(store, "2.4")

			store.Backup()
			updateVersion(store, "2.16")
			testVersion(store, "2.16", t)
			store.Restore()

			// check if the restore is successful and the version is correct
			testVersion(store, "2.4", t)
		})
		t.Run(fmt.Sprintf("Basic Restore After Multiple Backup for %s", editionLabel), func(t *testing.T) {
			// override and set initial db version and edition
			updateEdition(store, e)
			updateVersion(store, "2.4")
			store.Backup()
			updateVersion(store, "2.14")
			updateVersion(store, "2.16")
			testVersion(store, "2.16", t)
			store.Restore()

			// check if the restore is successful and the version is correct
			testVersion(store, "2.4", t)
		})
	}
}
