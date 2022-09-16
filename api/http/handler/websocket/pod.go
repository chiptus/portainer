package websocket

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// @summary Execute a websocket on pod
// @description The request will be upgraded to the websocket protocol.
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags websocket
// @accept json
// @produce json
// @param endpointId query int true "environment(endpoint) ID of the environment(endpoint) where the resource is located"
// @param namespace query string true "namespace where the container is located"
// @param podName query string true "name of the pod containing the container"
// @param containerName query string true "name of the container"
// @param command query string true "command to execute in the container"
// @param token query string true "JWT token used for authentication against this environment(endpoint)"
// @success 200
// @failure 400
// @failure 403
// @failure 404
// @failure 500
// @router /websocket/pod [get]
func (handler *Handler) websocketPodExec(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}

	namespace, err := request.RetrieveQueryParameter(r, "namespace", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: namespace", err)
	}

	podName, err := request.RetrieveQueryParameter(r, "podName", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: podName", err)
	}

	containerName, err := request.RetrieveQueryParameter(r, "containerName", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: containerName", err)
	}

	command, err := request.RetrieveQueryParameter(r, "command", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: command", err)
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find the environment associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment associated to the stack inside the database", err)
	}

	cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to create Kubernetes client", err)
	}

	permissionDeniedErr := "Permission denied to access environment"
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.Forbidden(permissionDeniedErr, err)
	}

	if tokenData.Role != portaineree.AdministratorRole {
		// check if the user has console RW access in the environment(endpoint)
		endpointRole, err := handler.authorizationService.GetUserEndpointRole(int(tokenData.ID), int(endpoint.ID))
		if err != nil {
			return httperror.Forbidden(permissionDeniedErr, err)
		} else if !endpointRole.Authorizations[portaineree.OperationK8sApplicationConsoleRW] {
			err = errors.New(permissionDeniedErr)
			return httperror.Forbidden(permissionDeniedErr, err)
		}
		// will skip if user can access all namespaces
		if !endpointRole.Authorizations[portaineree.OperationK8sAccessAllNamespaces] {
			// check if the user has RW access to the namespace
			namespaceAuthorizations, err := handler.authorizationService.GetNamespaceAuthorizations(int(tokenData.ID), *endpoint, cli)
			if err != nil {
				return httperror.Forbidden(permissionDeniedErr, err)
			} else if auth, ok := namespaceAuthorizations[namespace]; !ok || !auth[portaineree.OperationK8sAccessNamespaceWrite] {
				err = errors.New(permissionDeniedErr)
				return httperror.Forbidden(permissionDeniedErr, err)
			}
		}
	}

	serviceAccountToken, isAdminToken, err := handler.getToken(r, endpoint, false)
	if err != nil {
		return httperror.InternalServerError("Unable to get user service account token", err)
	}

	params := &webSocketRequestParams{
		endpoint: endpoint,
		token:    serviceAccountToken,
	}

	r.Header.Del("Origin")

	if endpoint.Type == portaineree.AgentOnKubernetesEnvironment {
		err := handler.proxyAgentWebsocketRequest(w, r, params)

		if err != nil {
			return httperror.InternalServerError("Unable to proxy websocket request to agent", err)
		}

		return nil
	} else if endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {

		err := handler.proxyEdgeAgentWebsocketRequest(w, r, params)
		if err != nil {
			return httperror.InternalServerError("Unable to proxy websocket request to Edge agent", err)
		}

		return nil
	}

	handlerErr := handler.hijackPodExecStartOperation(w, r, cli, serviceAccountToken, isAdminToken, endpoint, namespace, podName, containerName, command)
	if handlerErr != nil {
		return handlerErr
	}

	return nil
}

func (handler *Handler) hijackPodExecStartOperation(
	w http.ResponseWriter,
	r *http.Request,
	cli portaineree.KubeClient,
	serviceAccountToken string,
	isAdminToken bool,
	endpoint *portaineree.Endpoint,
	namespace, podName, containerName, command string,
) *httperror.HandlerError {
	commandArray := strings.Split(command, " ")

	websocketConn, err := handler.connectionUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return httperror.InternalServerError("Unable to upgrade the connection", err)
	}
	defer websocketConn.Close()

	stdinReader, stdinWriter := io.Pipe()
	defer stdinWriter.Close()
	stdoutReader, stdoutWriter := io.Pipe()
	defer stdoutWriter.Close()

	// errorChan is used to propagate errors from the go routines to the caller.
	errorChan := make(chan error, 1)
	go streamFromWebsocketToWriter(websocketConn, stdinWriter, errorChan)
	go streamFromReaderToWebsocket(websocketConn, stdoutReader, errorChan)

	// StartExecProcess is a blocking operation which streams IO to/from pod;
	// this must execute in asynchronously, since the websocketConn could return errors (e.g. client disconnects) before
	// the blocking operation is completed.
	go cli.StartExecProcess(serviceAccountToken, isAdminToken, namespace, podName, containerName, commandArray, stdinReader, stdoutWriter, errorChan)

	err = <-errorChan

	// websocket client successfully disconnected
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		log.Debug().Err(err).Msg("websocket error")

		return nil
	}

	return httperror.InternalServerError("Unable to start exec process inside container", err)
}

func (handler *Handler) getToken(request *http.Request, endpoint *portaineree.Endpoint, setLocalAdminToken bool) (string, bool, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return "", false, err
	}

	kubecli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return "", false, err
	}

	tokenCache := handler.kubernetesTokenCacheManager.GetOrCreateTokenCache(int(endpoint.ID))

	tokenManager, err := kubernetes.NewTokenManager(kubecli, handler.dataStore, tokenCache, setLocalAdminToken, handler.authorizationService)
	if err != nil {
		return "", false, err
	}

	if tokenData.Role == portaineree.AdministratorRole {
		return tokenManager.GetAdminServiceAccountToken(), true, nil
	}

	token, err := tokenManager.GetUserServiceAccountToken(int(tokenData.ID), int(endpoint.ID))
	if err != nil {
		return "", false, err
	}

	if token == "" {
		return "", false, fmt.Errorf("can not get a valid user service account token")
	}

	return token, false, nil
}
