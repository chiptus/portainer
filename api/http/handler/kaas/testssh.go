package kaas

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/providers"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/sshtest"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id testSSH
// @summary Test SSH connection to nodes
// @description Test SSH connection to nodes.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body providers.Microk8sTestSSHPayload true "Node IPs and credential ID"
// @success 200 {object} []sshtest.SSHTestResult "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/testssh [post]
func (handler *Handler) sshTestNodeIPs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload providers.Microk8sTestSSHPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err == nil {
		// validate if the user has access to the environment
		securityContext, err := security.RetrieveRestrictedRequestContext(r)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve info from request context", err)
		}

		user, err := handler.dataStore.User().Read(securityContext.UserID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve security context", err)
		}

		authorized := canWriteK8sClusterNode(user, portaineree.EndpointID(endpoint.ID))
		if !authorized {
			return httperror.Forbidden("Permission denied to remove nodes from the cluster", nil)
		}
	}

	credentials, err := handler.dataStore.CloudCredential().Read(payload.CredentialID)
	if err != nil {
		return httperror.InternalServerError("unable to read credentials from the database", err)
	}

	// get ip ranges and run ssh tests
	config, err := sshUtil.NewSSHConfig(
		credentials.Credentials["username"],
		credentials.Credentials["password"],
		credentials.Credentials["passphrase"],
		credentials.Credentials["privateKey"],
	)
	if err != nil {
		return httperror.InternalServerError("unable to create ssh config with given credentials", err)
	}
	results := sshtest.SSHTest(&config.Config, payload.Nodes)
	return response.JSON(w, results)
}
