package endpoints

import (
	"fmt"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer/api/dataservices/errors"
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
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [delete]
func (handler *Handler) endpointDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	if handler.demoService.IsDemoEnvironment(portaineree.EndpointID(endpointID)) {
		return &httperror.HandlerError{http.StatusForbidden, httperrors.ErrNotAvailableInDemo.Error(), httperrors.ErrNotAvailableInDemo}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == errors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	if endpoint.TLSConfig.TLS {
		folder := strconv.Itoa(endpointID)
		err = handler.FileService.DeleteTLSFiles(folder)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove TLS files from disk", err}
		}
	}

	err = handler.dataStore.Endpoint().DeleteEndpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove environment from the database", err}
	}

	handler.ProxyManager.DeleteEndpointProxy(endpoint.ID)

	if len(endpoint.UserAccessPolicies) > 0 || len(endpoint.TeamAccessPolicies) > 0 {
		err = handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update user authorizations", err}
		}
	}

	err = handler.dataStore.EndpointRelation().DeleteEndpointRelation(endpoint.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove environment relation from the database", err}
	}

	for _, tagID := range endpoint.TagIDs {
		tag, err := handler.dataStore.Tag().Tag(tagID)
		if err != nil {
			return &httperror.HandlerError{http.StatusNotFound, "Unable to find tag inside the database", err}
		}

		delete(tag.Endpoints, endpoint.ID)

		err = handler.dataStore.Tag().UpdateTag(tagID, tag)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist tag relation inside the database", err}
		}
	}

	edgeGroups, err := handler.dataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from the database", err}
	}

	for idx := range edgeGroups {
		edgeGroup := &edgeGroups[idx]
		endpointIdx := findEndpointIndex(edgeGroup.Endpoints, endpoint.ID)
		if endpointIdx != -1 {
			edgeGroup.Endpoints = removeElement(edgeGroup.Endpoints, endpointIdx)
			err = handler.dataStore.EdgeGroup().UpdateEdgeGroup(edgeGroup.ID, edgeGroup)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update edge group", err}
			}
		}
	}

	edgeStacks, err := handler.dataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge stacks from the database", err}
	}

	for idx := range edgeStacks {
		edgeStack := &edgeStacks[idx]
		if _, ok := edgeStack.Status[endpoint.ID]; ok {
			delete(edgeStack.Status, endpoint.ID)
			err = handler.dataStore.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update edge stack", err}
			}
		}
	}

	registries, err := handler.dataStore.Registry().Registries()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
	}

	for idx := range registries {
		registry := &registries[idx]
		if _, ok := registry.RegistryAccesses[endpoint.ID]; ok {
			delete(registry.RegistryAccesses, endpoint.ID)
			err = handler.dataStore.Registry().UpdateRegistry(registry.ID, registry)
			if err != nil {
				return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to update registry accesses", Err: err}
			}
		}
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	err = handler.deleteAccessPolicies(*endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to delete endpoint access policies", err}
	}

	return response.Empty(w)
}

func findEndpointIndex(tags []portaineree.EndpointID, searchEndpointID portaineree.EndpointID) int {
	for idx, tagID := range tags {
		if searchEndpointID == tagID {
			return idx
		}
	}
	return -1
}

func removeElement(arr []portaineree.EndpointID, index int) []portaineree.EndpointID {
	if index < 0 {
		return arr
	}
	lastTagIdx := len(arr) - 1
	arr[index] = arr[lastTagIdx]
	return arr[:lastTagIdx]
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

	kcl, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
	if err != nil {
		return fmt.Errorf("Unable to get k8s environment access @ %d: %w", int(endpoint.ID), err)
	}

	emptyPolicies := make(map[string]portaineree.K8sNamespaceAccessPolicy)
	err = kcl.UpdateNamespaceAccessPolicies(emptyPolicies)
	if err != nil {
		return fmt.Errorf("Unable to update environment namespace access @ %d: %w", int(endpoint.ID), err)
	}

	return nil
}
