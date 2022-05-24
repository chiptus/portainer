package endpointedge

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
	"github.com/sirupsen/logrus"
)

type Credentials struct {
	ServerURL string
	Username  string
	Secret    string
}

type configResponse struct {
	StackFileContent    string
	Name                string
	RegistryCredentials []Credentials
}

// @summary Inspect an Edge Stack for an Environment(Endpoint)
// @description **Access policy**: public
// @tags edge, endpoints, edge_stacks
// @accept json
// @produce json
// @param id path string true "Environment(Endpoint) Id"
// @param stackId path string true "EdgeStack Id"
// @success 200 {object} configResponse
// @failure 500
// @failure 400
// @failure 404
// @router /endpoints/{id}/edge/stacks/{stackId} [get]
func (handler *Handler) endpointEdgeStackInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access environment", err}
	}

	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "stackId")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid edge stack identifier route variable", err}
	}

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an edge stack with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an edge stack with the specified identifier inside the database", err}
	}

	fileName := edgeStack.EntryPoint
	if endpointutils.IsDockerEndpoint(endpoint) {
		if fileName == "" {
			return &httperror.HandlerError{http.StatusBadRequest, "Docker is not supported by this stack", errors.New("Docker is not supported by this stack")}
		}
	}

	if endpointutils.IsKubernetesEndpoint(endpoint) {
		fileName = edgeStack.ManifestPath

		if fileName == "" {
			return &httperror.HandlerError{http.StatusBadRequest, "Kubernetes is not supported by this stack", errors.New("Kubernetes is not supported by this stack")}
		}
	}

	stackFileContent, err := handler.FileService.GetFileContent(edgeStack.ProjectPath, fileName)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Compose file from disk", err}
	}

	// Only provide registry credentials if we are sure that the agent connection is https
	// We will still allow the stack deployment to be attempted without credentials so that
	// failure can be seen rather than having the stack sit in deploying state forever
	registryCredentials := handler.getRegistryCredentialsForEdgeStack(edgeStack)
	if len(registryCredentials) > 0 && !secureEndpoint(endpoint) {
		logrus.Debugf("Insecure endpoint detected, private edge registries")
		registryCredentials = []Credentials{}
	}

	return response.JSON(w, configResponse{
		StackFileContent:    string(stackFileContent),
		Name:                edgeStack.Name,
		RegistryCredentials: registryCredentials,
	})
}

// secureEndpoint returns true if the endpoint is secure, false otherwise
// security is determined by the scheme being https.  We use the edge key because
// it's gauranteed not to have been altered
func secureEndpoint(endpoint *portaineree.Endpoint) bool {
	portainerUrl, error := getPortainerServerUrlFromEdgeKey(endpoint.EdgeKey)
	if error != nil {
		return false
	}

	u, err := url.Parse(portainerUrl)
	if err != nil {
		return false
	}

	return u.Scheme == "https"
}

// getPortainerServerUrlFromEdgeKey decodes a base64 encoded key and extract the portainer server URL
// edge key format: <portainer_instance_url>|<tunnel_server_addr>|<tunnel_server_fingerprint>|<endpoint_id>
func getPortainerServerUrlFromEdgeKey(key string) (string, error) {
	decodedKey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	keyInfo := strings.Split(string(decodedKey), "|")

	if len(keyInfo) != 4 {
		return "", errors.New("invalid key format")
	}

	return keyInfo[0], nil
}
