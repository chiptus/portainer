package kubernetes

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

func (h *Handler) getRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest(
			"Invalid environment identifier route variable",
			err,
		)
	}

	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest(
			"Invalid namespace identifier route variable",
			err,
		)
	}

	endpoint, err := h.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if h.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	} else if err != nil {
		return httperror.InternalServerError(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	}

	cli, err := h.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError(
			"Unable to create Kubernetes client",
			err,
		)
	}

	clusterRoleBindings, err := cli.GetRoleBindings(namespace)
	if err != nil {
		return httperror.InternalServerError(
			"Failed to fetch role bindings",
			err,
		)
	}

	return response.JSON(w, clusterRoleBindings)
}

func (h *Handler) deleteRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest(
			"Invalid environment identifier route variable",
			err,
		)
	}

	var payload models.K8sRoleBindingDeleteRequests
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest(
			"Invalid request payload",
			err,
		)
	}

	endpoint, err := h.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if h.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	} else if err != nil {
		return httperror.InternalServerError(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	}

	cli, err := h.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError(
			"Unable to create Kubernetes client",
			err,
		)
	}

	err = cli.DeleteRoleBindings(payload)
	if err != nil {
		return httperror.InternalServerError(
			"Failed to delete role bindings",
			err,
		)
	}

	return nil
}
