package endpoints

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

// @id EndpointDelete
// @summary Remove an environment(endpoint)
// @description Remove an environment(endpoint).
// @description **Access policy**: administrator
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Environment(Endpoint) identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [delete]
func (handler *Handler) endpointDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	// This is a Portainer provisioned cloud environment
	deleteCluster, err := request.RetrieveBooleanQueryParameter(r, "deleteCluster", true)
	if err != nil {
		return httperror.BadRequest("Invalid boolean query parameter", err)
	}

	if handler.demoService.IsDemoEnvironment(portaineree.EndpointID(endpointID)) {
		return httperror.Forbidden(httperrors.ErrNotAvailableInDemo.Error(), httperrors.ErrNotAvailableInDemo)
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		return handler.deleteEndpoint(tx, portaineree.EndpointID(endpointID), deleteCluster)
	})
	if err != nil {
		var handlerError *httperror.HandlerError
		if errors.As(err, &handlerError) {
			return handlerError
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func (handler *Handler) deleteEndpoint(tx dataservices.DataStoreTx, endpointID portaineree.EndpointID, deleteCluster bool) error {
	endpoint, err := tx.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to read the environment record from the database", err)
	}

	if endpoint.TLSConfig.TLS {
		folder := strconv.Itoa(int(endpointID))
		err = handler.FileService.DeleteTLSFiles(folder)
		if err != nil {
			log.Error().Err(err).Msgf("Unable to remove TLS files from disk when deleting endpoint %d", endpointID)
		}
	}

	err = tx.Snapshot().Delete(endpointID)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to remove the snapshot from the database")
	}

	err = handler.deleteAccessPolicies(*endpoint)
	if err != nil && !handler.DataStore.IsErrObjectNotFound(err) {
		// log as an error because we still want to continue deletion steps
		log.Error().Err(err).Msg("Unable to delete endpoint access policies - continuing environment deletion")
		log.Warn().Msg("If the environment removed from Portainer still exists, Portainer access policies will remain")
	}

	// if edge endpoint, remove from edge update schedules
	if endpointutils.IsEdgeEndpoint(endpoint) {
		edgeUpdates, err := tx.EdgeUpdateSchedule().ReadAll()
		if err != nil {
			// skip
			log.Warn().Err(err).Msg("Unable to retrieve edge update schedules from the database")
		} else {
			for i := range edgeUpdates {
				edgeUpdate := edgeUpdates[i]
				if edgeUpdate.EnvironmentsPreviousVersions[endpoint.ID] != "" {
					delete(edgeUpdate.EnvironmentsPreviousVersions, endpoint.ID)
					err = tx.EdgeUpdateSchedule().Update(edgeUpdate.ID, &edgeUpdate)
					if err != nil {
						// skip
						log.Warn().Err(err).Msg("Unable to update edge update schedule")
					}
				}
			}
		}
	}

	if endpoint.CloudProvider != nil {
		if deleteCluster {
			log.Info().Msgf("Deleting the remote cluster associated to environment %d (%s)", endpoint.ID, endpoint.Name)
			handler.cloudManagementService.DeleteCluster(tx, endpoint)
		}
	}

	handler.ProxyManager.DeleteEndpointProxy(endpoint.ID)

	if len(endpoint.UserAccessPolicies) > 0 || len(endpoint.TeamAccessPolicies) > 0 {
		err = handler.AuthorizationService.UpdateUsersAuthorizationsTx(tx)
		if err != nil {
			log.Warn().Err(err).Msgf("Unable to update user authorizations")
		}
	}

	err = tx.EndpointRelation().DeleteEndpointRelation(endpoint.ID)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to remove environment relation from the database")
	}

	for _, tagID := range endpoint.TagIDs {
		var tag *portaineree.Tag
		tag, err = tx.Tag().Read(tagID)
		if err == nil {
			delete(tag.Endpoints, endpoint.ID)
			err = tx.Tag().Update(tagID, tag)
		}

		if handler.DataStore.IsErrObjectNotFound(err) {
			log.Warn().Err(err).Msgf("Unable to find tag inside the database")
		} else if err != nil {
			log.Warn().Err(err).Msgf("Unable to delete tag relation from the database")
		}
	}

	edgeGroups, err := tx.EdgeGroup().ReadAll()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to retrieve edge groups from the database")
	}

	for _, edgeGroup := range edgeGroups {
		edgeGroup.Endpoints = slices.DeleteFunc(edgeGroup.Endpoints, func(e portaineree.EndpointID) bool {
			return e == endpoint.ID
		})

		err = tx.EdgeGroup().Update(edgeGroup.ID, &edgeGroup)
		if err != nil {
			log.Warn().Err(err).Msgf("Unable to update edge group")
		}
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to retrieve edge stacks from the database")
	}

	for idx := range edgeStacks {
		edgeStack := &edgeStacks[idx]
		if _, ok := edgeStack.Status[endpoint.ID]; ok {
			delete(edgeStack.Status, endpoint.ID)
			err = tx.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack, true)
			if err != nil {
				log.Warn().Err(err).Msgf("Unable to update edge stack")
			}
		}
	}

	registries, err := tx.Registry().ReadAll()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to retrieve registries from the database")
	}

	for idx := range registries {
		registry := &registries[idx]
		if _, ok := registry.RegistryAccesses[endpoint.ID]; ok {
			delete(registry.RegistryAccesses, endpoint.ID)
			err = tx.Registry().Update(registry.ID, registry)
			if err != nil {
				log.Warn().Err(err).Msgf("Unable to update registry accesses")
			}
		}
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	if endpointutils.IsEdgeEndpoint(endpoint) {
		edgeJobs, err := handler.DataStore.EdgeJob().ReadAll()
		if err != nil {
			log.Warn().Err(err).Msgf("Unable to retrieve edge jobs from the database")
		}

		for idx := range edgeJobs {
			edgeJob := &edgeJobs[idx]
			if _, ok := edgeJob.Endpoints[endpoint.ID]; ok {
				delete(edgeJob.Endpoints, endpoint.ID)

				err = tx.EdgeJob().Update(edgeJob.ID, edgeJob)
				if err != nil {
					log.Warn().Err(err).Msgf("Unable to update edge job")
				}
			}
		}
	}

	if err = handler.deleteEdgeConfigs(tx, endpoint); err != nil {
		return httperror.InternalServerError("Unable to update edge configurations", err)
	}

	err = tx.Endpoint().DeleteEndpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		return httperror.InternalServerError("Unable to delete the environment from the database", err)
	}

	return nil
}

func (handler *Handler) deleteAccessPolicies(endpoint portaineree.Endpoint) error {
	if endpoint.Type != portaineree.KubernetesLocalEnvironment &&
		endpoint.Type != portaineree.AgentOnKubernetesEnvironment &&
		endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment {
		return nil
	}

	if endpoint.URL == "" {
		return nil
	}

	// run as a non blocking function for deleting edge environment with long check in intervals
	go func() {
		log.Info().Msg("Starting to update access policies")
		kcl, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("Unable to get k8s environment access while deleting environment @ %d", int(endpoint.ID))
			return
		}

		emptyPolicies := make(map[string]portaineree.K8sNamespaceAccessPolicy)
		err = kcl.UpdateNamespaceAccessPolicies(emptyPolicies)
		if err != nil {
			log.Error().Err(err).Msgf("Unable to update environment namespace access while deleting environment @ %d", int(endpoint.ID))
		}
	}()

	return nil
}

func (handler *Handler) deleteEdgeConfigs(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint) error {
	configState, err := tx.EdgeConfigState().Read(endpoint.ID)
	if dataservices.IsErrObjectNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	for configID, state := range configState.States {
		config, err := tx.EdgeConfig().Read(configID)
		if err != nil {
			return err
		}

		if state == portaineree.EdgeConfigIdleState {
			config.Progress.Success--
		}

		config.Progress.Total--

		if config.Progress.Success == config.Progress.Total {
			config.State = portaineree.EdgeConfigIdleState
		}

		tx.EdgeConfig().Update(config.ID, config)
	}

	return tx.EdgeConfigState().Delete(endpoint.ID)
}
