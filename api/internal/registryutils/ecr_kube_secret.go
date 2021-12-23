package registryutils

import portaineree "github.com/portainer/portainer-ee/api"

func isRegistryAssignedToNamespace(registry portaineree.Registry, endpointID portaineree.EndpointID, namespace string) (in bool) {
	for _, ns := range registry.RegistryAccesses[endpointID].Namespaces {
		if ns == namespace {
			return true
		}
	}

	return
}

func RefreshEcrSecret(cli portaineree.KubeClient, endpoint *portaineree.Endpoint, dataStore portaineree.DataStore, namespace string) (err error) {
	registries, err := dataStore.Registry().Registries()
	if err != nil {
		return
	}

	for _, registry := range registries {
		if registry.Type != portaineree.EcrRegistry {
			continue
		}

		if !isRegistryAssignedToNamespace(registry, endpoint.ID, namespace) {
			continue
		}

		err = EnsureRegTokenValid(dataStore, &registry)
		if err != nil {
			return
		}

		err = cli.DeleteRegistrySecret(&registry, namespace)
		if err != nil {
			return
		}

		err = cli.CreateRegistrySecret(&registry, namespace)
		if err != nil {
			return
		}
	}

	return
}
