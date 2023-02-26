package endpointutils

import (
	"fmt"
	"strings"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	log "github.com/rs/zerolog/log"
)

// IsLocalEndpoint returns true if this is a local environment(endpoint)
func IsLocalEndpoint(endpoint *portaineree.Endpoint) bool {
	return strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") || endpoint.Type == 5
}

// IsKubernetesEndpoint returns true if this is a kubernetes environment(endpoint)
func IsKubernetesEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.KubernetesLocalEnvironment ||
		endpoint.Type == portaineree.AgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment
}

// IsDockerEndpoint returns true if this is a docker environment(endpoint)
func IsDockerEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.DockerEnvironment ||
		endpoint.Type == portaineree.AgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment
}

// IsNomadEndpoint returns true if this is a nomad environment(endpoint)
func IsNomadEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.EdgeAgentOnNomadEnvironment
}

// IsEdgeEndpoint returns true if endpoint is an Edge Endpoint
func IsEdgeEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnNomadEnvironment
}

// IsAgentEndpoint returns true if this is an Agent endpoint
func IsAgentEndpoint(endpoint *portaineree.Endpoint) bool {
	return endpoint.Type == portaineree.AgentOnDockerEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment ||
		endpoint.Type == portaineree.AgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment ||
		endpoint.Type == portaineree.EdgeAgentOnNomadEnvironment
}

// FilterByExcludeIDs receives an environment(endpoint) array and returns a filtered array using an excludeIds param
func FilterByExcludeIDs(endpoints []portaineree.Endpoint, excludeIds []portaineree.EndpointID) []portaineree.Endpoint {
	if len(excludeIds) == 0 {
		return endpoints
	}

	filteredEndpoints := make([]portaineree.Endpoint, 0)

	idsSet := make(map[portaineree.EndpointID]bool)
	for _, id := range excludeIds {
		idsSet[id] = true
	}

	for _, endpoint := range endpoints {
		if !idsSet[endpoint.ID] {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

// EndpointSet receives an environment(endpoint) array and returns a set
func EndpointSet(endpointIDs []portaineree.EndpointID) map[portaineree.EndpointID]bool {
	set := map[portaineree.EndpointID]bool{}

	for _, endpointID := range endpointIDs {
		set[endpointID] = true
	}

	return set
}

func InitialIngressClassDetection(endpoint *portaineree.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	log.Info().Msg("attempting to detect ingress classes in the cluster")
	cli, err := factory.GetKubeClient(endpoint)
	if err != nil {
		log.Info().Err(err).Msg("unable to create kubernetes client for ingress class detection")
		return
	}
	controllers, err := cli.GetIngressControllers()
	if err != nil {
		log.Info().Err(err).Msg("failed to fetch ingressclasses")
		return
	}

	var updatedClasses []portaineree.KubernetesIngressClassConfig
	for i := range controllers {
		var updatedClass portaineree.KubernetesIngressClassConfig
		updatedClass.Name = controllers[i].ClassName
		updatedClass.Type = controllers[i].Type
		updatedClasses = append(updatedClasses, updatedClass)
	}

	endpoint.Kubernetes.Configuration.IngressClasses = updatedClasses
	err = endpointService.UpdateEndpoint(
		portaineree.EndpointID(endpoint.ID),
		endpoint,
	)
	if err != nil {
		log.Info().Err(err).Msg("unable to store found IngressClasses inside the database")
		return
	}
}

func InitialMetricsDetection(endpoint *portaineree.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	log.Info().Msg("attempting to detect metrics api in the cluster")
	cli, err := factory.GetKubeClient(endpoint)
	if err != nil {
		log.Info().Err(err).Msg("unable to create kubernetes client for initial metrics detection")
		return
	}
	_, err = cli.GetMetrics()
	if err != nil {
		log.Info().Err(err).Msg("unable to fetch metrics: leaving metrics collection disabled.")
		return
	}
	endpoint.Kubernetes.Configuration.UseServerMetrics = true
	endpoint.Kubernetes.Flags.IsServerMetricsDetected = true
	err = endpointService.UpdateEndpoint(
		portaineree.EndpointID(endpoint.ID),
		endpoint,
	)
	if err != nil {
		log.Info().Err(err).Msg("unable to enable UseServerMetrics inside the database")
		return
	}
}

func storageDetect(endpoint *portaineree.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) error {
	cli, err := factory.GetKubeClient(endpoint)
	if err != nil {
		log.Info().Err(err).Msg("unable to create Kubernetes client for initial storage detection")
		return err
	}

	storage, err := cli.GetStorage()
	if err != nil {
		log.Info().Err(err).Msg("unable to fetch storage classes: leaving storage classes disabled")
		return err
	}
	if len(storage) == 0 {
		log.Info().Err(err).Msg("zero storage classes found: they may be still building, retrying in 30 seconds")
		return fmt.Errorf("zero storage classes found: they may be still building, retrying in 30 seconds")
	}
	endpoint.Kubernetes.Configuration.StorageClasses = storage
	endpoint.Kubernetes.Flags.IsServerStorageDetected = true
	err = endpointService.UpdateEndpoint(
		portaineree.EndpointID(endpoint.ID),
		endpoint,
	)
	if err != nil {
		log.Info().Err(err).Msg("unable to enable storage class inside the database")
		return err
	}

	return nil
}

func InitialStorageDetection(endpoint *portaineree.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	log.Info().Msg("attempting to detect storage classes in the cluster")
	err := storageDetect(endpoint, endpointService, factory)
	if err == nil {
		return
	}
	log.Err(err).Msg("error while detecting storage classes")
	go func() {
		// Retry after 30 seconds if the initial detection failed.
		log.Info().Msg("retrying storage detection in 30 seconds")
		time.Sleep(30 * time.Second)
		err := storageDetect(endpoint, endpointService, factory)
		log.Err(err).Msg("final error while detecting storage classes")
	}()
}
