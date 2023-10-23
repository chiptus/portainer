package pendingactions

import (
	"context"
	"fmt"
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	kubecli "github.com/portainer/portainer-ee/api/kubernetes/cli"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

const (
	CleanNAPWithOverridePolicies      = "CleanNAPWithOverridePolicies"
	UpsertPortainerK8sClusterRoles    = "UpsertPortainerK8sClusterRoles"
	DeletePortainerK8sRegistrySecrets = "DeletePortainerK8sRegistrySecrets"
)

type (
	PendingActionsService struct {
		authorizationService *authorization.Service
		clientFactory        *kubecli.ClientFactory
		dataStore            dataservices.DataStore
		shutdownCtx          context.Context

		mu sync.Mutex
	}
)

func NewService(
	dataStore dataservices.DataStore,
	clientFactory *kubecli.ClientFactory,
	authorizationService *authorization.Service,
	shutdownCtx context.Context,
) *PendingActionsService {
	return &PendingActionsService{
		dataStore:            dataStore,
		shutdownCtx:          shutdownCtx,
		authorizationService: authorizationService,
		clientFactory:        clientFactory,
		mu:                   sync.Mutex{},
	}
}

func (service *PendingActionsService) Create(r portainer.PendingActions) error {
	return service.dataStore.PendingActions().Create(&r)
}

func (service *PendingActionsService) Execute(id portainer.EndpointID) error {

	service.mu.Lock()
	defer service.mu.Unlock()

	endpoint, err := service.dataStore.Endpoint().Endpoint(id)
	if err != nil {
		return fmt.Errorf("failed to retrieve environment %d: %w", id, err)
	}

	if endpoint.Status != portainer.EndpointStatusUp {
		log.Debug().Msgf("Environment %q (id: %d) is not up", endpoint.Name, id)
		return fmt.Errorf("environment %q (id: %d) is not up", endpoint.Name, id)
	}

	pendingActions, err := service.dataStore.PendingActions().ReadAll()
	if err != nil {
		log.Error().Err(err).Msgf("failed to retrieve pending actions")
		return fmt.Errorf("failed to retrieve pending actions for environment %d: %w", id, err)
	}

	for _, endpointPendingAction := range pendingActions {
		if endpointPendingAction.EndpointID == id {
			err := service.executePendingAction(endpointPendingAction, endpoint)
			if err != nil {
				log.Warn().Err(err).Msgf("failed to execute pending action")
				return fmt.Errorf("failed to execute pending action: %w", err)
			}

			err = service.dataStore.PendingActions().Delete(endpointPendingAction.ID)
			if err != nil {
				log.Warn().Err(err).Msgf("failed to delete pending action")
				return fmt.Errorf("failed to delete pending action: %w", err)
			}
		}
	}

	return nil
}

func (service *PendingActionsService) executePendingAction(pendingAction portainer.PendingActions, endpoint *portaineree.Endpoint) error {
	log.Debug().Msgf("Executing pending action %s for environment %d", pendingAction.Action, pendingAction.EndpointID)

	defer func() {
		log.Debug().Msgf("End executing pending action %s for environment %d", pendingAction.Action, pendingAction.EndpointID)
	}()

	switch pendingAction.Action {
	case CleanNAPWithOverridePolicies:
		if (pendingAction.ActionData == nil) || (pendingAction.ActionData.(portainer.EndpointGroupID) == 0) {
			service.authorizationService.CleanNAPWithOverridePolicies(service.dataStore, endpoint, nil)
			return nil
		}

		endpointGroupID := pendingAction.ActionData.(portainer.EndpointGroupID)
		endpointGroup, err := service.dataStore.EndpointGroup().Read(portainer.EndpointGroupID(endpointGroupID))
		if err != nil {
			log.Error().Err(err).Msgf("Error reading environment group to clean NAP with override policies for environment %d and environment group %d", endpoint.ID, endpointGroup.ID)
			return fmt.Errorf("failed to retrieve environment group %d: %w", endpointGroupID, err)
		}
		err = service.authorizationService.CleanNAPWithOverridePolicies(service.dataStore, endpoint, endpointGroup)
		if err != nil {
			log.Error().Err(err).Msgf("Error cleaning NAP with override policies for environment %d and environment group %d", endpoint.ID, endpointGroup.ID)
			return fmt.Errorf("failed to clean NAP with override policies for environment %d and environment group %d: %w", endpoint.ID, endpointGroup.ID, err)
		}

		return nil

	case UpsertPortainerK8sClusterRoles:
		kubeClient, err := service.clientFactory.GetKubeClient(endpoint)
		if err != nil {
			return fmt.Errorf("failed to get kube client for environment %d: %w", endpoint.ID, err)
		}

		err = kubeClient.UpsertPortainerK8sClusterRoles(endpoint.Kubernetes.Configuration)
		if err != nil {
			log.Warn().Err(err).Int("endpoint_id", int(endpoint.ID)).Msgf("Unable to update kubernetes cluster roles")
			return fmt.Errorf("failed to upsert portainer kubernetes cluster roles for environment %d: %w", endpoint.ID, err)
		}

		return nil

	case DeletePortainerK8sRegistrySecrets:
		if pendingAction.ActionData == nil {
			return nil
		}

		// This shouldn't ever fail because we have full control over the ActionData, but just in case, lets log an error.
		// If this error message is ever seen, it indicates a bug in the code.
		registryData, err := convertToDeletePortainerK8sRegistrySecretsData(pendingAction.ActionData)
		if err != nil {
			return fmt.Errorf("failed to parse pendingActionData: %w", err)
		}

		err = service.DeleteKubernetesRegistrySecrets(endpoint, registryData)
		if err != nil {
			log.Warn().Err(err).Int("endpoint_id", int(endpoint.ID)).Msgf("Unable to delete kubernetes registry secrets")
			return fmt.Errorf("failed to delete kubernetes registry secrets for environment %d: %w", endpoint.ID, err)
		}

		return nil
	}

	return nil
}
