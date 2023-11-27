package datastore

import (
	"fmt"
	"os"
	"path"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database/models"

	"github.com/rs/zerolog/log"
)

var backupDefaults = struct {
	backupDir string
	commonDir string
	editions  []string
}{
	"backups",
	"common",
	[]string{"CE", "BE", "EE"},
}

//
// Backup Helpers
//

// createBackupFolders create initial folders for backups
func (store *Store) createBackupFolders() {
	// create common dir
	commonDir := store.commonBackupDir()
	if exists, _ := store.fileService.FileExists(commonDir); !exists {
		if err := os.MkdirAll(commonDir, 0700); err != nil {
			log.Error().Err(err).Msg("error while creating common backup folder")
		}
	}

	// create backup folders for editions
	for _, e := range backupDefaults.editions {
		p := path.Join(store.path, backupDefaults.backupDir, e)
		if exists, _ := store.fileService.FileExists(p); !exists {
			err := os.MkdirAll(p, 0700)
			if err != nil {
				log.Error().Err(err).Msg("error while creating edition backup folders")
			}
		}
	}
}

// getBackupRestoreOptions returns options to store db at common backup dir location; used by:
// - db backup prior to version upgrade
// - db rollback
func getBackupRestoreOptions(backupDir string) *BackupOptions {
	return &BackupOptions{
		BackupDir:      backupDir,
		BackupFileName: beforePortainerVersionUpgradeBackup,
	}
}

func (store *Store) databasePath() string {
	return store.connection.GetDatabaseFilePath()
}

func (store *Store) commonBackupDir() string {
	return path.Join(store.connection.GetStorePath(), backupDefaults.backupDir, backupDefaults.commonDir)
}

func (store *Store) editionBackupDir(edition portainer.SoftwareEdition) string {
	return path.Join(store.path, backupDefaults.backupDir, edition.GetEditionLabel())
}

func (store *Store) copyDBFile(from string, to string) error {
	log.Info().Str("from", from).Str("to", to).Msg("copying DB file")

	err := store.fileService.Copy(from, to, true)
	if err != nil {
		log.Error().Err(err).Msg("failed")
	}

	return err
}

// BackupOptions provide a helper to inject backup options
type BackupOptions struct {
	Edition        portainer.SoftwareEdition
	Version        string
	BackupDir      string
	BackupFileName string
	BackupPath     string
}

func (store *Store) setDefaultBackupOptions(options *BackupOptions) *BackupOptions {
	if options == nil {
		options = &BackupOptions{}
	}

	v, err := store.VersionService.Version()
	if options.Version == "" && err == nil {
		options.Version = v.SchemaVersion
	}
	if options.Edition == 0 {
		if err != nil {
			options.Edition = portaineree.PortainerEE
		} else {
			options.Edition = portainer.SoftwareEdition(v.Edition)
		}
	}

	if options.BackupDir == "" {
		options.BackupDir = store.editionBackupDir(options.Edition)
	}
	if options.BackupFileName == "" {
		options.BackupFileName = fmt.Sprintf("%s.%s.%s", store.connection.GetDatabaseFileName(), options.Version, time.Now().Format("20060102150405"))
	}
	if options.BackupPath == "" {
		options.BackupPath = path.Join(options.BackupDir, options.BackupFileName)
	}
	return options
}

func (store *Store) listEditionBackups(edition portainer.SoftwareEdition) ([]string, error) {
	var fileNames = []string{}

	files, err := os.ReadDir(store.editionBackupDir(edition))

	if err != nil {
		log.Error().Err(err).Msg("error while retrieving backup files")

		return fileNames, err
	}

	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}

	return fileNames, nil
}

func (store *Store) LatestEditionBackup() (string, error) {
	var edition portainer.SoftwareEdition
	v, err := store.VersionService.Version()
	if err != nil {
		edition = portaineree.PortainerEE
	} else {
		edition = portainer.SoftwareEdition(v.Edition)
	}

	files, err := store.listEditionBackups(edition)
	if err != nil {
		log.Error().Err(err).Msg("error while retrieving backup files")
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no backup files found for Portainer %s", edition.GetEditionLabel())
	}

	return files[len(files)-1], nil
}

// BackupWithOptions backup current database with options
func (store *Store) backupWithOptions(options *BackupOptions) (string, error) {
	log.Info().Msg("creating DB backup")

	store.createBackupFolders()

	options = store.setDefaultBackupOptions(options)
	dbPath := store.databasePath()

	if err := store.Close(); err != nil {
		return options.BackupPath, fmt.Errorf(
			"error closing datastore before creating backup: %w",
			err,
		)
	}

	if err := store.copyDBFile(dbPath, options.BackupPath); err != nil {
		return options.BackupPath, err
	}

	if _, err := store.Open(); err != nil {
		return options.BackupPath, fmt.Errorf(
			"error opening datastore after creating backup: %w",
			err,
		)
	}

	return options.BackupPath, nil
}

// Backup current database with default options
func (store *Store) Backup(version *models.Version) (string, error) {
	if version == nil {
		return store.backupWithOptions(nil)
	}

	return store.backupWithOptions(&BackupOptions{
		Version: version.SchemaVersion,
		Edition: portainer.SoftwareEdition(version.Edition),
	})
}

// RestoreWithOptions previously saved backup for the current Edition  with options
// Restore strategies:
// - default: restore latest from current edition
// - restore a specific
func (store *Store) restoreWithOptions(options *BackupOptions) error {
	options = store.setDefaultBackupOptions(options)

	// Check if backup file exist before restoring
	_, err := os.Stat(options.BackupPath)
	if os.IsNotExist(err) {
		log.Error().Str("path", options.BackupPath).Err(err).Msg("backup file to restore does not exist")
		return err
	}

	err = store.Close()
	if err != nil {
		log.Error().Err(err).Msg("error while closing store before restore")
		return err
	}

	log.Info().Msg("restoring DB backup")
	err = store.copyDBFile(options.BackupPath, store.databasePath())
	if err != nil {
		return err
	}

	_, err = store.Open()
	return err
}

// Restore previously saved backup for the current Edition  with default options
func (store *Store) Restore() error {
	var options = &BackupOptions{}
	var err error
	options.BackupFileName, err = store.LatestEditionBackup()
	if err != nil {
		return err
	}

	return store.restoreWithOptions(options)
}

// RemoveWithOptions removes backup database based on supplied options
func (store *Store) removeWithOptions(options *BackupOptions) error {
	log.Info().Msg("removing DB backup")

	options = store.setDefaultBackupOptions(options)
	_, err := os.Stat(options.BackupPath)

	if os.IsNotExist(err) {
		log.Error().Str("path", options.BackupPath).Err(err).Msg("backup file to remove does not exist")

		return err
	}

	log.Info().Str("path", options.BackupPath).Msg("removing DB file")
	err = os.Remove(options.BackupPath)
	if err != nil {
		log.Error().Err(err).Msg("failed")

		return err
	}

	return nil
}
