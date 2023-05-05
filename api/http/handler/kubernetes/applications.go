package kubernetes

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"

	"github.com/rs/zerolog/log"
)

// @id getKubernetesApplications
// @summary gets a list of Kubernetes applications
// @description Gets a list of Kubernetes deployments, statefulsets and daemonsets
// @description **Access policy**: authenticated
// @tags kubernetes
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param namespace path string true "specify an optional namespace"
// @success 200 {array} models.K8sApplication "Success"
// @failure 400 "Invalid request"
// @failure 401 "Unauthorized"
// @failure 403 "Permission denied"
// @failure 404 "Environment(Endpoint) or ServiceAccount not found"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace}/applications [get]
func (handler *Handler) getKubernetesApplications(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	cli, ok := handler.KubernetesClientFactory.GetProxyKubeClient(
		strconv.Itoa(endpointID), r.Header.Get("Authorization"),
	)
	if !ok {
		return httperror.InternalServerError("Failed to lookup KubeClient", nil)
	}

	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest("Invalid namespace identifier route variable", err)
	}

	applications, err := cli.GetApplications(namespace, "")
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve application", err)
	}

	return response.JSON(w, applications)
}

// @id getKubernetesApplication
// @summary gets a Kubernetes application
// @description Gets a Kubernetes deployment, statefulset and daemonset application details
// @description **Access policy**: authenticated
// @tags kubernetes
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param namespace path string true "The namespace"
// @param kind path string true "deployment, statefulset or daemonset"
// @param name path string true "name of the application"
// @param rollout-restart query string true "specify true to perform a rolling restart of the application"
// @success 200 {object} models.K8sApplication "Success"
// @failure 400 "Invalid request"
// @failure 401 "Unauthorized"
// @failure 403 "Permission denied"
// @failure 404 "Environment(Endpoint) or ServiceAccount not found"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace}/applications/{kind}/{name} [get]
func (handler *Handler) getKubernetesApplication(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	cli, ok := handler.KubernetesClientFactory.GetProxyKubeClient(
		strconv.Itoa(endpointID), r.Header.Get("Authorization"),
	)
	if !ok {
		return httperror.InternalServerError("Failed to get KubeClient", nil)
	}

	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest("Invalid namespace identifier route variable", err)
	}

	kind, err := request.RetrieveRouteVariableValue(r, "kind")
	if err != nil {
		return httperror.BadRequest("Invalid kind identifier route variable", err)
	}

	name, err := request.RetrieveRouteVariableValue(r, "name")
	if err != nil {
		return httperror.BadRequest("Invalid application identifier route variable", err)
	}

	// Ensure the application exists
	_, err = cli.GetApplication(namespace, kind, name)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve application", err)
	}

	restart, err := request.RetrieveBooleanQueryParameter(r, "rollout-restart", true)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter", err)
	}

	if restart {
		tokenData, err := security.RetrieveTokenData(r)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve user authentication token", err)
		}

		output, httperr := handler.restartKubernetesApplication(tokenData.ID, portaineree.EndpointID(endpointID), namespace, kind, name)
		if httperr != nil {
			return httperr
		}
		log.Debug().Str("output", output).Msg("Restarted application")
	}

	return response.Empty(w)
}

func (handler *Handler) restartKubernetesApplication(userID portaineree.UserID, endpointID portaineree.EndpointID, namespace, kind, name string) (string, *httperror.HandlerError) {
	resourceList := []string{kind + "/" + name}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err != nil {
		returnCode := http.StatusInternalServerError
		if handler.DataStore.IsErrObjectNotFound(err) {
			returnCode = http.StatusNotFound
		}

		return "", httperror.NewError(returnCode, "Unable to find the environment", err)
	}

	log.Debug().Msg("Restarting " + resourceList[0])
	output, err := handler.KubernetesDeployer.Restart(userID, endpoint, resourceList, namespace)
	if err != nil {
		return output, httperror.NewError(http.StatusInternalServerError, "Unable to restart application", err)
	}

	return output, nil
}
