package datastore

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

func TestCreateBackupFolders(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, false, true)
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
		t.Error("Expect to create a store")
	}

	if store.edition() != portaineree.PortainerEE {
		t.Error("Expect to get EE Edition")
	}

	version, err := store.version()
	if err != nil {
		log.Fatal(err)
	}

	if version != portaineree.DBVersionEE {
		t.Error("Expect to get EE DBVersion")
	}
}

func TestBackup(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, true)
	connection := store.GetConnection()
	defer teardown()

	tests := []struct {
		edition portaineree.SoftwareEdition
		version int
	}{
		{edition: portaineree.PortainerCE, version: portaineree.DBVersion},
		{edition: portaineree.PortainerEE, version: portaineree.DBVersionEE},
	}

	for _, tc := range tests {
		backupFileName := fmt.Sprintf("%s/backups/%s/%s.%03d.*", store.connection.GetStorePath(), tc.edition.GetEditionLabel(), store.connection.GetDatabaseFileName(), tc.version)
		t.Run(fmt.Sprintf("Backup should create %s", backupFileName), func(t *testing.T) {
			store.VersionService.StoreDBVersion(tc.version)
			store.VersionService.StoreEdition(tc.edition)
			store.Backup()

			if !isFileExist(backupFileName) {
				t.Errorf("Expect backup file to be created %s", backupFileName)
			}
		})
	}
	t.Run("BackupWithOption should create a name specific backup", func(t *testing.T) {
		store.VersionService.StoreEdition(portaineree.PortainerCE)
		store.VersionService.StoreDBVersion(portaineree.DBVersion)
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
		store.VersionService.StoreEdition(portaineree.PortainerCE)
		store.VersionService.StoreDBVersion(portaineree.DBVersion)

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
	currentVersion := 0

	_, store, teardown := MustNewTestStore(t, true, true)
	defer teardown()

	for i, e := range editions {
		editionLabel := e.GetEditionLabel()
		currentVersion = 10 ^ i + 1

		t.Run(fmt.Sprintf("Basic Restore for %s", editionLabel), func(t *testing.T) {
			// override and set initial db version and edition
			updateEdition(store, e)
			updateVersion(store, currentVersion)

			store.Backup()
			updateVersion(store, currentVersion+1)
			testVersion(store, currentVersion+1, t)
			store.Restore()

			// check if the restore is successful and the version is correct
			testVersion(store, currentVersion, t)
		})
		t.Run(fmt.Sprintf("Basic Restore After Multiple Backup for %s", editionLabel), func(t *testing.T) {
			// override and set initial db version and edition
			updateEdition(store, e)
			updateVersion(store, currentVersion)

			currentVersion = currentVersion + 5
			updateVersion(store, currentVersion)
			store.Backup()
			updateVersion(store, currentVersion+2)
			testVersion(store, currentVersion+2, t)
			store.Restore()

			// check if the restore is successful and the version is correct
			testVersion(store, currentVersion, t)
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
