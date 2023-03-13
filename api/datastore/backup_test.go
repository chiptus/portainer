package datastore

import (
	"fmt"
	"os"
	"path"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"

	"github.com/rs/zerolog/log"
)

func TestCreateBackupFolders(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, true)
	defer teardown()

	connection := store.GetConnection()
	backupPath := path.Join(connection.GetStorePath(), backupDefaults.backupDir)

	if isFileExist(backupPath) {
		t.Error("Expect backups folder to not exist")
	}

	store.createBackupFolders()
	if !isFileExist(backupPath) {
		t.Error("Expect backups folder to exist")
	}
}

func TestStoreCreation(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, true)
	defer teardown()

	if store == nil {
		t.Fatal("Expect to create a store")
	}

	v, err := store.VersionService.Version()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	if portaineree.SoftwareEdition(v.Edition) != portaineree.PortainerEE {
		t.Error("Expect to get EE Edition")
	}

	if v.SchemaVersion != portaineree.APIVersion {
		t.Error("Expect to get APIVersion")
	}
}

func TestBackup(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, true)
	connection := store.GetConnection()
	defer teardown()

	tests := []struct {
		edition portaineree.SoftwareEdition
		version string
	}{
		{edition: portaineree.PortainerCE, version: portaineree.APIVersion},
		{edition: portaineree.PortainerEE, version: portaineree.APIVersion},
	}

	for _, tc := range tests {
		backupFileName := fmt.Sprintf("%s/backups/%s/%s.%s.*", store.connection.GetStorePath(), tc.edition.GetEditionLabel(), store.connection.GetDatabaseFileName(), tc.version)
		t.Run(fmt.Sprintf("Backup should create %s", backupFileName), func(t *testing.T) {
			v := models.Version{
				Edition:       int(tc.edition),
				SchemaVersion: tc.version,
			}
			store.VersionService.UpdateVersion(&v)
			store.Backup(nil)

			if !isFileExist(backupFileName) {
				t.Errorf("Expect backup file to be created %s", backupFileName)
			}
		})
	}
	t.Run("BackupWithOption should create a name specific backup", func(t *testing.T) {
		v := models.Version{
			Edition:       int(portaineree.PortainerCE),
			SchemaVersion: portaineree.APIVersion,
		}
		store.VersionService.UpdateVersion(&v)
		store.backupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerUpgradeToEEBackup,
			Edition:        portaineree.PortainerCE,
		})
		backupFileName := fmt.Sprintf("%s/backups/%s/%s", store.connection.GetStorePath(), portaineree.PortainerCE.GetEditionLabel(), beforePortainerUpgradeToEEBackup)
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
	})
	t.Run("BackupWithOption should create a name specific backup at common path", func(t *testing.T) {
		v := models.Version{
			Edition:       int(portaineree.PortainerCE),
			SchemaVersion: portaineree.APIVersion,
		}
		store.VersionService.UpdateVersion(&v)

		store.backupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerVersionUpgradeBackup,
			BackupDir:      store.commonBackupDir(),
		})
		backupFileName := path.Join(connection.GetStorePath(), "backups", "common", beforePortainerVersionUpgradeBackup)
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
	})
}

func TestRestore(t *testing.T) {
	editions := []portaineree.SoftwareEdition{portaineree.PortainerCE, portaineree.PortainerEE}

	_, store, teardown := MustNewTestStore(t, true, true)
	defer teardown()

	for _, e := range editions {
		editionLabel := e.GetEditionLabel()

		t.Run(fmt.Sprintf("Basic Restore for %s", editionLabel), func(t *testing.T) {
			// override and set initial db version and edition
			updateEdition(store, e)
			updateVersion(store, "2.4")

			store.Backup(nil)
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

			updateVersion(store, "2.14")
			store.Backup(nil)
			updateVersion(store, "2.16")
			testVersion(store, "2.16", t)
			store.Restore()

			// check if the restore is successful and the version is correct
			testVersion(store, "2.4", t)
		})
	}
}

func TestRemoveWithOptions(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, true)
	defer teardown()

	t.Run("successfully removes file if existent", func(t *testing.T) {
		store.createBackupFolders()
		options := &BackupOptions{
			BackupDir:      store.commonBackupDir(),
			BackupFileName: "test.txt",
		}

		filePath := path.Join(options.BackupDir, options.BackupFileName)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("file should be created; err=%s", err)
		}
		f.Close()

		err = store.removeWithOptions(options)
		if err != nil {
			t.Errorf("RemoveWithOptions should successfully remove file; err=%v", err)
		}

		if isFileExist(f.Name()) {
			t.Errorf("RemoveWithOptions should successfully remove file; file=%s", f.Name())
		}
	})

	t.Run("fails to removes file if non-existent", func(t *testing.T) {
		options := &BackupOptions{
			BackupDir:      store.commonBackupDir(),
			BackupFileName: "test.txt",
		}

		err := store.removeWithOptions(options)
		if err == nil {
			t.Error("RemoveWithOptions should fail for non-existent file")
		}
	})
}
