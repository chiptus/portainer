package websocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	sshutil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/kubernetes/cli"
	"github.com/rs/zerolog/log"

	"golang.org/x/crypto/ssh"
)

// @summary Connect to a remote SSH Shell via a websocket
// @description When called, an SSH session to a microk8s cluster node will be created
// @description an ssh session will be created and hijacked.
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags websocket
// @accept json
// @produce json
// @param endpointId query int true "environment(endpoint) ID of the environment(endpoint) where the resource is located"
// @param nodeIp query string true "node ip address"
// @param token query string true "JWT token used for authentication against this environment(endpoint)"
// @success 200
// @failure 400
// @failure 409
// @failure 500
// @router /websocket/microk8s-shell [get]
func (handler *Handler) websocketMicrok8sShell(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve security context", err)
	}

	authorized, err := hasMicrok8sShellAccess(user, portaineree.EndpointID(endpointID))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	if !authorized {
		return httperror.Forbidden("Permission denied to access ssh shell", err)
	}

	nodeIP, err := request.RetrieveQueryParameter(r, "nodeIp", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: node", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment in the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment in the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	// Check this is one of our cluster nodes and not something random
	client, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError("Failed to get kubeClient", nil)
	}

	validNode, err := handler.isNodeInCluster(client, nodeIP)
	if err != nil {
		return httperror.InternalServerError("Unable to check if authorized to access node", err)
	}

	if !validNode {
		log.Error().Msgf("Requested node %s is not part of the cluster", nodeIP)
		return httperror.Forbidden("Forbidden", fmt.Errorf("requested node %s is not part of the cluster", nodeIP))
	}

	err = handler.handleSSHRequest(w, r, endpoint, nodeIP)
	if err != nil {
		return httperror.InternalServerError("An error occurred during handle SSH request", err)
	}

	return nil
}

func (handler *Handler) isNodeInCluster(client *cli.KubeClient, nodeIP string) (bool, error) {
	nodes, err := client.GetNodes()
	if err != nil {
		return false, fmt.Errorf("failed to get a list of valid nodes %w", err)
	}

	found := false
	for _, n := range nodes {
		if n.Address == nodeIP {
			found = true
			break
		}
	}

	return found, nil
}

func (handler *Handler) handleSSHRequest(w http.ResponseWriter, r *http.Request, endpoint *portaineree.Endpoint, nodeIP string) error {
	credential, err := handler.DataStore.CloudCredential().Read(endpoint.CloudProvider.CredentialID)
	if err != nil {
		return err
	}

	conn, err := sshutil.NewConnection(credential.Credentials["username"],
		credential.Credentials["password"],
		credential.Credentials["passphrase"],
		credential.Credentials["privateKey"],
		nodeIP,
	)
	if err != nil {
		return err
	}

	websocketConn, err := handler.connectionUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer websocketConn.Close()

	session, err := conn.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return hijackSSHSession(websocketConn, session)
}

func hijackSSHSession(websocketConn *websocket.Conn, session *ssh.Session) error {
	errorChan := make(chan error, 1)

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	go streamFromWebsocketToWriter(websocketConn, stdin, errorChan)
	go streamFromReaderToWebsocket(websocketConn, stdout, errorChan)
	go streamFromReaderToWebsocket(websocketConn, stderr, errorChan)

	modes := ssh.TerminalModes{ssh.ECHO: 1}
	if err := session.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		return fmt.Errorf("requestPty failed: %w", err)
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	log.Debug().Msgf("ssh session started")

	err = <-errorChan
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return err
	}

	log.Debug().Msgf("ssh session ended")
	return nil
}

// Check if the user is an admin or can have sheel access
func hasMicrok8sShellAccess(user *portaineree.User, endpointID portaineree.EndpointID) (bool, error) {
	isAdmin := user.Role == portaineree.AdministratorRole
	_, hasAccess := user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.OperationK8sClusterNodeW]
	return isAdmin || hasAccess, nil
}
