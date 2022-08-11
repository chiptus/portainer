package snapshot

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/agent"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
)

// Service represents a service to manage environment(endpoint) snapshots.
// It provides an interface to start background snapshots as well as
// specific Docker/Kubernetes environment(endpoint) snapshot methods.
type Service struct {
	dataStore                 dataservices.DataStore
	snapshotIntervalCh        chan time.Duration
	snapshotIntervalInSeconds float64
	dockerSnapshotter         portaineree.DockerSnapshotter
	kubernetesSnapshotter     portaineree.KubernetesSnapshotter
	nomadSnapshotter          portaineree.NomadSnapshotter
	shutdownCtx               context.Context
}

// NewService creates a new instance of a service
func NewService(
	snapshotIntervalFromFlag string,
	dataStore dataservices.DataStore,
	dockerSnapshotter portaineree.DockerSnapshotter,
	kubernetesSnapshotter portaineree.KubernetesSnapshotter,
	nomadSnapshotter portaineree.NomadSnapshotter,
	shutdownCtx context.Context,
) (*Service, error) {
	interval, err := parseSnapshotFrequency(snapshotIntervalFromFlag, dataStore)
	if err != nil {
		return nil, err
	}

	return &Service{
		dataStore:                 dataStore,
		snapshotIntervalCh:        make(chan time.Duration),
		snapshotIntervalInSeconds: interval,
		dockerSnapshotter:         dockerSnapshotter,
		kubernetesSnapshotter:     kubernetesSnapshotter,
		nomadSnapshotter:          nomadSnapshotter,
		shutdownCtx:               shutdownCtx,
	}, nil
}

func parseSnapshotFrequency(snapshotInterval string, dataStore dataservices.DataStore) (float64, error) {
	if snapshotInterval == "" {
		settings, err := dataStore.Settings().Settings()
		if err != nil {
			return 0, err
		}
		snapshotInterval = settings.SnapshotInterval
		if snapshotInterval == "" {
			snapshotInterval = portaineree.DefaultSnapshotInterval
		}
	}
	snapshotFrequency, err := time.ParseDuration(snapshotInterval)
	if err != nil {
		return 0, err
	}
	return snapshotFrequency.Seconds(), nil
}

// Start will start a background routine to execute periodic snapshots of environments(endpoints)
func (service *Service) Start() {
	go service.startSnapshotLoop()
}

// SetSnapshotInterval sets the snapshot interval and resets the service
func (service *Service) SetSnapshotInterval(snapshotInterval string) error {
	interval, err := time.ParseDuration(snapshotInterval)
	if err != nil {
		return err
	}

	service.snapshotIntervalCh <- interval

	return nil
}

// SupportDirectSnapshot checks whether an environment(endpoint) can be used to trigger a direct a snapshot.
// It is mostly true for all environments(endpoints) except Edge and Azure environments(endpoints).
func SupportDirectSnapshot(endpoint *portaineree.Endpoint) bool {
	switch endpoint.Type {
	case portaineree.EdgeAgentOnDockerEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment, portaineree.AzureEnvironment, portaineree.EdgeAgentOnNomadEnvironment:
		return false
	}
	return true
}

// SnapshotEndpoint will create a snapshot of the environment(endpoint) based on the environment(endpoint) type.
// If the snapshot is a success, it will be associated to the environment(endpoint).
func (service *Service) SnapshotEndpoint(endpoint *portaineree.Endpoint) error {
	if endpoint.Type == portaineree.AgentOnDockerEnvironment || endpoint.Type == portaineree.AgentOnKubernetesEnvironment {
		var err error
		var tlsConfig *tls.Config
		if endpoint.TLSConfig.TLS {
			tlsConfig, err = crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
			if err != nil {
				return err
			}
		}

		_, version, err := agent.GetAgentVersionAndPlatform(endpoint.URL, tlsConfig)
		if err != nil {
			return err
		}

		endpoint.Agent.Version = version
	}

	switch endpoint.Type {
	case portaineree.AzureEnvironment:
		return nil
	case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment, portaineree.EdgeAgentOnKubernetesEnvironment:
		return service.snapshotKubernetesEndpoint(endpoint)
	case portaineree.EdgeAgentOnNomadEnvironment:
		return service.snapshotNomadEndpoint(endpoint)
	}

	return service.snapshotDockerEndpoint(endpoint)
}

func (service *Service) snapshotKubernetesEndpoint(endpoint *portaineree.Endpoint) error {
	snapshot, err := service.kubernetesSnapshotter.CreateSnapshot(endpoint)
	if err != nil {
		return err
	}

	if snapshot != nil {
		endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{*snapshot}
	}

	return nil
}

func (service *Service) snapshotNomadEndpoint(endpoint *portaineree.Endpoint) (err error) {
	snapshot, err := service.nomadSnapshotter.CreateSnapshot(endpoint)
	if err != nil {
		return
	}

	endpoint.Nomad.Snapshots = []portaineree.NomadSnapshot{*snapshot}

	return
}

func (service *Service) snapshotDockerEndpoint(endpoint *portaineree.Endpoint) error {
	snapshot, err := service.dockerSnapshotter.CreateSnapshot(endpoint)
	if err != nil {
		return err
	}

	if snapshot != nil {
		endpoint.Snapshots = []portainer.DockerSnapshot{*snapshot}
	}

	return nil
}

func (service *Service) startSnapshotLoop() {
	ticker := time.NewTicker(time.Duration(service.snapshotIntervalInSeconds) * time.Second)

	err := service.snapshotEndpoints()
	if err != nil {
		log.Printf("[ERROR] [internal,snapshot] [message: background schedule error (environment snapshot).] [error: %s]", err)
	}

	for {
		select {
		case <-ticker.C:
			err := service.snapshotEndpoints()
			if err != nil {
				log.Printf("[ERROR] [internal,snapshot] [message: background schedule error (environment snapshot).] [error: %s]", err)
			}
		case <-service.shutdownCtx.Done():
			log.Println("[DEBUG] [internal,snapshot] [message: shutting down snapshotting]")
			ticker.Stop()
			return
		case interval := <-service.snapshotIntervalCh:
			ticker.Reset(interval)
		}
	}
}

func (service *Service) snapshotEndpoints() error {
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		if !SupportDirectSnapshot(&endpoint) {
			continue
		}

		if endpoint.URL == "" {
			continue
		}

		snapshotError := service.SnapshotEndpoint(&endpoint)

		latestEndpointReference, err := service.dataStore.Endpoint().Endpoint(endpoint.ID)
		if latestEndpointReference == nil {
			log.Printf("background schedule error (environment snapshot). Environment not found inside the database anymore (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
			continue
		}

		latestEndpointReference.Status = portaineree.EndpointStatusUp
		if snapshotError != nil {
			log.Printf("background schedule error (environment snapshot). Unable to create snapshot (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, snapshotError)
			latestEndpointReference.Status = portaineree.EndpointStatusDown
		}

		latestEndpointReference.Snapshots = endpoint.Snapshots
		latestEndpointReference.Kubernetes.Snapshots = endpoint.Kubernetes.Snapshots
		latestEndpointReference.Agent.Version = endpoint.Agent.Version

		err = service.dataStore.Endpoint().UpdateEndpoint(latestEndpointReference.ID, latestEndpointReference)
		if err != nil {
			log.Printf("background schedule error (environment snapshot). Unable to update environment (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
			continue
		}
	}

	return nil
}

// FetchDockerID fetches info.Swarm.Cluster.ID if environment(endpoint) is swarm and info.ID otherwise
func FetchDockerID(snapshot portainer.DockerSnapshot) (string, error) {
	info := snapshot.SnapshotRaw.Info

	if !snapshot.Swarm {
		return info.ID, nil
	}

	swarmInfo := info.Swarm
	if swarmInfo.Cluster == nil {
		return "", errors.New("swarm environment is missing cluster info snapshot")
	}

	clusterInfo := swarmInfo.Cluster
	return clusterInfo.ID, nil
}
