package datastore

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	dserrors "github.com/portainer/portainer-ee/api/dataservices/errors"
	portainer "github.com/portainer/portainer/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/rs/zerolog/log"
)

// NewStore initializes a new Store and the associated services
func NewStore(storePath string, fileService portainer.FileService, connection portainer.Connection) *Store {
	return &Store{
		path:        storePath,
		fileService: fileService,
		connection:  connection,
	}
}

// Open opens and initializes the BoltDB database.
func (store *Store) Open() (newStore bool, err error) {
	encryptionReq, err := store.connection.NeedsEncryptionMigration()
	if err != nil {
		return false, err
	}

	if encryptionReq {
		err = store.encryptDB()
		if err != nil {
			return false, err
		}
	}

	err = store.connection.Open()
	if err != nil {
		return newStore, err
	}

	err = store.initServices()
	if err != nil {
		return newStore, err
	}

	// if no settings object exists then we have a new store
	_, err = store.SettingsService.Settings()
	if err != nil {
		if store.IsErrObjectNotFound(err) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (store *Store) Close() error {
	return store.connection.Close()
}

func (store *Store) UpdateTx(fn func(dataservices.DataStoreTx) error) error {
	return store.connection.UpdateTx(func(tx portainer.Transaction) error {
		return fn(&StoreTx{
			store: store,
			tx:    tx,
		})
	})
}

func (store *Store) ViewTx(fn func(dataservices.DataStoreTx) error) error {
	return store.connection.ViewTx(func(tx portainer.Transaction) error {
		return fn(&StoreTx{
			store: store,
			tx:    tx,
		})
	})
}

// BackupTo backs up db to a provided writer.
// It does hot backup and doesn't block other database reads and writes
func (store *Store) BackupTo(w io.Writer) error {
	return store.connection.BackupTo(w)
}

// CheckCurrentEdition checks if current edition is community edition
func (store *Store) CheckCurrentEdition() error {
	edition := portaineree.PortainerCE
	v, _ := store.VersionService.Version()
	if v != nil {
		edition = portaineree.SoftwareEdition(v.Edition)
	}

	if edition == portaineree.PortainerCE {
		return dserrors.ErrWrongDBEdition
	}
	return nil
}

// TODO: move the use of this to dataservices.IsErrObjectNotFound()?
func (store *Store) IsErrObjectNotFound(e error) bool {
	return e == portainerDsErrors.ErrObjectNotFound
}

func (store *Store) Connection() portainer.Connection {
	return store.connection
}

func (store *Store) Rollback(force bool) error {
	return store.connectionRollback(force)
}

func (store *Store) RollbackToCE() error {
	return store.rollbackToCE(false)
}

func (store *Store) encryptDB() error {
	store.connection.SetEncrypted(false)
	err := store.connection.Open()
	if err != nil {
		return err
	}

	err = store.initServices()
	if err != nil {
		return err
	}

	// The DB is not currently encrypted.  First save the encrypted db filename
	oldFilename := store.connection.GetDatabaseFilePath()
	log.Info().Msg("encrypting database")

	// export file path for backup
	exportFilename := path.Join(store.databasePath() + "." + fmt.Sprintf("backup-%d.json", time.Now().Unix()))

	log.Info().Str("filename", exportFilename).Msg("exporting database backup")

	err = store.Export(exportFilename)
	if err != nil {
		log.Error().Str("filename", exportFilename).Err(err).Msg("failed to export")

		return err
	}

	log.Info().Msg("database backup exported")

	// Close existing un-encrypted db so that we can delete the file later
	store.connection.Close()

	// Tell the db layer to create an encrypted db when opened
	store.connection.SetEncrypted(true)
	store.connection.Open()

	// We have to init services before import
	err = store.initServices()
	if err != nil {
		return err
	}

	err = store.Import(exportFilename)
	if err != nil {
		// Remove the new encrypted file that we failed to import
		os.Remove(store.connection.GetDatabaseFilePath())

		log.Fatal().Err(portainerDsErrors.ErrDBImportFailed).Msg("")
	}

	err = os.Remove(oldFilename)
	if err != nil {
		log.Error().Msg("failed to remove the un-encrypted db file")
	}

	err = os.Remove(exportFilename)
	if err != nil {
		log.Error().Msg("failed to remove the json backup file")
	}

	// Close db connection
	store.connection.Close()

	log.Info().Msg("database successfully encrypted")

	return nil
}
