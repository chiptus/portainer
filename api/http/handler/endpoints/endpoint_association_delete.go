package endpoints

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	bolterrors "github.com/portainer/portainer-ee/api/bolt/errors"
)

// @id EndpointAssociationDelete
// @summary De-association an edge environment(endpoint)
// @description De-association an edge environment(endpoint).
// @description **Access policy**: administrator
// @security ApiKeyAuth
// @security jwt
// @tags endpoints
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/association [put]
func (handler *Handler) endpointAssociationDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	if endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment && endpoint.Type != portaineree.EdgeAgentOnDockerEnvironment {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment type", errors.New("Invalid environment type")}
	}

	endpoint.EdgeID = ""
	endpoint.Snapshots = []portaineree.DockerSnapshot{}
	endpoint.Kubernetes.Snapshots = []portaineree.KubernetesSnapshot{}

	endpoint.EdgeKey, err = handler.updateEdgeKey(endpoint.EdgeKey)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Invalid EdgeKey", err}
	}

	err = handler.dataStore.Endpoint().UpdateEndpoint(portaineree.EndpointID(endpointID), endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed persisting environment in database", err}
	}

	handler.ReverseTunnelService.SetTunnelStatusToIdle(endpoint.ID)

	return response.JSON(w, endpoint)
}

func (handler *Handler) updateEdgeKey(edgeKey string) (string, error) {
	oldEdgeKeyByte, err := base64.RawStdEncoding.DecodeString(edgeKey)
	if err != nil {
		return "", err
	}

	oldEdgeKeyStr := string(oldEdgeKeyByte)

	httpPort := getPort(handler.BindAddress)
	httpsPort := getPort(handler.BindAddressHTTPS)

	// replace "http://" with "https://" and replace ":9000" with ":9443", in the case of default values
	// oldEdgeKeyStr example: http://10.116.1.178:9000|10.116.1.178:8000|46:99:4a:8d:a6:de:6a:bd:d8:e2:1c:99:81:60:54:55|52
	r := regexp.MustCompile(fmt.Sprintf("^(http://)([^|]+)(:%s)(|.*)", httpPort))
	newEdgeKeyStr := r.ReplaceAllString(oldEdgeKeyStr, fmt.Sprintf("https://$2:%s$4", httpsPort))

	return base64.RawStdEncoding.EncodeToString([]byte(newEdgeKeyStr)), nil
}

func getPort(url string) string {
	items := strings.Split(url, ":")
	return items[len(items)-1]
}
