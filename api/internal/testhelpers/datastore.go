package testhelpers

import (
	"io"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/errors"
)

type datastore struct {
	customTemplate          portaineree.CustomTemplateService
	edgeGroup               portaineree.EdgeGroupService
	edgeJob                 portaineree.EdgeJobService
	edgeStack               portaineree.EdgeStackService
	endpoint                portaineree.EndpointService
	endpointGroup           portaineree.EndpointGroupService
	endpointRelation        portaineree.EndpointRelationService
	helmUserRepository      portaineree.HelmUserRepositoryService
	license                 portaineree.LicenseRepository
	registry                portaineree.RegistryService
	resourceControl         portaineree.ResourceControlService
	apiKeyRepositoryService portaineree.APIKeyRepository
	role                    portaineree.RoleService
	sslSettings             portaineree.SSLSettingsService
	settings                portaineree.SettingsService
	s3backup                portaineree.S3BackupService
	stack                   portaineree.StackService
	tag                     portaineree.TagService
	teamMembership          portaineree.TeamMembershipService
	team                    portaineree.TeamService
	tunnelServer            portaineree.TunnelServerService
	user                    portaineree.UserService
	version                 portaineree.VersionService
	webhook                 portaineree.WebhookService
}

func (d *datastore) BackupTo(io.Writer) error                              { return nil }
func (d *datastore) Open() error                                           { return nil }
func (d *datastore) Init() error                                           { return nil }
func (d *datastore) Close() error                                          { return nil }
func (d *datastore) IsNew() bool                                           { return false }
func (d *datastore) MigrateData(force bool) error                          { return nil }
func (d *datastore) Rollback(force bool) error                             { return nil }
func (d *datastore) RollbackToCE() error                                   { return nil }
func (d *datastore) CustomTemplate() portaineree.CustomTemplateService     { return d.customTemplate }
func (d *datastore) EdgeGroup() portaineree.EdgeGroupService               { return d.edgeGroup }
func (d *datastore) EdgeJob() portaineree.EdgeJobService                   { return d.edgeJob }
func (d *datastore) EdgeStack() portaineree.EdgeStackService               { return d.edgeStack }
func (d *datastore) Endpoint() portaineree.EndpointService                 { return d.endpoint }
func (d *datastore) EndpointGroup() portaineree.EndpointGroupService       { return d.endpointGroup }
func (d *datastore) EndpointRelation() portaineree.EndpointRelationService { return d.endpointRelation }
func (d *datastore) HelmUserRepository() portaineree.HelmUserRepositoryService {
	return d.helmUserRepository
}
func (d *datastore) License() portaineree.LicenseRepository              { return d.license }
func (d *datastore) Registry() portaineree.RegistryService               { return d.registry }
func (d *datastore) ResourceControl() portaineree.ResourceControlService { return d.resourceControl }
func (d *datastore) Role() portaineree.RoleService                       { return d.role }
func (d *datastore) APIKeyRepository() portaineree.APIKeyRepository {
	return d.apiKeyRepositoryService
}
func (d *datastore) S3Backup() portaineree.S3BackupService             { return d.s3backup }
func (d *datastore) Settings() portaineree.SettingsService             { return d.settings }
func (d *datastore) SSLSettings() portaineree.SSLSettingsService       { return d.sslSettings }
func (d *datastore) Stack() portaineree.StackService                   { return d.stack }
func (d *datastore) Tag() portaineree.TagService                       { return d.tag }
func (d *datastore) TeamMembership() portaineree.TeamMembershipService { return d.teamMembership }
func (d *datastore) Team() portaineree.TeamService                     { return d.team }
func (d *datastore) TunnelServer() portaineree.TunnelServerService     { return d.tunnelServer }
func (d *datastore) User() portaineree.UserService                     { return d.user }
func (d *datastore) Version() portaineree.VersionService               { return d.version }
func (d *datastore) Webhook() portaineree.WebhookService               { return d.webhook }

type datastoreOption = func(d *datastore)

// NewDatastore creates new instance of datastore.
// Will apply options before returning, opts will be applied from left to right.
func NewDatastore(options ...datastoreOption) *datastore {
	d := datastore{}
	for _, o := range options {
		o(&d)
	}
	return &d
}

type stubSettingsService struct {
	settings *portaineree.Settings
}

func (s *stubSettingsService) Settings() (*portaineree.Settings, error) {
	return s.settings, nil
}
func (s *stubSettingsService) UpdateSettings(settings *portaineree.Settings) error {
	s.settings = settings
	return nil
}
func (s *stubSettingsService) IsFeatureFlagEnabled(feature portaineree.Feature) bool {
	return false
}
func WithSettingsService(settings *portaineree.Settings) datastoreOption {
	return func(d *datastore) {
		d.settings = &stubSettingsService{settings: settings}
	}
}

type stubUserService struct {
	users []portaineree.User
}

func (s *stubUserService) User(ID portaineree.UserID) (*portaineree.User, error)     { return nil, nil }
func (s *stubUserService) UserByUsername(username string) (*portaineree.User, error) { return nil, nil }
func (s *stubUserService) Users() ([]portaineree.User, error)                        { return s.users, nil }
func (s *stubUserService) UsersByRole(role portaineree.UserRole) ([]portaineree.User, error) {
	return s.users, nil
}
func (s *stubUserService) CreateUser(user *portaineree.User) error                        { return nil }
func (s *stubUserService) UpdateUser(ID portaineree.UserID, user *portaineree.User) error { return nil }
func (s *stubUserService) DeleteUser(ID portaineree.UserID) error                         { return nil }

// WithUsers datastore option that will instruct datastore to return provided users
func WithUsers(us []portaineree.User) datastoreOption {
	return func(d *datastore) {
		d.user = &stubUserService{users: us}
	}
}

type stubEdgeJobService struct {
	jobs []portaineree.EdgeJob
}

func (s *stubEdgeJobService) EdgeJobs() ([]portaineree.EdgeJob, error) { return s.jobs, nil }
func (s *stubEdgeJobService) EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error) {
	return nil, nil
}
func (s *stubEdgeJobService) CreateEdgeJob(edgeJob *portaineree.EdgeJob) error { return nil }
func (s *stubEdgeJobService) UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	return nil
}
func (s *stubEdgeJobService) DeleteEdgeJob(ID portaineree.EdgeJobID) error { return nil }
func (s *stubEdgeJobService) GetNextIdentifier() int                       { return 0 }

// WithEdgeJobs option will instruct datastore to return provided jobs
func WithEdgeJobs(js []portaineree.EdgeJob) datastoreOption {
	return func(d *datastore) {
		d.edgeJob = &stubEdgeJobService{jobs: js}
	}
}

type stubEndpointRelationService struct {
	relations []portaineree.EndpointRelation
}

func (s *stubEndpointRelationService) EndpointRelations() ([]portaineree.EndpointRelation, error) {
	return s.relations, nil
}
func (s *stubEndpointRelationService) EndpointRelation(ID portaineree.EndpointID) (*portaineree.EndpointRelation, error) {
	for _, relation := range s.relations {
		if relation.EndpointID == ID {
			return &relation, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}
func (s *stubEndpointRelationService) CreateEndpointRelation(EndpointRelation *portaineree.EndpointRelation) error {
	return nil
}
func (s *stubEndpointRelationService) UpdateEndpointRelation(ID portaineree.EndpointID, relation *portaineree.EndpointRelation) error {
	for i, r := range s.relations {
		if r.EndpointID == ID {
			s.relations[i] = *relation
		}
	}

	return nil
}
func (s *stubEndpointRelationService) DeleteEndpointRelation(ID portaineree.EndpointID) error {
	return nil
}
func (s *stubEndpointRelationService) GetNextIdentifier() int { return 0 }

// WithEndpointRelations option will instruct datastore to return provided jobs
func WithEndpointRelations(relations []portaineree.EndpointRelation) datastoreOption {
	return func(d *datastore) {
		d.endpointRelation = &stubEndpointRelationService{relations: relations}
	}
}

type stubS3BackupService struct {
	status   *portaineree.S3BackupStatus
	settings *portaineree.S3BackupSettings
}

func (s *stubS3BackupService) GetStatus() (portaineree.S3BackupStatus, error) { return *s.status, nil }
func (s *stubS3BackupService) DropStatus() error {
	*s.status = portaineree.S3BackupStatus{}
	return nil
}
func (s *stubS3BackupService) UpdateStatus(status portaineree.S3BackupStatus) error {
	s.status = &status
	return nil
}
func (s *stubS3BackupService) UpdateSettings(settings portaineree.S3BackupSettings) error {
	*s.settings = settings
	return nil
}
func (s *stubS3BackupService) GetSettings() (portaineree.S3BackupSettings, error) {
	return *s.settings, nil
}

// WithS3BackupService option will instruct datastore to use provide status and settins
func WithS3BackupService(status *portaineree.S3BackupStatus, settings *portaineree.S3BackupSettings) datastoreOption {
	return func(d *datastore) {
		d.s3backup = &stubS3BackupService{
			status:   status,
			settings: settings,
		}
	}
}

type stubEndpointService struct {
	endpoints []portaineree.Endpoint
}

func (s *stubEndpointService) Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error) {
	for _, endpoint := range s.endpoints {
		if endpoint.ID == ID {
			return &endpoint, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubEndpointService) Endpoints() ([]portaineree.Endpoint, error) {
	return s.endpoints, nil
}

func (s *stubEndpointService) CreateEndpoint(endpoint *portaineree.Endpoint) error {
	s.endpoints = append(s.endpoints, *endpoint)

	return nil
}

func (s *stubEndpointService) UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error {
	for i, e := range s.endpoints {
		if e.ID == ID {
			s.endpoints[i] = *endpoint
		}
	}

	return nil
}

func (s *stubEndpointService) DeleteEndpoint(ID portaineree.EndpointID) error {
	endpoints := []portaineree.Endpoint{}

	for _, endpoint := range s.endpoints {
		if endpoint.ID != ID {
			endpoints = append(endpoints, endpoint)
		}
	}

	s.endpoints = endpoints

	return nil
}

func (s *stubEndpointService) Synchronize(toCreate []*portaineree.Endpoint, toUpdate []*portaineree.Endpoint, toDelete []*portaineree.Endpoint) error {
	panic("not implemented")
}

func (s *stubEndpointService) GetNextIdentifier() int {
	return len(s.endpoints)
}

// WithEndpoints option will instruct datastore to return provided environments(endpoints)
func WithEndpoints(endpoints []portaineree.Endpoint) datastoreOption {
	return func(d *datastore) {
		d.endpoint = &stubEndpointService{endpoints: endpoints}
	}
}
