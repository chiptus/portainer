package kaas

import (
	"errors"
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id providerInfo
// @summary Get information about the provisioning options for a cloud provider.
// @description The information returned can be used to provision a KaaS cluster.
// @description **Access policy**: administrator
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param provider path string true "The cloud provider to get information about."
// @param force query bool false "If true, get the up-to-date information (instead of cached information)."
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/{provider}/info [get]
func (handler *Handler) providerInfo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	provider, err := request.RetrieveRouteVariableValue(r, "provider")
	if err != nil {
		return httperror.BadRequest("Invalid user identifier route variable", err)
	}

	force, err := request.RetrieveBooleanQueryParameter(r, "force", true)
	if err != nil {
		return httperror.BadRequest("Failed parsing \"force\" boolean", err)
	}

	credential := &models.CloudCredential{}
	if provider != portaineree.CloudProviderMicrok8s {
		credentialId, _ := request.RetrieveNumericQueryParameter(r, "credentialId", true)
		if credentialId == 0 {
			return httperror.InternalServerError("Missing credential id in the query parameter", err)
		}

		credential, err = handler.dataStore.CloudCredential().Read(models.CloudCredentialID(credentialId))
		if err != nil {
			return httperror.InternalServerError(fmt.Sprintf("Unable to retrieve %s information", provider), err)
		}
	}

	switch provider {
	case portaineree.CloudProviderCivo:
		civoInfo, err := handler.cloudInfoService.CivoGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Civo information", err)
		}

		return response.JSON(w, civoInfo)

	case portaineree.CloudProviderDigitalOcean:

		digitalOceanInfo, err := handler.cloudInfoService.DigitalOceanGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve DigitalOcean information", err)
		}

		return response.JSON(w, digitalOceanInfo)

	case portaineree.CloudProviderLinode:

		linodeInfo, err := handler.cloudInfoService.LinodeGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Linode information", err)
		}

		return response.JSON(w, linodeInfo)

	case portaineree.CloudProviderGKE:
		gkeInfo, err := handler.cloudInfoService.GKEGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve GKE information", err)
		}

		return response.JSON(w, gkeInfo)

	case portaineree.CloudProviderAzure:

		azureInfo, err := handler.cloudInfoService.AzureGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Azure information", err)
		}

		return response.JSON(w, azureInfo)

	case portaineree.CloudProviderAmazon:
		awsInfo, err := handler.cloudInfoService.AmazonEksGetInfo(credential, force)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve AWS information", err)
		}

		return response.JSON(w, awsInfo)

	case portaineree.CloudProviderMicrok8s:
		microk8sInfo := handler.cloudInfoService.MicroK8sGetInfo()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Microk8s information", err)
		}

		return response.JSON(w, microk8sInfo)
	}

	return httperror.InternalServerError("Unable to get Kaas provider info", errors.New("invalid provider route parameter. Valid values: civo, digitalocean, linode, azure, gke"))
}
