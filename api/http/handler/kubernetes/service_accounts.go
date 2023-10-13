package kubernetes

import (
	"net/http"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GetServiceAccounts
// @summary Get a list of service accounts
// @description Get a list of service accounts for the given kubernetes environment in the given namespace
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param namespace path string true "Kubernetes namespace"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace}/service_accounts [get]
func (h *Handler) getKubernetesServiceAccounts(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest(
			"Invalid namespace identifier route variable",
			err,
		)
	}

	cli, handlerErr := h.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	services, err := cli.GetServiceAccounts(namespace)
	if err != nil {
		return httperror.InternalServerError(
			"Unable to retrieve services",
			err,
		)
	}

	return response.JSON(w, services)
}

// @id DeleteServiceAccounts
// @summary Delete the provided service accounts
// @description Delete the provided roles for the given Kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body models.K8sServiceAccountDeleteRequests true "Service accounts to delete "
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/service_accounts/delete [POST]
func (handler *Handler) deleteKubernetesServiceAccounts(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.K8sServiceAccountDeleteRequests
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cli, handlerErr := handler.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	err = cli.DeleteServiceAccounts(payload)
	if err != nil {
		return httperror.InternalServerError("Unable to delete service accounts", err)
	}

	return nil
}
