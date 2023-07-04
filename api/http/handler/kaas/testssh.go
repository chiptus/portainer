package kaas

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	sshUtil "github.com/portainer/portainer-ee/api/cloud/util/ssh"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/providers"
	"github.com/portainer/portainer-ee/api/internal/sshtest"
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
