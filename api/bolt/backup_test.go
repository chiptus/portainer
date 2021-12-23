package bolt

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

func TestCreateBackupFolders(t *testing.T) {
	store := NewTestStore(portaineree.PortainerEE, portaineree.DBVersionEE, false)
	if exists, _ := store.fileService.FileExists("tmp/backups"); exists {
		t.Error("Expect backups folder to not exist")
	}
	store.createBackupFolders()
	if exists, _ := store.fileService.FileExists("tmp/backups"); !exists {
		t.Error("Expect backups folder to exist")
	}
	store.createBackupFolders()
	store.Close()
	teardown()
}

func TestStoreCreation(t *testing.T) {
	store := NewTestStore(portaineree.PortainerEE, portaineree.DBVersionEE, false)
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

	store.Close()
	teardown()
}

func TestBackup(t *testing.T) {

	tests := []struct {
		edition portaineree.SoftwareEdition
		version int
	}{
		{edition: portaineree.PortainerCE, version: portaineree.DBVersion},
		{edition: portaineree.PortainerEE, version: portaineree.DBVersionEE},
	}

	for _, tc := range tests {
		backupFileName := fmt.Sprintf("tmp/backups/%s/portainer.db.%03d.*", tc.edition.GetEditionLabel(), tc.version)
		t.Run(fmt.Sprintf("Backup should create %s", backupFileName), func(t *testing.T) {
			store := NewTestStore(tc.edition, tc.version, false)
			store.Backup()

			if !isFileExist(backupFileName) {
				t.Errorf("Expect backup file to be created %s", backupFileName)
			}
			store.Close()
		})
	}
	t.Run("BackupWithOption should create a name specific backup", func(t *testing.T) {
		edition := portaineree.PortainerCE
		version := portaineree.DBVersion
		store := NewTestStore(edition, version, false)
		store.BackupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerUpgradeToEEBackup,
			Edition:        portaineree.PortainerCE,
		})
		backupFileName := fmt.Sprintf("tmp/backups/%s/%s", edition.GetEditionLabel(), beforePortainerUpgradeToEEBackup)
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
		store.Close()
	})
	t.Run("BackupWithOption should create a name specific backup at common path", func(t *testing.T) {
		store := NewTestStore(portaineree.PortainerCE, portaineree.DBVersion, false)
		store.BackupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerVersionUpgradeBackup,
			BackupDir:      store.commonBackupDir(),
		})
		backupFileName := path.Join("tmp", "backups", "common", beforePortainerVersionUpgradeBackup)
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
		store.Close()
	})

	teardown()
}

// TODO restore / backup failed test cases
func TestRestore(t *testing.T) {

	editions := []portaineree.SoftwareEdition{portaineree.PortainerCE, portaineree.PortainerEE}
	var currentVersion = 0

	for i, e := range editions {
		editionLabel := e.GetEditionLabel()
		currentVersion = 10 ^ i + 1
		store := NewTestStore(e, currentVersion, false)
		t.Run(fmt.Sprintf("Basic Restore for %s", editionLabel), func(t *testing.T) {
			store.Backup()
			updateVersion(store, currentVersion+1)
			testVersion(store, currentVersion+1, t)
			store.Restore()
			testVersion(store, currentVersion, t)
		})
		t.Run(fmt.Sprintf("Basic Restore After Multiple Backup for %s", editionLabel), func(t *testing.T) {
			currentVersion = currentVersion + 5
			updateVersion(store, currentVersion)
			store.Backup()
			updateVersion(store, currentVersion+2)
			testVersion(store, currentVersion+2, t)
			store.Restore()
			testVersion(store, currentVersion, t)
		})
		store.Close()
	}

	teardown()
}

func TestRemoveWithOptions(t *testing.T) {
	store := NewTestStore(portaineree.PortainerCE, portaineree.DBVersion, false)

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

		err = store.RemoveWithOptions(options)
		if err != nil {
			t.Errorf("RemoveWithOptions should successfully remove file; err=%w", err)
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

		err := store.RemoveWithOptions(options)
		if err == nil {
			t.Error("RemoveWithOptions should fail for non-existent file")
		}
	})

	store.Close()
	teardown()
}
