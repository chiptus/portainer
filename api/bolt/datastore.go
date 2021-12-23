package bolt

import (
	"io"
	"path"
	"time"

	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/apikeyrepository"
	"github.com/portainer/portainer-ee/api/bolt/customtemplate"
	"github.com/portainer/portainer-ee/api/bolt/dockerhub"
	"github.com/portainer/portainer-ee/api/bolt/edgegroup"
	"github.com/portainer/portainer-ee/api/bolt/edgejob"
	"github.com/portainer/portainer-ee/api/bolt/edgestack"
	"github.com/portainer/portainer-ee/api/bolt/endpoint"
	"github.com/portainer/portainer-ee/api/bolt/endpointgroup"
	"github.com/portainer/portainer-ee/api/bolt/endpointrelation"
	"github.com/portainer/portainer-ee/api/bolt/errors"
	"github.com/portainer/portainer-ee/api/bolt/extension"
	"github.com/portainer/portainer-ee/api/bolt/helmuserrepository"
	"github.com/portainer/portainer-ee/api/bolt/internal"
	"github.com/portainer/portainer-ee/api/bolt/license"
	"github.com/portainer/portainer-ee/api/bolt/registry"
	"github.com/portainer/portainer-ee/api/bolt/resourcecontrol"
	"github.com/portainer/portainer-ee/api/bolt/role"
	"github.com/portainer/portainer-ee/api/bolt/s3backup"
	"github.com/portainer/portainer-ee/api/bolt/schedule"
	"github.com/portainer/portainer-ee/api/bolt/settings"
	"github.com/portainer/portainer-ee/api/bolt/ssl"
	"github.com/portainer/portainer-ee/api/bolt/stack"
	"github.com/portainer/portainer-ee/api/bolt/tag"
	"github.com/portainer/portainer-ee/api/bolt/team"
	"github.com/portainer/portainer-ee/api/bolt/teammembership"
	"github.com/portainer/portainer-ee/api/bolt/tunnelserver"
	"github.com/portainer/portainer-ee/api/bolt/user"
	"github.com/portainer/portainer-ee/api/bolt/version"
	"github.com/portainer/portainer-ee/api/bolt/webhook"
	portainer "github.com/portainer/portainer/api"
)

var (
	databaseFileName = "portainer.db"
)

// Store defines the implementation of portaineree.DataStore using
// BoltDB as the storage system.
type Store struct {
	path                      string
	connection                *internal.DbConnection
	isNew                     bool
	fileService               portainer.FileService
	CustomTemplateService     *customtemplate.Service
	DockerHubService          *dockerhub.Service
	EdgeGroupService          *edgegroup.Service
	EdgeJobService            *edgejob.Service
	EdgeStackService          *edgestack.Service
	EndpointGroupService      *endpointgroup.Service
	EndpointService           *endpoint.Service
	EndpointRelationService   *endpointrelation.Service
	ExtensionService          *extension.Service
	HelmUserRepositoryService *helmuserrepository.Service
	LicenseService            *license.Service
	RegistryService           *registry.Service
	ResourceControlService    *resourcecontrol.Service
	RoleService               *role.Service
	APIKeyRepositoryService   *apikeyrepository.Service
	S3BackupService           *s3backup.Service
	ScheduleService           *schedule.Service
	SettingsService           *settings.Service
	SSLSettingsService        *ssl.Service
	StackService              *stack.Service
	TagService                *tag.Service
	TeamMembershipService     *teammembership.Service
	TeamService               *team.Service
	TunnelServerService       *tunnelserver.Service
	UserService               *user.Service
	VersionService            *version.Service
	WebhookService            *webhook.Service
}

func (store *Store) version() (int, error) {
	version, err := store.VersionService.DBVersion()
	if err == errors.ErrObjectNotFound {
		version = 0
	}
	return version, err
}

func (store *Store) edition() portaineree.SoftwareEdition {
	edition, err := store.VersionService.Edition()
	if err == errors.ErrObjectNotFound {
		edition = portaineree.PortainerCE
	}
	return edition
}

// NewStore initializes a new Store and the associated services
func NewStore(storePath string, fileService portainer.FileService) *Store {
	return &Store{
		path:        storePath,
		fileService: fileService,
		isNew:       true,
		connection:  &internal.DbConnection{},
	}
}

// Open opens and initializes the BoltDB database.
func (store *Store) Open() error {
	databasePath := path.Join(store.path, databaseFileName)
	db, err := bolt.Open(databasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	store.connection.DB = db

	err = store.initServices()
	if err != nil {
		return err
	}

	// if we have DBVersion in the database then ensure we flag this as NOT a new store
	if _, err := store.VersionService.DBVersion(); err == nil {
		store.isNew = false
	}

	return nil
}

// Close closes the BoltDB database.
// Safe to being called multiple times.
func (store *Store) Close() error {
	if store.connection.DB != nil {
		return store.connection.Close()
	}
	return nil
}

// IsNew returns true if the database was just created and false if it is re-using
// existing data.
func (store *Store) IsNew() bool {
	return store.isNew
}

// BackupTo backs up db to a provided writer.
// It does hot backup and doesn't block other database reads and writes
func (store *Store) BackupTo(w io.Writer) error {
	return store.connection.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}
