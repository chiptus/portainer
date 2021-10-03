package bolt

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/boltdb/bolt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
)

// New Database should be EE and DBVersion
//

func TestMigrateData(t *testing.T) {
	var store *Store

	t.Run("MigrateData for New Store & Re-Open Check", func(t *testing.T) {
		fileService, err := filesystem.NewService(dataStorePath, "")
		if err != nil {
			log.Fatal(err)
		}

		store := NewStore(dataStorePath, fileService)
		err = store.Open()
		if err != nil {
			log.Fatal(err)
		}

		err = store.Init()
		if err != nil {
			log.Fatal(err)
		}

		store.MigrateData(false)

		testVersion(store, portainer.DBVersionEE, t)
		testEdition(store, portainer.PortainerEE, t)

		store.Close()

		store.Open()
		if err != nil {
			log.Fatal(err)
		}

		if store.IsNew() {
			t.Error("Expect store to NOT be new DB")
		}

		store.Close()
	})

	tests := []struct {
		edition         portainer.SoftwareEdition
		version         int
		expectedVersion int
	}{
		{edition: portainer.PortainerCE, version: 5, expectedVersion: portainer.DBVersionEE},
		{edition: portainer.PortainerCE, version: 21, expectedVersion: portainer.DBVersionEE},
	}

	for _, tc := range tests {
		store = NewTestStore(tc.edition, tc.version, true)
		t.Run(fmt.Sprintf("MigrateData for %s version %d", tc.edition.GetEditionLabel(), tc.version), func(t *testing.T) {
			store.MigrateData(false)
			testVersion(store, tc.expectedVersion, t)
			testEdition(store, portainer.PortainerEE, t)
		})

		t.Run(fmt.Sprintf("Restoring DB after migrateData for %s version %d", tc.edition.GetEditionLabel(), tc.version), func(t *testing.T) {
			store.rollbackToCE(true)
			testVersion(store, tc.version, t)
			testEdition(store, tc.edition, t)
		})

		store.Close()
	}

	t.Run("Error in MigrateData should restore backup before MigrateData", func(t *testing.T) {
		version := 21
		store = NewTestStore(portainer.PortainerCE, version, true)

		deleteBucket(store.connection.DB, "settings")
		store.MigrateData(false)

		testVersion(store, version, t)
		testEdition(store, portainer.PortainerCE, t)

		store.Close()
	})

	t.Run("MigrateData should create backup file upon update", func(t *testing.T) {
		version := 21
		store = NewTestStore(portainer.PortainerCE, version, true)

		store.MigrateData(true)

		options := getBackupRestoreOptions(store)
		options = store.setupOptions(options)

		if !isFileExist(options.BackupPath) {
			t.Errorf("Backup file should exist; file=%s", options.BackupPath)
		}

		os.Remove(options.BackupPath)
		store.Close()
	})

	t.Run("MigrateData should fail to create backup if database file is set to updating", func(t *testing.T) {
		version := 21
		store = NewTestStore(portainer.PortainerCE, version, true)
		store.VersionService.StoreIsUpdating(true)

		store.MigrateData(true)

		options := getBackupRestoreOptions(store)
		options = store.setupOptions(options)

		if isFileExist(options.BackupPath) {
			t.Errorf("Backup file should not exist for dirty database; file=%s", options.BackupPath)
		}

		store.Close()
	})

	t.Run("MigrateData should not create backup on startup if portainer version matches db", func(t *testing.T) {
		store = NewTestStore(portainer.PortainerCE, portainer.DBVersion, true)

		store.MigrateData(true)

		options := getBackupRestoreOptions(store)
		options = store.setupOptions(options)

		if isFileExist(options.BackupPath) {
			t.Errorf("Backup file should not exist for dirty database; file=%s", options.BackupPath)
		}

		store.Close()
	})

	teardown()
}

func Test_getBackupRestoreOptions(t *testing.T) {
	store := NewTestStore(portainer.PortainerCE, portainer.DBVersion, true)
	defer store.Close()

	options := getBackupRestoreOptions(store)

	wantDir := store.commonBackupDir()
	if !strings.HasSuffix(options.BackupDir, wantDir) {
		log.Fatalf("incorrect backup dir; got=%s, want=%s", options.BackupDir, wantDir)
	}

	wantFilename := "portainer.db.bak"
	if options.BackupFileName != wantFilename {
		log.Fatalf("incorrect backup file; got=%s, want=%s", options.BackupFileName, wantFilename)
	}

	teardown()
}

func deleteBucket(db *bolt.DB, bucketName string) {
	db.Update(func(tx *bolt.Tx) error {
		log.Printf("Delete bucket %s\n", bucketName)
		err := tx.DeleteBucket([]byte(bucketName))
		if err != nil {
			log.Println(err)
		}
		return err
	})
}

func TestRollback(t *testing.T) {

	t.Run("Rollback should restore upgrade after backup", func(t *testing.T) {
		version := 21
		store := NewTestStore(portainer.PortainerEE, version, true)

		_, err := store.BackupWithOptions(getBackupRestoreOptions(store))
		if err != nil {
			log.Fatal(err)
		}

		// Change the current edition
		err = store.VersionService.StoreDBVersion(version + 10)
		if err != nil {
			log.Fatal(err)
		}

		store.Close()

		err = store.Rollback(true)
		if err != nil {
			t.Logf("Rollback failed: %s", err)
			t.Fail()
			return
		}

		store.Open()

		testVersion(store, version, t)

		store.Close()
	})

	teardown()
}
