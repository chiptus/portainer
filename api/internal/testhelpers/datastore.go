package testhelpers

import (
	"io"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices/errors"
)

type testDatastore struct {
	customTemplate          dataservices.CustomTemplateService
	edgeAsyncCommand        dataservices.EdgeAsyncCommandService
	cloudProvisioning       dataservices.CloudProvisioningService
	edgeGroup               dataservices.EdgeGroupService
	edgeJob                 dataservices.EdgeJobService
	edgeUpdateSchedule      dataservices.EdgeUpdateScheduleService
	edgeStack               dataservices.EdgeStackService
	edgeStackLog            dataservices.EdgeStackLogService
	endpoint                dataservices.EndpointService
	endpointGroup           dataservices.EndpointGroupService
	endpointRelation        dataservices.EndpointRelationService
	enforcement             dataservices.EnforcementService
	fdoProfile              dataservices.FDOProfileService
	helmUserRepository      dataservices.HelmUserRepositoryService
	license                 dataservices.LicenseRepository
	registry                dataservices.RegistryService
	resourceControl         dataservices.ResourceControlService
	apiKeyRepositoryService dataservices.APIKeyRepository
	gitCredential           dataservices.GitCredential
	role                    dataservices.RoleService
	sslSettings             dataservices.SSLSettingsService
	settings                dataservices.SettingsService
	s3backup                dataservices.S3BackupService
	stack                   dataservices.StackService
	tag                     dataservices.TagService
	podSecurity             dataservices.PodSecurityService
	teamMembership          dataservices.TeamMembershipService
	team                    dataservices.TeamService
	tunnelServer            dataservices.TunnelServerService
	user                    dataservices.UserService
	version                 dataservices.VersionService
	webhook                 dataservices.WebhookService
	cloudCredential         dataservices.CloudCredentialService
	connection              portainer.Connection
}

func (d *testDatastore) BackupTo(io.Writer) error { return nil }
func (d *testDatastore) CloudProvisioning() dataservices.CloudProvisioningService {
	return nil
}
func (d *testDatastore) CloudCredential() dataservices.CloudCredentialService {
	return d.cloudCredential
}
func (d *testDatastore) Open() (bool, error)                                { return false, nil }
func (d *testDatastore) Init() error                                        { return nil }
func (d *testDatastore) Close() error                                       { return nil }
func (d *testDatastore) CheckCurrentEdition() error                         { return nil }
func (d *testDatastore) MigrateData() error                                 { return nil }
func (d *testDatastore) Rollback(force bool) error                          { return nil }
func (d *testDatastore) RollbackToCE() error                                { return nil }
func (d *testDatastore) CustomTemplate() dataservices.CustomTemplateService { return d.customTemplate }
func (d *testDatastore) EdgeAsyncCommand() dataservices.EdgeAsyncCommandService {
	return d.edgeAsyncCommand
}
func (d *testDatastore) EdgeGroup() dataservices.EdgeGroupService { return d.edgeGroup }
func (d *testDatastore) EdgeJob() dataservices.EdgeJobService     { return d.edgeJob }
func (d *testDatastore) EdgeUpdateSchedule() dataservices.EdgeUpdateScheduleService {
	return d.edgeUpdateSchedule
}
func (d *testDatastore) EdgeStack() dataservices.EdgeStackService         { return d.edgeStack }
func (d *testDatastore) EdgeStackLog() dataservices.EdgeStackLogService   { return d.edgeStackLog }
func (d *testDatastore) Endpoint() dataservices.EndpointService           { return d.endpoint }
func (d *testDatastore) EndpointGroup() dataservices.EndpointGroupService { return d.endpointGroup }
func (d *testDatastore) Enforcement() dataservices.EnforcementService     { return d.enforcement }
func (d *testDatastore) FDOProfile() dataservices.FDOProfileService {
	return d.fdoProfile
}
func (d *testDatastore) EndpointRelation() dataservices.EndpointRelationService {
	return d.endpointRelation
}
func (d *testDatastore) HelmUserRepository() dataservices.HelmUserRepositoryService {
	return d.helmUserRepository
}
func (d *testDatastore) License() dataservices.LicenseRepository { return d.license }
func (d *testDatastore) Registry() dataservices.RegistryService  { return d.registry }
func (d *testDatastore) ResourceControl() dataservices.ResourceControlService {
	return d.resourceControl
}
func (d *testDatastore) Role() dataservices.RoleService { return d.role }
func (d *testDatastore) APIKeyRepository() dataservices.APIKeyRepository {
	return d.apiKeyRepositoryService
}
func (d *testDatastore) GitCredential() dataservices.GitCredential          { return d.gitCredential }
func (d *testDatastore) S3Backup() dataservices.S3BackupService             { return d.s3backup }
func (d *testDatastore) Settings() dataservices.SettingsService             { return d.settings }
func (d *testDatastore) SSLSettings() dataservices.SSLSettingsService       { return d.sslSettings }
func (d *testDatastore) Stack() dataservices.StackService                   { return d.stack }
func (d *testDatastore) PodSecurity() dataservices.PodSecurityService       { return d.podSecurity }
func (d *testDatastore) Tag() dataservices.TagService                       { return d.tag }
func (d *testDatastore) TeamMembership() dataservices.TeamMembershipService { return d.teamMembership }
func (d *testDatastore) Team() dataservices.TeamService                     { return d.team }
func (d *testDatastore) TunnelServer() dataservices.TunnelServerService     { return d.tunnelServer }
func (d *testDatastore) User() dataservices.UserService                     { return d.user }
func (d *testDatastore) Version() dataservices.VersionService               { return d.version }
func (d *testDatastore) Webhook() dataservices.WebhookService               { return d.webhook }
func (d *testDatastore) Connection() portainer.Connection {
	return d.connection
}

func (d *testDatastore) IsErrObjectNotFound(e error) bool {
	return false
}

func (d *testDatastore) Export(filename string) (err error) {
	return nil
}
func (d *testDatastore) Import(filename string) (err error) {
	return nil
}

type datastoreOption = func(d *testDatastore)

// NewDatastore creates new instance of testDatastore.
// Will apply options before returning, opts will be applied from left to right.
func NewDatastore(options ...datastoreOption) *testDatastore {
	conn, _ := database.NewDatabase("boltdb", "", nil)
	d := testDatastore{connection: conn}
	for _, o := range options {
		o(&d)
	}
	return &d
}

type stubSettingsService struct {
	settings *portaineree.Settings
}

func (s *stubSettingsService) BucketName() string { return "settings" }

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
	return func(d *testDatastore) {
		d.settings = &stubSettingsService{
			settings: settings,
		}
	}
}

type stubUserService struct {
	users []portaineree.User
}

func (s *stubUserService) BucketName() string                                        { return "users" }
func (s *stubUserService) User(ID portaineree.UserID) (*portaineree.User, error)     { return nil, nil }
func (s *stubUserService) UserByUsername(username string) (*portaineree.User, error) { return nil, nil }
func (s *stubUserService) Users() ([]portaineree.User, error)                        { return s.users, nil }
func (s *stubUserService) UsersByRole(role portaineree.UserRole) ([]portaineree.User, error) {
	return s.users, nil
}
func (s *stubUserService) Create(user *portaineree.User) error                            { return nil }
func (s *stubUserService) UpdateUser(ID portaineree.UserID, user *portaineree.User) error { return nil }
func (s *stubUserService) DeleteUser(ID portaineree.UserID) error                         { return nil }

// WithUsers testDatastore option that will instruct testDatastore to return provided users
func WithUsers(us []portaineree.User) datastoreOption {
	return func(d *testDatastore) {
		d.user = &stubUserService{users: us}
	}
}

type stubEdgeJobService struct {
	jobs []portaineree.EdgeJob
}

func (s *stubEdgeJobService) BucketName() string                       { return "edgejobs" }
func (s *stubEdgeJobService) EdgeJobs() ([]portaineree.EdgeJob, error) { return s.jobs, nil }
func (s *stubEdgeJobService) EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error) {
	return nil, nil
}
func (s *stubEdgeJobService) Create(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	return nil
}
func (s *stubEdgeJobService) UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	return nil
}
func (s *stubEdgeJobService) DeleteEdgeJob(ID portaineree.EdgeJobID) error { return nil }
func (s *stubEdgeJobService) GetNextIdentifier() int                       { return 0 }

// WithEdgeJobs option will instruct testDatastore to return provided jobs
func WithEdgeJobs(js []portaineree.EdgeJob) datastoreOption {
	return func(d *testDatastore) {
		d.edgeJob = &stubEdgeJobService{jobs: js}
	}
}

type stubEndpointRelationService struct {
	relations []portaineree.EndpointRelation
}

func (s *stubEndpointRelationService) BucketName() string { return "endpoint_relations" }
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
func (s *stubEndpointRelationService) Create(EndpointRelation *portaineree.EndpointRelation) error {
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

// WithEndpointRelations option will instruct testDatastore to return provided jobs
func WithEndpointRelations(relations []portaineree.EndpointRelation) datastoreOption {
	return func(d *testDatastore) {
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
	return func(d *testDatastore) {
		d.s3backup = &stubS3BackupService{
			status:   status,
			settings: settings,
		}
	}
}

type stubEndpointService struct {
	endpoints []portaineree.Endpoint
}

func (s *stubEndpointService) BucketName() string { return "endpoints" }
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

func (s *stubEndpointService) Create(endpoint *portaineree.Endpoint) error {
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

// WithEndpoints option will instruct testDatastore to return provided environments(endpoints)
func WithEndpoints(endpoints []portaineree.Endpoint) datastoreOption {
	return func(d *testDatastore) {
		d.endpoint = &stubEndpointService{endpoints: endpoints}
	}
}
