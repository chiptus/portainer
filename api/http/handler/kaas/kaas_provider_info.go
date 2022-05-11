package kaas

import (
	"errors"
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id kaasProviderInfo
// @summary Returns information about a Cloud provider.
// @description The information returned can be used to provision a KaaS cluster.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud [post]
func (handler *Handler) kaasProviderInfo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	provider, err := request.RetrieveRouteVariableValue(r, "provider")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid user identifier route variable", err}
	}

	credentialId, _ := request.RetrieveNumericQueryParameter(r, "credentialId", true)
	if credentialId == 0 {
		return &httperror.HandlerError{http.StatusInternalServerError, "Missing credential id in the query parameter", err}
	}

	credential, err := handler.DataStore.CloudCredential().GetByID(models.CloudCredentialID(credentialId))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve %s information", provider), err}
	}

	switch provider {
	case portaineree.CloudProviderCivo:
		civoInfo, err := handler.cloudClusterInfoService.CivoGetInfo(credential)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Civo information", err}
		}

		return response.JSON(w, civoInfo)

	case portaineree.CloudProviderDigitalOcean:

		digitalOceanInfo, err := handler.cloudClusterInfoService.DigitalOceanGetInfo(credential)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve DigitalOcean information", err}
		}

		return response.JSON(w, digitalOceanInfo)

	case portaineree.CloudProviderLinode:

		linodeInfo, err := handler.cloudClusterInfoService.LinodeGetInfo(credential)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Linode information", err}
		}

		return response.JSON(w, linodeInfo)
	}

	return &httperror.HandlerError{http.StatusInternalServerError, "Unable to provision Kaas cluster", errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode")}
}
