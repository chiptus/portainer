package demo

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

func initDemoUser(
	store dataservices.DataStore,
	cryptoService portainer.CryptoService,
) (portainer.UserID, error) {

	password, err := cryptoService.Hash("tryportainer")
	if err != nil {
		return 0, errors.WithMessage(err, "failed creating password hash")
	}

	admin := &portaineree.User{
		Username: "admin",
		Password: password,
		Role:     portaineree.AdministratorRole,
	}

	err = store.User().Create(admin)
	return admin.ID, errors.WithMessage(err, "failed creating user")
}

func initDemoEndpoints(store dataservices.DataStore) ([]portainer.EndpointID, error) {
	localEndpointId, err := initDemoLocalEndpoint(store)
	if err != nil {
		return nil, errors.WithMessage(err, "failed creating local endpoint")
	}

	// second and third endpoints are going to be created with docker-compose as a part of the demo environment set up.
	// ref: https://github.com/portainer/portainer-demo/blob/master/docker-compose.yml
	return []portainer.EndpointID{localEndpointId, localEndpointId + 1, localEndpointId + 2}, nil
}

func initDemoLocalEndpoint(store dataservices.DataStore) (portainer.EndpointID, error) {
	id := portainer.EndpointID(store.Endpoint().GetNextIdentifier())
	localEndpoint := &portaineree.Endpoint{
		ID:        id,
		Name:      "local",
		URL:       "unix:///var/run/docker.sock",
		PublicURL: "demo.portaineree.io",
		Type:      portaineree.DockerEnvironment,
		GroupID:   portainer.EndpointGroupID(1),
		TLSConfig: portainer.TLSConfiguration{
			TLS: false,
		},
		AuthorizedUsers:    []portainer.UserID{},
		AuthorizedTeams:    []portainer.TeamID{},
		UserAccessPolicies: portainer.UserAccessPolicies{},
		TeamAccessPolicies: portainer.TeamAccessPolicies{},
		TagIDs:             []portainer.TagID{},
		Status:             portainer.EndpointStatusUp,
		Snapshots:          []portainer.DockerSnapshot{},
		Kubernetes:         portaineree.KubernetesDefault(),
	}

	err := store.Endpoint().Create(localEndpoint)
	if err != nil {
		return id, errors.WithMessage(err, "failed creating local endpoint")
	}

	err = store.Snapshot().Create(&portaineree.Snapshot{EndpointID: id})
	if err != nil {
		return id, errors.WithMessage(err, "failed creating snapshot")
	}

	return id, errors.WithMessage(err, "failed creating local endpoint")
}

func initDemoSettings(
	store dataservices.DataStore,
) error {
	settings, err := store.Settings().Settings()
	if err != nil {
		return errors.WithMessage(err, "failed fetching settings")
	}

	settings.EnableTelemetry = false
	settings.LogoURL = ""

	err = store.Settings().UpdateSettings(settings)
	return errors.WithMessage(err, "failed updating settings")
}
