package datastore

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/sirupsen/logrus"
)

var errTempDir = errors.New("can't create a temp dir")

func (store *Store) GetConnection() portainer.Connection {
	return store.connection
}

func MustNewTestStore(init bool) (bool, *Store, func()) {
	newStore, store, teardown, err := NewTestStore(init)
	if err != nil {
		if !errors.Is(err, errTempDir) && teardown != nil {
			teardown()
		}
		log.Fatal(err)
	}

	return newStore, store, teardown
}

func NewTestStore(init bool) (bool, *Store, func(), error) {
	// Creates unique temp directory in a concurrency friendly manner.
	storePath, err := ioutil.TempDir("", "test-store")
	if err != nil {
		return false, nil, nil, errors.Wrap(errTempDir, err.Error())
	}

	fileService, err := filesystem.NewService(storePath, "")
	if err != nil {
		return false, nil, nil, err
	}

	connection, err := database.NewDatabase("boltdb", storePath, []byte("apassphrasewhichneedstobe32bytes"))
	if err != nil {
		panic(err)
	}

	store := NewStore(storePath, fileService, connection)
	newStore, err := store.Open()
	if err != nil {
		return newStore, nil, nil, err
	}

	logrus.Error("Openned")

	if init {
		err = store.Init()
		if err != nil {
			return newStore, nil, nil, err
		}
	}

	logrus.Error("Initialised")

	if newStore {
		// from MigrateData
		store.VersionService.StoreDBVersion(portaineree.DBVersion)
		store.VersionService.StoreEdition(portaineree.PortainerEE)
		if err != nil {
			return newStore, nil, nil, err
		}
	}

	teardown := func() {
		teardown(store, storePath)
	}

	return newStore, store, teardown, nil
}

func teardown(store *Store, storePath string) {
	err := store.Close()
	if err != nil {
		log.Fatalln(err)
	}

	err = os.RemoveAll(storePath)
	if err != nil {
		log.Fatalln(err)
	}
}
