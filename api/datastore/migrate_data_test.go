package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/boltdb"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/rs/zerolog/log"
)

func TestMigrateData(t *testing.T) {
	tests := []struct {
		testName string
		srcPath  string
		wantPath string
	}{
		{
			testName: "migrate version 34 to latest",
			srcPath:  "test_data/input_34.json",
			wantPath: "test_data/output_34_to_latest.json",
		},
		{
			testName: "migrate version 31 to latest",
			srcPath:  "test_data/input_31.json",
			wantPath: "test_data/output_31_to_latest.json",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := migrateDBTestHelper(t, test.srcPath, test.wantPath)
			if err != nil {
				t.Errorf(
					"Failed migrating mock database %v: %v",
					test.srcPath,
					err,
				)
			}
		})
	}

	t.Run("MigrateData for New Store & Re-Open Check", func(t *testing.T) {
		newStore, store, teardown := MustNewTestStore(t, true, false)
		defer teardown()

		if !newStore {
			t.Error("Expect a new DB")
		}

		testVersion(store, portaineree.APIVersion, t)
		store.Close()

		newStore, _ = store.Open()
		if newStore {
			t.Error("Expect store to NOT be new DB")
		}
	})

	t.Run("Error in MigrateData should restore backup before MigrateData", func(t *testing.T) {
		_, store, teardown := MustNewTestStore(t, false, false)
		defer teardown()

		v := models.Version{
			SchemaVersion: "2.14",
		}

		store.VersionService.UpdateVersion(&v)

		store.MigrateData()

		testVersion(store, v.SchemaVersion, t)
	})

	t.Run("MigrateData should create backup file upon update", func(t *testing.T) {
		_, store, teardown := MustNewTestStore(t, true, false)
		defer teardown()

		store.VersionService.UpdateVersion(&models.Version{SchemaVersion: "1.0"})
		store.MigrateData()

		backupfile, err := store.LatestEditionBackup()
		if err != nil && backupfile == "" {
			t.Errorf("Backup file should exist")
		}
	})

	t.Run("MigrateData should fail to create backup if database file is set to updating", func(t *testing.T) {
		_, store, teardown := MustNewTestStore(t, true, false)
		defer teardown()

		store.VersionService.StoreIsUpdating(true)

		store.MigrateData()

		// If you get an error, it usually means that the backup folder doesn't exist (no backups). Expected!
		// If the backup file is not blank, then it means a backup was created.  We don't want that because we
		// only create a backup when the version changes.
		backupfile, err := store.LatestEditionBackup()
		if err == nil && backupfile != "" {
			t.Errorf("Backup file should not exist for dirty database")
		}
	})

	t.Run("MigrateData should not create backup on startup if portainer version matches db", func(t *testing.T) {
		_, store, teardown := MustNewTestStore(t, true, false)
		defer teardown()

		store.MigrateData()

		// If you get an error, it usually means that the backup folder doesn't exist (no backups). Expected!
		// If the backup file is not blank, then it means a backup was created.  We don't want that because we
		// only create a backup when the version changes.
		backupfile, err := store.LatestEditionBackup()
		if err == nil && backupfile != "" {
			t.Errorf("Backup file should not exist for dirty database")
		}
	})
}

func Test_getBackupRestoreOptions(t *testing.T) {
	_, store, teardown := MustNewTestStore(t, true, false)
	defer teardown()

	options := getBackupRestoreOptions(store.commonBackupDir())

	wantDir := store.commonBackupDir()
	if !strings.HasSuffix(options.BackupDir, wantDir) {
		log.Fatal().Str("got", options.BackupDir).Str("want", wantDir).Msg("incorrect backup dir")
	}

	wantFilename := "portainer.db.bak"
	if options.BackupFileName != wantFilename {
		log.Fatal().Str("got", options.BackupFileName).Str("want", wantFilename).Msg("incorrect backup file")
	}
}

func TestRollback(t *testing.T) {
	t.Run("Rollback should restore upgrade after backup", func(t *testing.T) {
		version := "2.11"

		v := models.Version{
			SchemaVersion: version,
		}

		_, store, teardown := MustNewTestStore(t, false, false)
		defer teardown()
		store.VersionService.UpdateVersion(&v)

		_, err := store.backupWithOptions(getBackupRestoreOptions(store.commonBackupDir()))
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}

		v.SchemaVersion = "2.14"
		// Change the current edition
		err = store.VersionService.UpdateVersion(&v)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}

		err = store.Rollback(true)
		if err != nil {
			t.Logf("Rollback failed: %s", err)
			t.Fail()
			return
		}

		store.Open()
		testVersion(store, version, t)
	})
}

// migrateDBTestHelper loads a json representation of a bolt database from srcPath,
// parses it into a database, runs a migration on that database, and then
// compares it with an expected output database.
func migrateDBTestHelper(t *testing.T, srcPath, wantPath string) error {
	srcJSON, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed loading source JSON file %v: %v", srcPath, err)
	}

	// Parse source json to db.
	_, store, teardown := MustNewTestStore(t, false, false)
	defer teardown()

	// need to remove new automatic addition of new version bucket before
	// importing the old one to properly test the version bucket migration
	store.connection.DeleteObject("version", []byte("VERSION"))

	err = importJSON(t, bytes.NewReader(srcJSON), store)
	if err != nil {
		return err
	}

	// Run the actual migrations on our input database.
	err = store.MigrateData()
	if err != nil {
		return err
	}

	// Assert that our database connection is using bolt so we can call
	// exportJson rather than ExportRaw. The exportJson function allows us to
	// strip out the metadata which we don't want for our tests.
	// TODO: update connection interface in CE to allow us to use ExportRaw and pass meta false
	err = store.connection.Close()
	if err != nil {
		t.Fatalf("err closing bolt connection: %v", err)
	}
	con, ok := store.connection.(*boltdb.DbConnection)
	if !ok {
		t.Fatalf("backing database is not using boltdb, but the migrations test requires it")
	}

	// Convert database back to json.
	databasePath := con.GetDatabaseFilePath()
	if _, err := os.Stat(databasePath); err != nil {
		return fmt.Errorf("stat on %s failed: %s", databasePath, err)
	}

	gotJSON, err := con.ExportJSON(databasePath, false)
	if err != nil {
		t.Logf(
			"failed re-exporting database %s to JSON: %v",
			store.path,
			err,
		)
	}

	wantJSON, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("failed loading want JSON file %v: %v", wantPath, err)
	}

	// Compare the result we got with the one we wanted.
	if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
		gotPath := filepath.Join(os.TempDir(), "portainer-migrator-test-fail.json")
		os.WriteFile(
			gotPath,
			gotJSON,
			0600,
		)
		t.Errorf(
			"migrate data from %s to %s failed\nwrote migrated input to %s\nmismatch (-want +got):\n%s",
			srcPath,
			wantPath,
			gotPath,
			diff,
		)
	}
	return nil
}

// importJSON reads input JSON and commits it to a portainer datastore.Store.
// Errors are logged with the testing package.
func importJSON(t *testing.T, r io.Reader, store *Store) error {
	objects := make(map[string]interface{})

	// Parse json into map of objects.
	d := json.NewDecoder(r)
	d.UseNumber()
	err := d.Decode(&objects)
	if err != nil {
		return err
	}

	// Get database connection from store.
	con := store.connection

	for k, v := range objects {
		switch k {
		case "version":
			versions, ok := v.(map[string]interface{})
			if !ok {
				t.Logf("failed casting %s to map[string]interface{}", k)
			}

			dbVersion, ok := versions["DB_VERSION"]
			if !ok {
				t.Logf("failed getting DB_VERSION from %s", k)
			}

			numDBVersion, ok := dbVersion.(json.Number)
			if !ok {
				t.Logf("failed parsing DB_VERSION as json number from %s", k)
			}

			intDBVersion, err := numDBVersion.Int64()
			if err != nil {
				t.Logf("failed casting %v to int: %v", numDBVersion, intDBVersion)
			}

			err = con.CreateObjectWithStringId(
				k,
				[]byte("DB_VERSION"),
				int(intDBVersion),
			)
			if err != nil {
				t.Logf("failed writing DB_VERSION in %s: %v", k, err)
			}

			instanceID, ok := versions["INSTANCE_ID"]
			if !ok {
				t.Logf("failed getting INSTANCE_ID from %s", k)
			}

			err = con.CreateObjectWithStringId(
				k,
				[]byte("INSTANCE_ID"),
				instanceID,
			)
			if err != nil {
				t.Logf("failed writing INSTANCE_ID in %s: %v", k, err)
			}

			// Edition doesn't existing in CE. But is present in EE
			edition, ok := versions["EDITION"]
			if ok {
				err := con.CreateObjectWithStringId(
					k,
					[]byte("EDITION"),
					edition,
				)
				if err != nil {
					t.Logf("failed writing EDITION in %s: %v", k, err)
				}
			}

		case "dockerhub":
			obj, ok := v.([]interface{})
			if !ok {
				t.Logf("failed to cast %s to []interface{}", k)
			}
			err := con.CreateObjectWithStringId(
				k,
				[]byte("DOCKERHUB"),
				obj[0],
			)
			if err != nil {
				t.Logf("failed writing DOCKERHUB in %s: %v", k, err)
			}

		case "ssl":
			obj, ok := v.(map[string]interface{})
			if !ok {
				t.Logf("failed to case %s to map[string]interface{}", k)
			}
			err := con.CreateObjectWithStringId(
				k,
				[]byte("SSL"),
				obj,
			)
			if err != nil {
				t.Logf("failed writing SSL in %s: %v", k, err)
			}

		case "settings":
			obj, ok := v.(map[string]interface{})
			if !ok {
				t.Logf("failed to case %s to map[string]interface{}", k)
			}
			err := con.CreateObjectWithStringId(
				k,
				[]byte("SETTINGS"),
				obj,
			)
			if err != nil {
				t.Logf("failed writing SETTINGS in %s: %v", k, err)
			}

		case "tunnel_server":
			obj, ok := v.(map[string]interface{})
			if !ok {
				t.Logf("failed to case %s to map[string]interface{}", k)
			}
			err := con.CreateObjectWithStringId(
				k,
				[]byte("INFO"),
				obj,
			)
			if err != nil {
				t.Logf("failed writing INFO in %s: %v", k, err)
			}

		default:
			objlist, ok := v.([]interface{})
			if !ok {
				t.Logf("failed to cast %s to []interface{}", k)
			}

			for _, obj := range objlist {
				value, ok := obj.(map[string]interface{})
				if !ok {
					t.Logf("failed to cast %v to map[string]interface{}", obj)
				} else {
					var ok bool
					var id interface{}
					switch k {
					case "endpoint_relations":
						// TODO: need to make into an int, then do that weird
						// stringification
						id, ok = value["EndpointID"]
					default:
						id, ok = value["Id"]
					}
					if !ok {
						// endpoint_relations: EndpointID
						t.Logf("missing Id field: %s", k)
						id = "error"
					}
					n, ok := id.(json.Number)
					if !ok {
						t.Logf("failed to cast %v to json.Number in %s", id, k)
					} else {
						key, err := n.Int64()
						if err != nil {
							t.Logf("failed to cast %v to int in %s", n, k)
						} else {
							err := con.CreateObjectWithId(
								k,
								int(key),
								value,
							)
							if err != nil {
								t.Logf("failed writing %v in %s: %v", key, k, err)
							}
						}
					}
				}
			}
		}
	}

	return nil
}
