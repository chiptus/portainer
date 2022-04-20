package kaas

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
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

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve settings from the database", err}
	}

	switch provider {
	case portaineree.CloudProviderCivo:
		if settings.CloudApiKeys.CivoApiKey == "" {
			return &httperror.HandlerError{http.StatusServiceUnavailable, "Missing Civo API key in cloud settings", errors.New("Missing Civo API key in cloud settings")}
		}

		civoInfo, err := handler.cloudClusterInfoService.CivoGetInfo(settings.CloudApiKeys.CivoApiKey)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Civo information", err}
		}

		return response.JSON(w, civoInfo)

	case portaineree.CloudProviderDigitalOcean:
		if settings.CloudApiKeys.DigitalOceanToken == "" {
			return &httperror.HandlerError{http.StatusServiceUnavailable, "Missing DigitalOcean API token in cloud settings", errors.New("Missing DigitalOcean API token in cloud settings")}
		}

		digitalOceanInfo, err := handler.cloudClusterInfoService.DigitalOceanGetInfo(settings.CloudApiKeys.DigitalOceanToken)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve DigitalOcean information", err}
		}

		return response.JSON(w, digitalOceanInfo)

	case portaineree.CloudProviderLinode:
		if settings.CloudApiKeys.LinodeToken == "" {
			return &httperror.HandlerError{http.StatusServiceUnavailable, "Missing Linode API token in cloud settings", errors.New("Missing Linode API token in cloud settings")}
		}

		linodeInfo, err := handler.cloudClusterInfoService.LinodeGetInfo(settings.CloudApiKeys.LinodeToken)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Linode information", err}
		}

		return response.JSON(w, linodeInfo)
	}

	return &httperror.HandlerError{http.StatusInternalServerError, "Unable to provision Kaas cluster", errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode")}
}
