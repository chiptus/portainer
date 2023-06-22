package kubernetes

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/pkg/errors"
)

func (transport *baseTransport) proxyNamespaceDeleteOperation(request *http.Request, namespace string) (*http.Response, error) {
	if err := transport.tokenManager.kubecli.NamespaceAccessPoliciesDeleteNamespace(namespace); err != nil {
		return nil, errors.WithMessagef(err, "failed to delete a namespace [%s] from portainer config", namespace)
	}

	registries, err := transport.dataStore.Registry().ReadAll()
	if err != nil {
		return nil, err
	}

	if err := transport.tokenManager.kubecli.NamespaceAccessPoliciesDeleteNamespace(namespace); err != nil {
		return nil, errors.WithMessagef(err, "failed to delete a namespace [%s] from portainer config", namespace)
	}

	for _, registry := range registries {
		for endpointID, registryAccessPolicies := range registry.RegistryAccesses {
			if endpointID != transport.endpoint.ID {
				continue
			}

			namespaces := []string{}
			for _, ns := range registryAccessPolicies.Namespaces {
				if ns == namespace {
					continue
				}
				namespaces = append(namespaces, ns)
			}

			if len(namespaces) != len(registryAccessPolicies.Namespaces) {
				updatedAccessPolicies := portaineree.RegistryAccessPolicies{
					Namespaces:         namespaces,
					UserAccessPolicies: registryAccessPolicies.UserAccessPolicies,
					TeamAccessPolicies: registryAccessPolicies.TeamAccessPolicies,
				}

				registry.RegistryAccesses[endpointID] = updatedAccessPolicies
				err := transport.dataStore.Registry().Update(registry.ID, &registry)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	stacks, err := transport.dataStore.Stack().ReadAll()
	if err != nil {
		return nil, err
	}

	for _, s := range stacks {
		if s.Namespace == namespace && s.EndpointID == transport.endpoint.ID {
			if err := transport.dataStore.Stack().Delete(s.ID); err != nil {
				return nil, err
			}
		}
	}

	return transport.executeKubernetesRequest(request)
}
